package gamecheckin

import (
	"encoding/json"
	"fmt"

	"auto-bgi/internal/http"
	"auto-bgi/internal/logger"
	"auto-bgi/internal/mysConfig"
	"auto-bgi/internal/utils"
)

// GameCheckin 游戏签到
type GameCheckin struct {
	client         *http.Client
	gameID         string
	gameMid        string
	gameName       string
	actID          string
	playerName     string
	rewardsAPI     string
	isSignAPI      string
	signAPI        string
	checkinRewards []RewardInfo
	accountList    []AccountInfo
}

// AccountInfo 账号信息
type AccountInfo struct {
	GameUID  string `json:"game_uid"`
	Region   string `json:"region"`
	Nickname string `json:"nickname"`
	Level    int    `json:"level"`
}

// RewardInfo 奖励信息
type RewardInfo struct {
	Name string `json:"name"`
	Cnt  int    `json:"cnt"`
}

// API响应结构
type APIResponse struct {
	Retcode int         `json:"retcode"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// AccountListResponse 账号列表响应
type AccountListResponse struct {
	Retcode int `json:"retcode"`
	Data    struct {
		List []AccountInfo `json:"list"`
	} `json:"data"`
}

// RewardsResponse 奖励列表响应
type RewardsResponse struct {
	Retcode int    `json:"retcode"`
	Message string `json:"message"`
	Data    struct {
		Awards []RewardInfo `json:"awards"`
	} `json:"data"`
}

// IsSignResponse 签到状态响应
type IsSignResponse struct {
	Retcode int    `json:"retcode"`
	Message string `json:"message"`
	Data    struct {
		IsSign  bool `json:"is_sign"`
		SignDay int  `json:"sign_day"`
	} `json:"data"`
}

// SignResponse 签到响应
type SignResponse struct {
	Retcode int    `json:"retcode"`
	Message string `json:"message"`
	Data    struct {
		Code string `json:"code"`
	} `json:"data"`
}

// GameInfo 游戏信息
type GameInfo struct {
	ID      string
	Mid     string
	Name    string
	ActID   string
	Checkin bool
}

// 游戏配置 - 只保留原神
var GameConfigs = map[string]GameInfo{
	"genshin": {
		ID:      "hk4e_cn",
		Mid:     "genshin",
		Name:    "原神",
		ActID:   "e202311201442471",
		Checkin: true,
	},
}

// NewGameCheckin 创建游戏签到实例
func NewGameCheckin(gameID, gameMid, gameName, actID, playerName string) *GameCheckin {
	client := http.NewClient()

	// 设置基础请求头 - 与Python版本保持一致
	headers := map[string]string{
		"DS":                utils.GetDS(true),
		"Referer":           "https://act.mihoyo.com/",
		"Cookie":            utils.TidyCookie(mysConfig.GlobalConfig.Account.Cookie), // 游戏签到使用普通cookie
		"x-rpc-device_id":   mysConfig.GlobalConfig.Device.ID,
		"User-Agent":        "Mozilla/5.0 (Linux; Android 12; Unspecified Device) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/103.0.5060.129 Mobile Safari/537.36 miHoYoBBS/" + utils.MihoyobbsVersion,
		"Accept":            "application/json, text/plain, */*",
		"x-rpc-channel":     "miyousheluodi",
		"Origin":            "https://webstatic.mihoyo.com",
		"x-rpc-app_version": utils.MihoyobbsVersion,
		"x-rpc-client_type": "5",
		"Accept-Language":   "zh-CN,en-US;q=0.8",
		"Accept-Encoding":   "gzip, deflate",
		"X-Requested-With":  "com.mihoyo.hyperion",
		"Connection":        "keep-alive",
	}

	// 为原神设置特殊头部 - 与Python版本保持一致
	if gameID == "hk4e_cn" {
		headers["Origin"] = "https://act.mihoyo.com" // 原神使用不同的Origin
		headers["x-rpc-signgame"] = "hk4e"           // 原神特有的签名游戏头部
	}

	client.SetHeaders(headers)

	gameCheckin := &GameCheckin{
		client:     client,
		gameID:     gameID,
		gameMid:    gameMid,
		gameName:   gameName,
		actID:      actID,
		playerName: playerName,
		rewardsAPI: "https://api-takumi.mihoyo.com/event/luna/home?lang=zh-cn",
		isSignAPI:  "https://api-takumi.mihoyo.com/event/luna/info?lang=zh-cn",
		signAPI:    "https://api-takumi.mihoyo.com/event/luna/sign",
	}

	return gameCheckin
}

// updateHeaders 更新请求头（当cookie_token被刷新时调用）
func (g *GameCheckin) updateHeaders() {
	// 重新设置基础请求头 - 与Python版本保持一致
	headers := map[string]string{
		"DS":                utils.GetDS(true),
		"Referer":           "https://act.mihoyo.com/",
		"Cookie":            utils.TidyCookie(mysConfig.GlobalConfig.Account.Cookie), // 游戏签到使用普通cookie
		"x-rpc-device_id":   mysConfig.GlobalConfig.Device.ID,
		"User-Agent":        "Mozilla/5.0 (Linux; Android 12; Unspecified Device) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/103.0.5060.129 Mobile Safari/537.36 miHoYoBBS/" + utils.MihoyobbsVersion,
		"Accept":            "application/json, text/plain, */*",
		"x-rpc-channel":     "miyousheluodi",
		"Origin":            "https://webstatic.mihoyo.com",
		"x-rpc-app_version": utils.MihoyobbsVersion,
		"x-rpc-client_type": "5",
		"Accept-Language":   "zh-CN,en-US;q=0.8",
		"Accept-Encoding":   "gzip, deflate",
		"X-Requested-With":  "com.mihoyo.hyperion",
		"Connection":        "keep-alive",
	}

	// 为原神设置特殊头部 - 与Python版本保持一致
	if g.gameID == "hk4e_cn" {
		headers["Origin"] = "https://act.mihoyo.com" // 原神使用不同的Origin
		headers["x-rpc-signgame"] = "hk4e"           // 原神特有的签名游戏头部
	}

	g.client.SetHeaders(headers)
}

// Run 运行游戏签到
func (g *GameCheckin) Run() error {
	if !mysConfig.GlobalConfig.Games.CN.Enable {
		logger.Info("国服游戏签到功能已禁用")
		return nil
	}

	logger.Info("开始游戏签到: %s", g.gameName)

	// 获取账号列表
	if err := g.getAccountList(); err != nil {
		return fmt.Errorf("获取账号列表失败: %v", err)
	}

	if len(g.accountList) == 0 {
		logger.Info("没有找到 %s 的账号", g.gameName)
		return nil
	}

	// 获取签到奖励列表
	if err := g.getCheckinRewards(); err != nil {
		logger.Error("获取签到奖励列表失败: %v", err)
	}

	// 为每个账号签到
	for _, account := range g.accountList {
		if err := g.checkinAccount(account); err != nil {
			logger.Error("账号 %s 签到失败: %v", account.GameUID, err)
		}
	}

	logger.Info("游戏签到完成: %s", g.gameName)
	return nil
}

// getAccountList 获取账号列表
func (g *GameCheckin) getAccountList() error {
	logger.Info("正在获取 %s 账号列表", g.gameName)

	url := fmt.Sprintf("https://api-takumi.mihoyo.com/binding/api/getUserGameRolesByCookie?game_biz=%s", g.gameID)

	resp, err := g.client.Get(url)
	if err != nil {
		return err
	}

	// 打印响应内容用于调试
	logger.Info("账号列表响应: %s", resp.String())

	var accountResp AccountListResponse
	if err := resp.JSON(&accountResp); err != nil {
		logger.Error("JSON解析失败: %v", err)
		logger.Error("响应内容: %s", resp.String())
		return fmt.Errorf("JSON解析失败: %v", err)
	}

	if accountResp.Retcode != 0 {
		return fmt.Errorf("获取账号列表失败: retcode=%d", accountResp.Retcode)
	}

	g.accountList = accountResp.Data.List
	logger.Info("找到 %d 个 %s 账号", len(g.accountList), g.gameName)

	return nil
}

// getCheckinRewards 获取签到奖励列表
func (g *GameCheckin) getCheckinRewards() error {
	logger.Info("正在获取 %s 签到奖励列表", g.gameName)

	url := fmt.Sprintf("%s&act_id=%s", g.rewardsAPI, g.actID)

	resp, err := g.client.Get(url)
	if err != nil {
		return err
	}

	// 打印响应内容用于调试
	//logger.Info("签到奖励列表响应: %s", resp.String())

	var rewardsResp RewardsResponse
	if err := resp.JSON(&rewardsResp); err != nil {
		return err
	}

	if rewardsResp.Retcode != 0 {
		if rewardsResp.Retcode == -500001 {
			return fmt.Errorf("获取签到奖励列表失败: Cookie可能已过期或权限不足 (retcode=%d, message=%s)", rewardsResp.Retcode, rewardsResp.Message)
		}
		return fmt.Errorf("获取签到奖励列表失败: %d", rewardsResp.Retcode)
	}

	g.checkinRewards = rewardsResp.Data.Awards
	logger.Info("获取到 %d 个签到奖励", len(g.checkinRewards))

	return nil
}

// checkinAccount 为单个账号签到
func (g *GameCheckin) checkinAccount(account AccountInfo) error {
	logger.Info("正在为账号 %s (%s) 签到", account.GameUID, account.Nickname)

	// 检查是否已签到 - 如果检查失败，直接尝试签到（与Python行为一致）
	isSigned, err := g.isSign(account.Region, account.GameUID, false)
	if err != nil {
		logger.Info("签到状态检查失败，直接尝试签到: %v", err)
	} else if isSigned {
		logger.Info("账号 %s 今日已签到", account.GameUID)
		return nil
	}

	// 执行签到
	retries := mysConfig.GlobalConfig.Games.CN.Retries
	for i := 1; i <= retries; i++ {
		if i > 1 {
			logger.Info("第 %d 次重试签到", i)
			utils.RandomSleep(5, 10)
		}

		if err := g.sign(account.Region, account.GameUID); err != nil {
			if i == retries {
				return fmt.Errorf("签到失败: %v", err)
			}
			logger.Error("签到失败，准备重试: %v", err)
			continue
		}

		logger.Info("账号 %s 签到成功", account.GameUID)
		break
	}

	return nil
}

// isSign 检查是否已签到
func (g *GameCheckin) isSign(region, uid string, update bool) (bool, error) {
	url := fmt.Sprintf("%s&act_id=%s&region=%s&uid=%s", g.isSignAPI, g.actID, region, uid)

	resp, err := g.client.Get(url)
	if err != nil {
		return false, err
	}

	// 打印响应内容用于调试
	logger.Info("签到状态检查响应: %s", resp.String())

	var isSignResp IsSignResponse
	if err := resp.JSON(&isSignResp); err != nil {
		return false, err
	}

	if isSignResp.Retcode != 0 {
		// 如果失败且尚未尝试更新，则尝试刷新cookie_token
		if !update && (isSignResp.Retcode == -100 || isSignResp.Retcode == -500001) {
			logger.Info("检查签到状态失败，尝试刷新cookie_token")
			if err := utils.UpdateCookieToken(); err == nil {
				// 刷新成功，重新设置headers并重试
				g.updateHeaders()
				return g.isSign(region, uid, true)
			} else {
				logger.Error("刷新cookie_token失败: %v", err)
			}
		}

		if isSignResp.Retcode == -500001 {
			return false, fmt.Errorf("检查签到状态失败: Cookie可能已过期或权限不足 (retcode=%d, message=%s)", isSignResp.Retcode, isSignResp.Message)
		}
		return false, fmt.Errorf("检查签到状态失败: %d", isSignResp.Retcode)
	}

	return isSignResp.Data.IsSign, nil
}

// sign 执行签到
func (g *GameCheckin) sign(region, uid string) error {
	url := g.signAPI

	data := map[string]interface{}{
		"act_id": g.actID,
		"region": region,
		"uid":    uid,
	}

	// 序列化为JSON字符串，与Python的json参数保持一致
	jsonData, _ := json.Marshal(data)

	// 打印发送的JSON数据用于调试
	logger.Info("发送的JSON数据: %s", string(jsonData))

	// Python游戏签到不重新生成DS，直接使用初始化时的DS
	resp, err := g.client.PostJSON(url, string(jsonData))
	if err != nil {
		return err
	}

	// 打印响应内容用于调试
	logger.Info("签到请求响应: %s", resp.String())

	var signResp SignResponse
	if err := resp.JSON(&signResp); err != nil {
		return err
	}

	if signResp.Retcode != 0 {
		return fmt.Errorf("签到失败: %d", signResp.Retcode)
	}

	return nil
}

// RunAllGames 运行所有游戏签到
func RunAllGames() error {
	logger.Info("开始游戏签到")

	// 检查国服游戏签到是否启用
	if !mysConfig.GlobalConfig.Games.CN.Enable {
		logger.Info("国服游戏签到功能已禁用")
		return nil
	}

	// 只处理原神
	gameInfo := GameConfigs["genshin"]

	// 检查原神是否启用签到
	if !mysConfig.GlobalConfig.Games.CN.Genshin.Checkin {
		logger.Info("原神签到已禁用")
		return nil
	}

	// 创建游戏签到实例并运行
	gameCheckin := NewGameCheckin(
		gameInfo.ID,
		gameInfo.Mid,
		gameInfo.Name,
		gameInfo.ActID,
		"玩家",
	)

	if err := gameCheckin.Run(); err != nil {
		logger.Error("原神签到失败: %v", err)
	}

	logger.Info("游戏签到完成")
	return nil
}
