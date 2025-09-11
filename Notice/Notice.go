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

	autoLog.Sugar.Error("通知-文本未知通知类型:%s", config.Cfg.Notice.Type)
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
	autoLog.Sugar.Error("通知-图片未知通知类型:%s", config.Cfg.Notice.Type)

	return fmt.Errorf("通知-图片未知通知类型")
}

// 电脑截图
func SendScreenshot() error {

	err := control.ScreenShot()
	if err != nil {
		return fmt.Errorf("通知-图片发送失败:%v", err)
	}
	if config.Cfg.Notice.Type == "TG" {
		err := sendTGImage("jt.png")
		if err != nil {
			autoLog.Sugar.Error("通知-TG图片发送失败:", err)
		}
		return fmt.Errorf("通知-TG图片发送失败:%v", err)
	} else if config.Cfg.Notice.Type == "Wechat" {
		err := sendWeChatImage("jt.png")
		if err != nil {
			autoLog.Sugar.Error("通知-微信图片发送失败:", err)
		}
		return fmt.Errorf("通知-微信图片发送失败:%v", err)
	}
	autoLog.Sugar.Error("通知-图片未知通知类型:%s", config.Cfg.Notice.Type)

	return fmt.Errorf("通知-图片未知通知类型")
}
