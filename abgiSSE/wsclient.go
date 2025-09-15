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
	Img    Img
}

type Img struct {
	Base64Str string `json:"base64Str"`
	Path      string `json:"path"` //保存路径
}

type aa struct {
	ID   int64
	UID  string
	Name string
}

var abgiClient *AbgiClient

// 是否是调试模式
var RunDebug bool

// Connect 连接 WebSocket 服务器
func Connect(url string, runDebug bool, headers http.Header) error {

	//如果已经在线，就不能请求
	if abgiClient != nil {
		autoLog.Sugar.Error("已经在线，请勿重复上线")
		return fmt.Errorf("已经在线，请勿重复上线")
	}

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

	RunDebug = runDebug

	autoLog.Sugar.Infof("是否是调试模式:%v", runDebug)

	// 启动接收消息
	go abgiClient.listen()
	return nil
}

var scriptGroupConfig ScriptGroup.ScriptGroupConfig

func (c *AbgiClient) listen() {
	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			log.Println("WebSocket 连接断开:", err.Error())
			// 关闭连接
			c.Conn.Close()
			abgiClient = nil
			return
		}

		var info Information
		err = json.Unmarshal(msg, &info)
		if err != nil {
			log.Println("解析消息失败:", err)
			continue
		}

		if info.Status == "1" {
			autoLog.Sugar.Infof("联机启动")
			//转成map
			xiaoxi := ""
			var dd []map[string]interface{}
			var names []string
			for _, v := range info.AA {
				dd = append(dd, map[string]interface{}{
					"ID":   v.ID,
					"UID":  v.UID,
					"Name": v.Name,
				})
				xiaoxi += fmt.Sprintf("序号%d -- %s\n", v.ID, v.Name)
				names = append(names, v.Name)
			}
			//生成图片
			if len(names) > 0 {
				//生成图片
				for _, name := range names {
					NameToImage(name)
				}
			}

			scriptGroupConfig.StartDogFoodOnline(RunDebug, dd)
			fmt.Println(xiaoxi)
			Notice.SentText(xiaoxi)
		} else if info.Status == "2" {

			//处理图片消息
			HandleImg(info)

		} else {
			autoLog.Sugar.Infof("收到消息:%s", info.Msg)
		}

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

type OnlineUser struct {
	GroupName string    `json:"group_name"`
	Count     int64     `json:"count"`
	Members   []Members `json:"members"`
}

type Members struct {
	Name        string `json:"name"`
	IsSimulated bool   `json:"is_simulated"`
}

// 获取在线人数
func GroupsStatusHandler() interface{} {
	decrypt, err2 := Decrypt(config.Cfg.Account.SecretKey, config.Cfg.Account.AccountKey)
	if err2 != nil {
		return 0
	}
	resp, err := http.Get(fmt.Sprintf("http://%s/api/GroupsStatusHandler?", decrypt))
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
