package abgiNewSEE

import (
	"auto-bgi/AbgiBot"
	"auto-bgi/autoLog"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
)

type AbgiClient struct {
	Conn    *websocket.Conn
	URL     string
	Headers http.Header
	mu      sync.Mutex
}

var abgiClient *AbgiClient

func Connect(url string) (err error) {
	//捕获异常
	defer func() {
		if err := recover(); err != nil {
			err = fmt.Errorf("捕获异常:%v", err)
			fmt.Println(err)
			return
		}
	}()

	//如果已经在线，就不能请求
	if abgiClient != nil {
		autoLog.Sugar.Error("已经在线，请勿重复上线")
		return fmt.Errorf("已经在线，请勿重复上线")
	}
	dialer := websocket.DefaultDialer
	conn, _, err := dialer.Dial(url, nil)
	if err != nil {
		return fmt.Errorf("连接失败")
	}

	abgiClient = &AbgiClient{
		Conn:    conn,
		URL:     url,
		Headers: nil,
	}

	//// 启动接收消息
	go abgiClient.listen()
	return nil
}

type ReceiveMessage struct {
	Event       string                 `json:"event"`
	ReceiveData map[string]interface{} `json:"data"`
}

func (c *AbgiClient) listen() {
	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			autoLog.Sugar.Error("接收消息失败:%v", err)
			// 关闭连接
			c.Conn.Close()
			abgiClient = nil
			return
		}
		var receive ReceiveMessage
		err = json.Unmarshal(msg, &receive)
		if err != nil {
			autoLog.Sugar.Error("解析消息失败:%v", err)
			continue
		}
		if receive.Event == "abgi/message" {
			response := AbgiBot.BotCommand(receive.ReceiveData["content"].(string))
			autoLog.Sugar.Info("收到消息:%v", string(msg))
			autoLog.Sugar.Infof("回复消息:%v", response)
			c.Send(response)
		}

	}
}

type SendMessage struct {
	Event string `json:"event"`
	Data  Data   `json:"data"`
}
type Data struct {
	Content string `json:"content"`
}

// 发送消息
func (c *AbgiClient) Send(msg string) error {
	if abgiClient == nil {
		return fmt.Errorf("WebSocket 未连接")
	}
	abgiClient.mu.Lock()
	defer abgiClient.mu.Unlock()
	var send SendMessage
	send.Event = "abgi/send"
	send.Data.Content = msg

	//转换为json字符串
	sendJson, err := json.Marshal(send)
	if err != nil {
		return fmt.Errorf("转换为json字符串失败:%v", err)
	}

	return abgiClient.Conn.WriteMessage(websocket.TextMessage, sendJson)
}
