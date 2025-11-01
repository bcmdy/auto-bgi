package Notice

import (
	"auto-bgi/config"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"net/http"
	"net/url"
)

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

func GetTGBot() *tgbotapi.BotAPI {

	return bot
}
