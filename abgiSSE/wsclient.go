package abgiSSE

import (
	"auto-bgi/Notice"
	"auto-bgi/abgiConstant"
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"auto-bgi/control"
	"auto-bgi/tools"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type AbgiClient struct {
	Conn    *websocket.Conn
	URL     string
	Headers http.Header
	mu      sync.Mutex
}

type Information struct {
	Status string
	Msg    string
	AA     []aa
	Img    Img
}

type Img struct {
	Base64Str string `json:"base64Str"`
	Path      string `json:"path"` //保存路径
}

type aa struct {
	ID       int64
	UID      string
	Name     string
	AbgiType string `json:"abgiType"`
}

var abgiClient *AbgiClient

// 是否是调试模式
var RunDebug bool

var RunCount = 0

var ABgiSeeStatus = "未知"

// Connect 连接 WebSocket 服务器
func Connect(url string, runDebug bool, headers http.Header) (err error) {

	//捕获异常
	defer func() {
		if err := recover(); err != nil {
			err = fmt.Errorf("捕获异常:%v", err)
		}
	}()

	go func() {
		IsNewestVersion()
	}()

	//如果已经在线，就不能请求
	if abgiClient != nil {
		autoLog.Sugar.Error("已经在线，请勿重复上线")
		return fmt.Errorf("已经在线，请勿重复上线")
	}

	//判断上线次数
	if RunCount <= 2 {
		if !abgiConstant.IsVipUser(config.Cfg.Account.Uid) {
			autoLog.Sugar.Infof("今日第%d次上线", RunCount)
			RunCount++
		}

	} else {
		autoLog.Sugar.Error("今日连续三次上线，你是炸弹")
		return fmt.Errorf("今日连续三次上线，你是炸弹")
	}

	dialer := websocket.DefaultDialer
	conn, _, err := dialer.Dial(url, headers)
	if err != nil {
		return fmt.Errorf("连接失败")
	}

	abgiClient = &AbgiClient{
		Conn:    conn,
		URL:     url,
		Headers: headers,
	}

	RunDebug = runDebug

	//// 启动接收消息
	go abgiClient.listen()
	return nil
}

func CheckNetwork() bool {
	_, err := net.LookupHost("www.baidu.com")
	return err == nil
}

func (c *AbgiClient) listen() {
	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			autoLog.Sugar.Error("连接断开")
			// 关闭连接
			c.Conn.Close()
			abgiClient = nil
			//判断是否是因为断网导致重连
			if !CheckNetwork() {
				decrypt, err := tools.Decrypt(config.Cfg.Account.SecretKey)
				if err != nil {
					autoLog.Sugar.Infof("密钥错误")
					return
				}
				runDebug := false
				abgiType := "noDebug"
				//查询今天狗粮批发是什么线路
				dogFoodLine := dogFood.DogFoodIsAOrB()
				//读取联机文件
				num := ReadFile("dogFoodNum.txt")

				if dogFoodLine == "" {
					autoLog.Sugar.Errorf("查询今天狗粮批发线路失败")
					runDebug = true
					abgiType = "debug"
				} else if dogFoodLine == "B" {
					runDebug = true
					abgiType = "debug"
				} else if num >= 100 {
					autoLog.Sugar.Infof("今天狗粮已经跑完，无需再跑")
					runDebug = true
					abgiType = "debug"
				}

				for i := range 5 {
					autoLog.Sugar.Errorf("网络断开，导致下线，开始重连,第%d次", i)
					Notice.SentText(fmt.Sprintf("网络断开，导致下线，开始重连,第%d次", i))
					//等待1分钟
					time.Sleep(time.Minute)
					//重新连接
					err := Connect(fmt.Sprintf("ws://%s/api/abgiWs/%s/%s/%s", decrypt, abgiType, config.Cfg.Account.Uid, config.Cfg.Account.Name), runDebug, nil)
					if err != nil {
						if err.Error() == "已经在线，请勿重复上线" {
							autoLog.Sugar.Infof("已经在线，请勿重复上线")
							break
						} else {
							autoLog.Sugar.Errorf("重连失败:%v", err)
						}

					} else {
						autoLog.Sugar.Infof("重新连接成功")
						break
					}
				}

			} else {
				autoLog.Sugar.Infof("当前网络正常，自然下线")
			}
			return
		}
		var info Information
		err = json.Unmarshal(msg, &info)
		if err != nil {
			log.Println()
			autoLog.Sugar.Infof("未知消息:%s", msg)
			continue
		}

		if info.Status == "1" {
			//联机狗粮
			autoLog.Sugar.Infof("联机启动")
			artifactsGroupPurchasing(info, true)
			ABgiSeeStatus = "联机中"

			//启动配置组
			err2 := startGroups([]string{config.Cfg.Account.GouLangGroupName})
			if err2 != nil {
				autoLog.Sugar.Errorf("启动配置组失败: %v", err)
				return
			}

		} else if info.Status == "3" {

			//autoLog.Sugar.Infof("联机准备")
			//artifactsGroupPurchasing(info, true)
			//ABgiSeeStatus = "联机中"
			////启动配置组
			//err2 := startGroups([]string{config.Cfg.Account.GouLangGroupName})
			//if err2 != nil {
			//	autoLog.Sugar.Errorf("启动配置组失败: %v", err)
			//	return
			//}
			////处理图片消息
			//HandleImg(info)

		} else if info.Status == "4" {
			//联机启动-全部调试调试
			autoLog.Sugar.Infof("联机启动-非备用组4个调试")
			artifactsGroupPurchasing(info, false)
			ABgiSeeStatus = "联机中"
			//启动配置组
			err2 := startGroups([]string{config.Cfg.Account.GouLangGroupName})
			if err2 != nil {
				autoLog.Sugar.Errorf("启动配置组失败: %v", err)
				return
			}

		} else {
			autoLog.Sugar.Infof("收到消息:%s", info.Msg)
		}

	}
}

// StartGroups 启动配置组
func startGroups(names []string) error {
	control.CloseSoftware()

	//暂停500毫秒
	time.Sleep(500 * time.Millisecond)

	betterGIPath := filepath.Join(config.Cfg.BetterGIAddress, "BetterGI.exe")

	// 检查文件是否存在
	if _, err := os.Stat(betterGIPath); err != nil {
		autoLog.Sugar.Errorf("BetterGI.exe 不存在: %v", err)
		return err
	}

	args := append([]string{"--startGroups"}, names...) // 每个组名单独参数

	maxRetries := 3
	var err error

	for i := 0; i < maxRetries; i++ {
		cmd := exec.Command(betterGIPath, args...)
		cmd.Dir = config.Cfg.BetterGIAddress // 设置工作目录，确保能读取到配置文件
		cmd.Stdout = nil
		cmd.Stderr = nil

		err = cmd.Start()
		if err != nil {
			autoLog.Sugar.Errorf("启动命令执行失败 (第 %d 次尝试): %v", i+1, err)
		} else {
			autoLog.Sugar.Infof("启动命令已下发，等待 20 秒检查进程状态 (第 %d 次尝试)...", i+1)
		}

		// 等待20秒检查进程
		time.Sleep(20 * time.Second)

		if control.CheckProcessRunning("BetterGI.exe") {
			autoLog.Sugar.Infof("检测到 BetterGI 进程运行正常，启动成功: %v", names)
			return nil
		}

		autoLog.Sugar.Warnf("未检测到 BetterGI 进程 (第 %d 次尝试)，将重试...", i+1)
	}

	return fmt.Errorf("启动配置组失败，已重试 %d 次: %w", maxRetries, err)
}

// Send 发送消息
func Send(message string) error {
	if abgiClient == nil {
		return fmt.Errorf("WebSocket 未连接")
	}
	abgiClient.mu.Lock()
	defer abgiClient.mu.Unlock()
	return abgiClient.Conn.WriteMessage(websocket.TextMessage, []byte(message))
}

// Status 返回当前连接状态
func Status() {
	if abgiClient == nil {
		ABgiSeeStatus = "离线"
	} else {
		ABgiSeeStatus = "在线"
	}
}

type OnlineUser struct {
	GroupName   string    `json:"group_name"`
	Count       int64     `json:"count"`
	Description string    `json:"description"`
	Members     []Members `json:"members"`
}

type Members struct {
	Name        string `json:"name"`
	IsSimulated bool   `json:"is_simulated"`
	AbgiType    string `json:"abgi_type"`
}

// 获取在线人数
func GroupsStatusHandler() interface{} {
	decrypt, err2 := tools.Decrypt(config.Cfg.Account.SecretKey)
	if err2 != nil {
		return 0
	}
	resp, err := http.Get(fmt.Sprintf("http://%s/api/GroupsStatusHandler?uid="+config.Cfg.Account.Uid, decrypt))
	if err != nil {
		autoLog.Sugar.Error("获取在线用户失败:")
		return 0
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		autoLog.Sugar.Error("读取响应失败:", err)
		return 0
	}
	var mapData []OnlineUser
	err = json.Unmarshal(body, &mapData)
	if err != nil {
		autoLog.Sugar.Error("解析JSON失败:", err)
		return 0
	}
	return mapData
}

// Close 关闭连接
func Close() {
	if abgiClient != nil {
		abgiClient.Conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "bye"))
		abgiClient.Conn.Close()
		abgiClient = nil
	}
}
