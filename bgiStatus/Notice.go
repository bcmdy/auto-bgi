package bgiStatus

import (
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

func init() {

	if err := InitTG(config.Cfg.Notice.TGNotice.TGToken, config.Cfg.Notice.TGNotice.Proxy); err != nil {

		sprintf := fmt.Sprintf("Telegram bot初始化失败: %v", err)
		fmt.Println(sprintf)
	} else {
		fmt.Println("Telegram bot配置成功")
	}
}

// 向企业微信发送通知（文本）
func sendWeChatNotification(content string) {

	// 通知内容
	message := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			//"content": "BetterGI 已经关闭:\n" + Config.Content + "/test",
			"content": content,
		},
	}
	jsonData, err := json.Marshal(message)
	if err != nil {
		autoLog.Sugar.Error("Error marshaling JSON:", err)
		return
	}

	client := &http.Client{}

	req, err := http.NewRequest("POST", config.Cfg.Notice.Wechat, bytes.NewBuffer(jsonData))
	if err != nil {

		autoLog.Sugar.Error("Error creating request:", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		autoLog.Sugar.Error("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		autoLog.Sugar.Error("企业微信机器人配置错误:", resp.Status)

	} else {
		autoLog.Sugar.Info("企业微信机器人配置成功:", resp.Status)
	}
}

// 向企业微信发送通知（图片）
func sendWeChatImage(path string) error {

	//获取本地文件
	// 读取图片文件
	imageData, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Printf("Error reading image file: %v\n", err)
		return err
	}
	// 计算 Base64 编码
	base64Data := base64.StdEncoding.EncodeToString(imageData)

	// 计算 MD5 哈希
	md5Hash := md5.Sum(imageData)
	md5String := hex.EncodeToString(md5Hash[:])

	// 通知内容
	message := map[string]interface{}{
		"msgtype": "image",
		"image": map[string]string{
			"base64": base64Data,
			"md5":    md5String,
		},
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		autoLog.Sugar.Error("Error marshaling JSON:", err)
		return err
	}

	client := &http.Client{}

	req, err := http.NewRequest("POST", config.Cfg.Notice.Wechat, bytes.NewBuffer(jsonData))
	if err != nil {

		autoLog.Sugar.Error("Error creating request:", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {

		autoLog.Sugar.Error("Error sending request:", err)
		return err
	}
	defer resp.Body.Close()

	return nil
}

var bot *tgbotapi.BotAPI

// 初始化机器人（带代理）
func InitTG(token, proxy string) error {
	if token == "" {
		return nil // 允许空 token，跳过初始化
	}
	var client *http.Client
	if proxy != "" {
		pu, err := url.Parse(proxy)
		if err != nil {
			return err
		}
		client = &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(pu)}}
	} else {
		client = http.DefaultClient
	}
	var err error
	bot, err = tgbotapi.NewBotAPIWithClient(token, tgbotapi.APIEndpoint, client)
	if err != nil {
		fmt.Println("TG配置错误", err)
		return err
	}
	log.Printf("[TG] bot authorized: @%s", bot.Self.UserName)
	return nil
}

// 发送纯文本
func sendTGNotification(text string) error {
	if bot == nil {
		return fmt.Errorf("TG配置错误")
	}
	_, err := bot.Send(tgbotapi.NewMessage(config.Cfg.Notice.TGNotice.ChatID, text))
	return err
}

// 发送图片（本地路径）
func sendTGImage(path string) error {
	if bot == nil {
		return fmt.Errorf("TG配置错误")
	}
	photo := tgbotapi.NewPhoto(config.Cfg.Notice.TGNotice.ChatID, tgbotapi.FilePath(path))
	_, err := bot.Send(photo)
	return err
}

func SentText(text string) {
	if config.Cfg.Notice.Type == "TG" {
		err := sendTGNotification(text)
		if err != nil {
			autoLog.Sugar.Error("通知-TG文本发送失败:", err)
		}
		return
	} else if config.Cfg.Notice.Type == "Wechat" {
		sendWeChatNotification(text)
		return
	}

	autoLog.Sugar.Error("通知-文本未知通知类型")
}

func SentImage(path string) error {
	if config.Cfg.Notice.Type == "TG" {
		err := sendTGImage(path)
		if err != nil {
			autoLog.Sugar.Error("通知-TG图片发送失败:", err)
		}
		return fmt.Errorf("通知-TG图片发送失败:%v", err)
	} else if config.Cfg.Notice.Type == "Wechat" {
		err := sendWeChatImage(path)
		if err != nil {
			autoLog.Sugar.Error("通知-微信图片发送失败:", err)
		}
		return fmt.Errorf("通知-微信图片发送失败:%v", err)
	}
	autoLog.Sugar.Error("通知-图片未知通知类型")
	return fmt.Errorf("通知-图片未知通知类型")
}
