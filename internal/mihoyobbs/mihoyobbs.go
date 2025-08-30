package mihoyobbs

import (
	"auto-bgi/autoLog"
	"auto-bgi/internal/http"
	"auto-bgi/internal/mysConfig"
	"auto-bgi/internal/utils"
	"encoding/json"
	"fmt"
	"strings"
)

// Mihoyobbs 米游社签到
type Mihoyobbs struct {
	client        *http.Client
	todayGetCoins int
	haveCoins     int
	bbsList       []BBSInfo
	postsList     []PostInfo
	taskDo        TaskStatus
}

// BBSInfo 社区信息
type BBSInfo struct {
	ID   string `json:"id"` // 应该是字符串，与Python版本保持一致
	Name string `json:"name"`
}

// PostInfo 帖子信息
type PostInfo struct {
	PostID    string `json:"post_id"`
	Subject   string `json:"subject"`
	ForumID   int    `json:"forum_id"`
	IsGood    int    `json:"is_good"`
	IsTop     int    `json:"is_top"`
	IsEssence int    `json:"is_essence"`
}

// TaskStatus 任务状态
type TaskStatus struct {
	Sign    bool `json:"sign"`
	Read    bool `json:"read"`
	ReadNum int  `json:"read_num"`
	Like    bool `json:"like"`
	LikeNum int  `json:"like_num"`
	Share   bool `json:"share"`
}

// API响应结构
type APIResponse struct {
	Retcode int         `json:"retcode"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// TaskListResponse 任务列表响应
type TaskListResponse struct {
	Retcode int    `json:"retcode"`
	Message string `json:"message"`
	Data    struct {
		States []struct {
			MissionID     int    `json:"mission_id"`
			Process       int    `json:"process"`
			HappenedTimes int    `json:"happened_times"`
			IsGetAward    bool   `json:"is_get_award"`
			MissionKey    string `json:"mission_key"`
		} `json:"states"`
		AlreadyReceivedPoints int  `json:"already_received_points"`
		TotalPoints           int  `json:"total_points"`
		TodayTotalPoints      int  `json:"today_total_points"`
		IsUnclaimed           bool `json:"is_unclaimed"`
		CanGetPoints          int  `json:"can_get_points"`
	} `json:"data"`
}

// SignResponse 签到响应
type SignResponse struct {
	Retcode int `json:"retcode"`
	Data    struct {
		Points int `json:"points"`
	} `json:"data"`
}

// NewMihoyobbs 创建米游社签到实例
func NewMihoyobbs() *Mihoyobbs {
	client := http.NewClient()

	// 设置基础请求头 - 与Python版本保持一致
	headers := map[string]string{
		"DS":                   utils.GetDS(false),
		"cookie":               utils.GetStokenCookie(), // 使用stoken cookie
		"x-rpc-client_type":    utils.MihoyobbsClientType,
		"x-rpc-app_version":    utils.MihoyobbsVersion,
		"x-rpc-sys_version":    "12",
		"x-rpc-channel":        "miyousheluodi",
		"x-rpc-device_id":      mysConfig.GlobalConfig.Device.ID,
		"x-rpc-device_name":    mysConfig.GlobalConfig.Device.Name,
		"x-rpc-device_model":   mysConfig.GlobalConfig.Device.Model,
		"x-rpc-h265_supported": "1",
		"Referer":              "https://app.mihoyo.com",
		"x-rpc-verify_key":     "bll8iq97cem8",
		"x-rpc-csm_source":     "discussion",
		"Content-Type":         "application/json; charset=UTF-8",
		"Host":                 "bbs-api.miyoushe.com",
		"Connection":           "Keep-Alive",
		"Accept-Encoding":      "gzip",
		"User-Agent":           "okhttp/4.9.3",
	}

	client.SetHeaders(headers)

	// 如果有设备指纹，添加
	if mysConfig.GlobalConfig.Device.Fp != "" {
		client.SetHeader("x-rpc-device_fp", mysConfig.GlobalConfig.Device.Fp)
	}

	return &Mihoyobbs{
		client: client,
		bbsList: []BBSInfo{
			{ID: "2", Name: "原神"}, // 对应Python代码中的ID "2"
		},
		taskDo: TaskStatus{
			ReadNum: 3,
			LikeNum: 5,
		},
	}
}

// Run 运行米游社签到
func (m *Mihoyobbs) Run() error {
	if !mysConfig.GlobalConfig.Mihoyobbs.Enable {
		autoLog.Sugar.Info("米游社-功能已禁用")
		return nil
	}

	autoLog.Sugar.Info("米游社-开始签到任务")

	// 获取任务列表
	if err := m.getTasksList(); err != nil {
		autoLog.Sugar.Error("米游社-获取任务列表失败: %v", err)
	}

	// 社区签到
	autoLog.Sugar.Info("米游社-社区签到")
	if mysConfig.GlobalConfig.Mihoyobbs.Checkin {
		if err := m.checkin(); err != nil {
			autoLog.Sugar.Error("米游社-社区签到失败: %v", err)
		}
	}

	utils.RandomSleep(1, 3)

	// 看帖任务
	autoLog.Sugar.Info("米游社-看帖任务")
	if mysConfig.GlobalConfig.Mihoyobbs.Read && !m.taskDo.Read {
		if err := m.readPosts(); err != nil {
			autoLog.Sugar.Error("米游社-看帖任务失败: %v", err)
		}
	}

	utils.RandomSleep(1, 3)

	// 点赞任务
	autoLog.Sugar.Info("米游社-点赞任务")
	if mysConfig.GlobalConfig.Mihoyobbs.Like && !m.taskDo.Like {
		if err := m.likePosts(); err != nil {
			autoLog.Sugar.Error("米游社-点赞任务失败: %v", err)
		}
	}

	utils.RandomSleep(1, 3)

	// 分享任务
	autoLog.Sugar.Info("米游社-分享任务")
	if mysConfig.GlobalConfig.Mihoyobbs.Share && !m.taskDo.Share {
		if err := m.sharePost(); err != nil {
			autoLog.Sugar.Error("米游社-分享任务失败: %v", err)
		}
	}

	autoLog.Sugar.Info("米游社-签到任务完成")
	return nil
}

// getTasksList 获取任务列表
func (m *Mihoyobbs) getTasksList() error {
	return m.getTasksListWithRetry(false)
}

// getTasksListWithRetry 获取任务列表（带重试）
func (m *Mihoyobbs) getTasksListWithRetry(update bool) error {
	autoLog.Sugar.Info("米游社-获取任务列表")

	// 使用专门的task header，与Python版本保持一致
	taskHeader := map[string]string{
		"Accept":           "application/json, text/plain, */*",
		"Origin":           "https://webstatic.mihoyo.com",
		"User-Agent":       "Mozilla/5.0 (Linux; Android 12; Unspecified Device) AppleWebKit/537.36 (KHTML, like Gecko) " + fmt.Sprintf("Version/4.0 Chrome/103.0.5060.129 Mobile Safari/537.36 miHoYoBBS/%s", utils.MihoyobbsVersion),
		"Referer":          "https://webstatic.mihoyo.com",
		"Accept-Encoding":  "gzip, deflate",
		"Accept-Language":  "zh-CN,en-US;q=0.8",
		"X-Requested-With": "com.mihoyo.hyperion",
		"Cookie":           utils.TidyCookie(mysConfig.GlobalConfig.Account.Cookie),
	}

	// 临时设置task header
	originalHeaders := m.client.GetHeaders()
	m.client.SetHeaders(taskHeader)
	defer m.client.SetHeaders(originalHeaders)

	url := "https://bbs-api.miyoushe.com/apihub/wapi/getUserMissionsState?point_sn=myb"
	resp, err := m.client.Get(url)
	if err != nil {
		return err
	}

	// 打印响应内容用于调试
	autoLog.Sugar.Info("米游社-获取任务列表响应: %s", resp.String())

	var taskResp TaskListResponse
	if err := resp.JSON(&taskResp); err != nil {
		//logger.Error("JSON解析失败: %v", err)
		//logger.Error("响应内容: %s", resp.String())
		return fmt.Errorf("JSON解析失败: %v", err)
	}

	if taskResp.Retcode != 0 {
		// 如果失败且尚未尝试更新，则尝试刷新cookie_token
		if !update && taskResp.Retcode == -100 {
			autoLog.Sugar.Info("米游社-获取任务列表失败，尝试刷新cookie_token")
			if err := utils.UpdateCookieToken(); err == nil {
				// 刷新成功，重新设置headers并重试
				return m.getTasksListWithRetry(true)
			}
		}

		if taskResp.Retcode == -100 {
			return fmt.Errorf("获取任务列表失败: Cookie可能已过期，请重新设置Cookie (retcode=%d)", taskResp.Retcode)
		}
		return fmt.Errorf("获取任务列表失败: retcode=%d", taskResp.Retcode)
	}

	// 解析任务状态
	for _, state := range taskResp.Data.States {
		switch state.MissionID {
		case 58: // 签到
			m.taskDo.Sign = state.IsGetAward
		case 59: // 看帖
			m.taskDo.Read = state.IsGetAward
		case 60: // 点赞
			m.taskDo.Like = state.IsGetAward
		case 61: // 分享
			m.taskDo.Share = state.IsGetAward
		}
	}

	autoLog.Sugar.Info("米游社-签到任务状态: 签到=%v, 看帖=%v, 点赞=%v, 分享=%v",
		m.taskDo.Sign, m.taskDo.Read, m.taskDo.Like, m.taskDo.Share)

	return nil
}

// checkin 社区签到
func (m *Mihoyobbs) checkin() error {
	for _, bbs := range m.bbsList {
		autoLog.Sugar.Info("米游社-签到任务")

		// 与Python版本保持一致，使用深拷贝headers
		header := make(map[string]string)
		for k, v := range m.client.GetHeaders() {
			header[k] = v
		}

		// 重试机制，与Python版本保持一致
		for retryCount := 0; retryCount < 2; retryCount++ {
			url := "https://bbs-api.miyoushe.com/apihub/app/api/signIn"
			data := map[string]interface{}{
				"gids": bbs.ID,
			}

			// 为POST请求重新生成DS - 与Python版本保持一致
			jsonData, _ := json.Marshal(data)
			postData := string(jsonData)
			postData = strings.ReplaceAll(postData, " ", "") // 去除空格，与Python保持一致
			ds := utils.GetDS2("", postData)

			autoLog.Sugar.Info("米游社-签到任务请求数据: %s", postData)
			autoLog.Sugar.Info("米游社-签到任务生成的DS: %s", ds)

			// 设置新的DS
			header["DS"] = ds

			// 使用JSON字符串发送请求，与Python版本保持一致
			resp, err := m.client.PostJSONWithHeaders(url, postData, header)
			if err != nil {
				autoLog.Sugar.Error("米游社-签到任务签到请求失败: %v", err)
				continue
			}

			var signResp SignResponse
			if err := resp.JSON(&signResp); err != nil {
				autoLog.Sugar.Error("米游社-签到任务解析签到响应失败: %v", err)
				continue
			}

			// 打印签到响应内容
			autoLog.Sugar.Info("米游社-签到任务签到响应: %s", resp.String())

			if signResp.Retcode == 0 {
				autoLog.Sugar.Info("米游社-签到任务签到成功，获得 %d 米游币", signResp.Data.Points)
				m.todayGetCoins += signResp.Data.Points
				break
			} else if signResp.Retcode == -100 {
				autoLog.Sugar.Error("米游社-签到任务签到失败: Cookie可能已过期，请重新设置Cookie (retcode=%d)", signResp.Retcode)
				// 尝试刷新cookie_token
				if err := utils.UpdateCookieToken(); err == nil {
					autoLog.Sugar.Info("米游社-签到任务CookieToken刷新成功，重新尝试签到")
					// 重新设置headers
					header["cookie"] = utils.GetStokenCookie()
					continue
				}
				return fmt.Errorf("Cookie可能已过期")
			} else if signResp.Retcode == 1008 {
				autoLog.Sugar.Info("米游社-签到任务签到失败: %d", signResp.Retcode)
				break
			} else {
				autoLog.Sugar.Error("米游社-签到任务签到失败: %d", signResp.Retcode)
				if retryCount == 1 {
					break
				}
			}
		}

		utils.RandomSleep(3, 8)
	}

	return nil
}

// readPosts 看帖任务
func (m *Mihoyobbs) readPosts() error {
	autoLog.Sugar.Info("米游社-看帖任务")

	// 获取帖子列表
	posts, err := m.getPostsList()
	if err != nil {
		return err
	}

	readCount := 0
	for _, post := range posts {
		if readCount >= m.taskDo.ReadNum {
			break
		}

		autoLog.Sugar.Info("米游社-看帖任务正在看帖: %s", post.Subject)

		url := "https://bbs-api.miyoushe.com/post/api/getPostFull"
		fullURL := fmt.Sprintf("%s?post_id=%s", url, post.PostID)

		resp, err := m.client.Get(fullURL)
		if err != nil {
			autoLog.Sugar.Error("米游社-看帖任务看帖请求失败: %v", err)
			continue
		}

		var apiResp APIResponse
		if err := resp.JSON(&apiResp); err != nil {
			autoLog.Sugar.Error("米游社-看帖任务解析看帖响应失败: %v", err)
			continue
		}

		if apiResp.Retcode == 0 {
			autoLog.Sugar.Info("米游社-看帖任务看帖成功: %s", post.Subject)
			readCount++
		} else {
			autoLog.Sugar.Error("米游社-看帖任务看帖失败: %s, 错误码: %d", post.Subject, apiResp.Retcode)
		}

		utils.RandomSleep(2, 5)
	}

	return nil
}

// likePosts 点赞任务
func (m *Mihoyobbs) likePosts() error {
	autoLog.Sugar.Info("米游社-点赞任务")

	// 获取帖子列表
	posts, err := m.getPostsList()
	if err != nil {
		return err
	}

	likeCount := 0
	for _, post := range posts {
		if likeCount >= m.taskDo.LikeNum {
			break
		}

		autoLog.Sugar.Info("米游社-点赞任务正在点赞: %s", post.Subject)

		url := "https://bbs-api.miyoushe.com/apihub/sapi/upvotePost"
		data := map[string]interface{}{
			"post_id":   post.PostID,
			"is_cancel": false,
		}

		resp, err := m.client.Post(url, data)
		if err != nil {
			autoLog.Sugar.Error("米游社-点赞任务点赞请求失败: %v", err)
			continue
		}

		var apiResp APIResponse
		if err := resp.JSON(&apiResp); err != nil {
			autoLog.Sugar.Error("米游社-点赞任务解析点赞响应失败: %v", err)
			continue
		}

		if apiResp.Retcode == 0 {
			autoLog.Sugar.Info("米游社-点赞任务点赞成功: %s", post.Subject)
			likeCount++
		} else {
			autoLog.Sugar.Error("米游社-点赞任务点赞失败: %s, 错误码: %d", post.Subject, apiResp.Retcode)
		}

		utils.RandomSleep(2, 5)
	}

	return nil
}

// sharePost 分享任务
func (m *Mihoyobbs) sharePost() error {
	autoLog.Sugar.Info("米游社-分享任务")

	// 获取帖子列表
	posts, err := m.getPostsList()
	if err != nil {
		return err
	}

	if len(posts) == 0 {
		return fmt.Errorf("没有可分享的帖子")
	}

	post := posts[0]
	autoLog.Sugar.Info("米游社-分享任务正在分享: %s", post.Subject)

	url := "https://bbs-api.miyoushe.com/apihub/api/getShareConf"
	fullURL := fmt.Sprintf("%s?entity_id=%s&entity_type=1", url, post.PostID)

	resp, err := m.client.Get(fullURL)
	if err != nil {
		return fmt.Errorf("分享请求失败: %v", err)
	}

	var apiResp APIResponse
	if err := resp.JSON(&apiResp); err != nil {
		return fmt.Errorf("解析分享响应失败: %v", err)
	}

	if apiResp.Retcode == 0 {
		autoLog.Sugar.Info("米游社-分享任务分享成功: %s", post.Subject)
	} else {
		return fmt.Errorf("分享失败: %d", apiResp.Retcode)
	}

	return nil
}

// getPostsList 获取帖子列表
func (m *Mihoyobbs) getPostsList() ([]PostInfo, error) {
	// 使用原神社区的帖子列表
	url := "https://bbs-api.miyoushe.com/post/api/getForumPostList"
	params := "forum_id=26&is_good=false&is_hot=false&page_size=20&sort_type=1"
	fullURL := fmt.Sprintf("%s?%s", url, params)

	resp, err := m.client.Get(fullURL)
	if err != nil {
		return nil, err
	}

	autoLog.Sugar.Debug("米游社-分享任务获取帖子列表响应: %s", resp.String())

	var apiResp struct {
		Retcode int    `json:"retcode"`
		Message string `json:"message"`
		Data    struct {
			List []struct {
				Post struct {
					PostID  string `json:"post_id"`
					Subject string `json:"subject"`
				} `json:"post"`
			} `json:"list"`
		} `json:"data"`
	}

	if err := resp.JSON(&apiResp); err != nil {

		autoLog.Sugar.Error("米游社-分享任务获取帖子列表JSON解析失败: %v", err)
		autoLog.Sugar.Error("米游社-分享任务获取帖子列表响应内容: %s", resp.String())

		return nil, fmt.Errorf("JSON解析失败: %v", err)
	}

	if apiResp.Retcode != 0 {
		return nil, fmt.Errorf("获取帖子列表失败: retcode=%d", apiResp.Retcode)
	}

	// 转换为PostInfo格式
	var posts []PostInfo
	for _, item := range apiResp.Data.List {
		posts = append(posts, PostInfo{
			PostID:  item.Post.PostID,
			Subject: item.Post.Subject,
			ForumID: 26, // 原神社区
		})
	}

	autoLog.Sugar.Info("米游社-分享任务获取帖子列表成功, 帖子数量: %d", len(posts))
	return posts, nil
}
