package abgiSSE

import (
	"auto-bgi/Notice"
	"auto-bgi/ScriptGroup"
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	"sync"
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
}

type aa struct {
	ID   int64
	UID  string
	Name string
}

var abgiClient *AbgiClient

// Connect 连接 WebSocket 服务器
func Connect(url string, headers http.Header) error {
	dialer := websocket.DefaultDialer
	conn, _, err := dialer.Dial(url, headers)
	if err != nil {
		return fmt.Errorf("连接 WebSocket 失败: %w", err)
	}

	abgiClient = &AbgiClient{
		Conn:    conn,
		URL:     url,
		Headers: headers,
	}

	// 启动接收消息
	go abgiClient.listen()
	return nil
}

var scriptGroupConfig ScriptGroup.ScriptGroupConfig

func (c *AbgiClient) listen() {
	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			log.Println("WebSocket 连接断开:", err)
			return
		}

		var info Information
		fmt.Println(string(msg))
		err = json.Unmarshal(msg, &info)
		if err != nil {
			log.Println("解析消息失败:", err)
			continue
		}

		if info.Status == "1" {
			autoLog.Sugar.Infof("联机启动")
			for _, v := range info.AA {
				autoLog.Sugar.Infof("玩家 %s 加入联机", v.Name)
			}
			//转成map
			var dd []map[string]interface{}
			for _, v := range info.AA {
				dd = append(dd, map[string]interface{}{
					"ID":   v.ID,
					"UID":  v.UID,
					"Name": v.Name,
				})
			}

			scriptGroupConfig.StartDogFoodOnline(dd)
		}

		//fmt.Printf("收到消息: %s\n", msg)
		Notice.SentText(string(msg))
	}
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
func Status() string {
	if abgiClient == nil {
		return "未连接"
	}
	return "已连接到 " + abgiClient.URL
}

// 获取在线人数
func GetAllOnlineUser() interface{} {
	decrypt, err2 := Decrypt(config.Cfg.Account.SecretKey, config.Cfg.Account.AccountKey)
	if err2 != nil {
		autoLog.Sugar.Infof("密钥错误:%s", err2)
		return 0
	}
	resp, err := http.Get(fmt.Sprintf("http://%s/api/GetAllOnlineUser?", decrypt))
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
	autoLog.Sugar.Infof("当前在线用户: %s", body)
	return string(body)
}

// Close 关闭连接
func Close() {
	if abgiClient != nil {
		abgiClient.Conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "bye"))
		abgiClient.Conn.Close()
		abgiClient = nil
	}
}
