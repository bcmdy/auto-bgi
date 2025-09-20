package Notice

import (
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"auto-bgi/control"
	"fmt"
)

func init() {

	if config.Cfg.Notice.Type == "TG" {
		if err := InitTG(config.Cfg.Notice.TGNotice.TGToken, config.Cfg.Notice.TGNotice.Proxy); err != nil {

			sprintf := fmt.Sprintf("Telegram bot初始化失败: %v", err)
			fmt.Println(sprintf)
		} else {
			fmt.Println("Telegram bot配置成功")
		}
	}

}

var oneBot OneBotClient

func SentText(text string) {

	switch config.Cfg.Notice.Type {
	case "TG":
		err := sendTGNotification(text)
		if err != nil {
			autoLog.Sugar.Error("通知-TG文本发送失败:", err)
		}
		return
	case "Wechat":
		sendWeChatNotification(text)
		return
	case "oneBot":
		err := oneBot.SendPrivateText(text)
		if err != nil {
			autoLog.Sugar.Error("通知-OneBot文本发送失败:", err)
		}
	default:
		autoLog.Sugar.Errorf("通知-文本未知通知类型:%s", config.Cfg.Notice.Type)
		return
	}
}

func SentImage(path string) error {

	var err error
	switch config.Cfg.Notice.Type {
	case "TG":
		err = sendTGImage(path)
		if err != nil {
			autoLog.Sugar.Error("通知-TG图片发送失败:", err)
		}
		return fmt.Errorf("通知-TG图片发送失败:%v", err)
	case "Wechat":
		err = sendWeChatImage(path)
		if err != nil {
			autoLog.Sugar.Error("通知-微信图片发送失败:", err)
		}
		return fmt.Errorf("通知-微信图片发送失败:%v", err)
	case "oneBot":
		err = oneBot.SendPrivateWithImage(path)
		if err != nil {
			autoLog.Sugar.Error("通知-OneBot图片发送失败:", err)
		}
		return fmt.Errorf("通知-OneBot图片发送失败:%v", err)
	default:
		autoLog.Sugar.Errorf("通知-图片未知通知类型:%s", config.Cfg.Notice.Type)
		return fmt.Errorf("通知-图片未知通知类型:%s", config.Cfg.Notice.Type)
	}
}

// 电脑截图
func SendScreenshot() error {

	err := control.ScreenShot()
	if err != nil {
		return fmt.Errorf("通知-图片发送失败:%v", err)
	}
	switch config.Cfg.Notice.Type {
	case "TG":
		err = sendTGImage("jt.png")
		if err != nil {
			autoLog.Sugar.Error("通知-TG图片发送失败:", err)
		}
		return fmt.Errorf("通知-TG图片发送失败:%v", err)
	case "Wechat":
		err = sendWeChatImage("jt.png")
		if err != nil {
			autoLog.Sugar.Error("通知-微信图片发送失败:", err)
		}
		return fmt.Errorf("通知-微信图片发送失败:%v", err)
	case "oneBot":
		err = oneBot.SendPrivateWithImage("jt.png")
		if err != nil {
			autoLog.Sugar.Error("通知-OneBot图片发送失败:", err)
		}
		return fmt.Errorf("通知-OneBot图片发送失败:%v", err)
	default:
		autoLog.Sugar.Errorf("通知-图片未知通知类型:%s", config.Cfg.Notice.Type)
		return fmt.Errorf("通知-图片未知通知类型:%s", config.Cfg.Notice.Type)
	}

}
