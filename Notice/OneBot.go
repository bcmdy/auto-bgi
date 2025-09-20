package Notice

import (
	"auto-bgi/config"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// 消息段
type MessageSegment struct {
	Type string            `json:"type"`
	Data map[string]string `json:"data"`
}

// 私聊请求
type PrivateMsgRequest struct {
	UserID  int              `json:"user_id"`
	Message []MessageSegment `json:"message"`
}

// 群聊请求
type GroupMsgRequest struct {
	GroupID int              `json:"group_id"`
	Message []MessageSegment `json:"message"`
}

type OneBotClient struct {
}

// doPost 统一请求
func (c *OneBotClient) doPost(path string, data interface{}) error {
	url := config.Cfg.Notice.OneBot.APIBase + path
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if config.Cfg.Notice.OneBot.Token != "" {
		req.Header.Set("Authorization", "Bearer "+config.Cfg.Notice.OneBot.Token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("请求失败: %s", resp.Status)
	}
	return nil
}

// ------------------ 私聊 ------------------

// SendPrivateText 私聊文字
func (c *OneBotClient) SendPrivateText(text string) error {
	body := PrivateMsgRequest{
		UserID: config.Cfg.Notice.OneBot.QQNum,
		Message: []MessageSegment{
			{Type: "text", Data: map[string]string{"text": text}},
		},
	}
	return c.doPost("/send_private_msg", body)
}

// SendPrivateWithImage 私聊文字 + 本地图片
func (c *OneBotClient) SendPrivateWithImage(imagePath string) error {
	imgBytes, err := os.ReadFile(imagePath)
	if err != nil {
		return err
	}
	b64 := base64.StdEncoding.EncodeToString(imgBytes)

	body := PrivateMsgRequest{
		UserID: config.Cfg.Notice.OneBot.QQNum,
		Message: []MessageSegment{
			{Type: "text", Data: map[string]string{"text": ""}},
			{Type: "image", Data: map[string]string{"file": "base64://" + b64}},
		},
	}
	return c.doPost("/send_private_msg", body)
}
