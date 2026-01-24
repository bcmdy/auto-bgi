package abgiSSE

import (
	"auto-bgi/abgiConstant"
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"auto-bgi/control"
	"auto-bgi/tools"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
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

func (c *AbgiClient) listen() {
	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			autoLog.Sugar.Error("连接断开")
			// 关闭连接
			c.Conn.Close()
			abgiClient = nil
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

			//启动配置组
			err2 := startGroups([]string{config.Cfg.Account.GouLangGroupName})
			if err2 != nil {
				autoLog.Sugar.Errorf("启动配置组失败: %v", err)
				return
			}

		} else if info.Status == "3" {

			autoLog.Sugar.Infof("联机准备")
			artifactsGroupPurchasing(info, true)

			////处理图片消息
			//HandleImg(info)

		} else if info.Status == "4" {
			//联机启动-全部调试调试
			autoLog.Sugar.Infof("联机启动-非备用组4个调试")
			artifactsGroupPurchasing(info, false)

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
	time.Sleep(1 * time.Second)

	betterGIPath := filepath.Join(config.Cfg.BetterGIAddress, "BetterGI.exe")

	// 检查文件是否存在
	if _, err := os.Stat(betterGIPath); err != nil {
		autoLog.Sugar.Errorf("BetterGI.exe 不存在: %v", err)
		return err
	}

	args := append([]string{"--startGroups"}, names...) // 每个组名单独参数

	cmd := exec.Command(betterGIPath, args...)

	cmd.Stdout = nil
	cmd.Stderr = nil
	err := cmd.Start()
	if err != nil {
		autoLog.Sugar.Errorf("启动配置组失败: %v", err)
		return err
	}
	autoLog.Sugar.Infof("启动配置组成功: %v", names)
	return nil
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

func Status() bool {
	if abgiClient == nil {
		return false
	}
	return true
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
