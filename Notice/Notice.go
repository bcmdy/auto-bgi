package Notice

import (
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"auto-bgi/control"
	"fmt"
	"sync"
	"time"
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

//func SentText(text string) {
//
//	switch config.Cfg.Notice.Type {
//	case "TG":
//		err := sendTGNotification(text)
//		if err != nil {
//			autoLog.Sugar.Error("通知-TG文本发送失败:", err)
//		}
//		return
//	case "Wechat":
//		sendWeChatNotification(text)
//		return
//	case "oneBot":
//		err := oneBot.SendPrivateText(text)
//		if err != nil {
//			autoLog.Sugar.Error("通知-OneBot文本发送失败:", err)
//		}
//	default:
//		autoLog.Sugar.Errorf("通知-文本未知通知类型:%s", config.Cfg.Notice.Type)
//		return
//	}
//}

var (
	lastSentMap = struct {
		sync.Mutex
		m map[string]time.Time
	}{m: make(map[string]time.Time)}

	noticeTTL   = 20 * time.Second
	maxCacheLen = 5
)

func SentText(text string) {
	lastSentMap.Lock()
	defer lastSentMap.Unlock()

	// 如果存在并且 20 秒内，直接忽略
	if t, ok := lastSentMap.m[text]; ok {
		if time.Since(t) < noticeTTL {
			return
		}
	}

	// 如果超过最大缓存大小，移除最旧的一条
	if len(lastSentMap.m) >= maxCacheLen {
		var oldestKey string
		var oldestTime time.Time
		for k, v := range lastSentMap.m {
			if oldestTime.IsZero() || v.Before(oldestTime) {
				oldestKey = k
				oldestTime = v
			}
		}
		delete(lastSentMap.m, oldestKey)
	}

	// 更新为当前时间
	lastSentMap.m[text] = time.Now()

	// 执行真正的发送逻辑
	switch config.Cfg.Notice.Type {
	case "TG":
		if err := sendTGNotification(text); err != nil {
			autoLog.Sugar.Error("通知-TG文本发送失败:", err)
		}
	case "Wechat":
		sendWeChatNotification(text)
	case "oneBot":
		if err := oneBot.SendPrivateText(text); err != nil {
			autoLog.Sugar.Error("通知-OneBot文本发送失败:", err)
		}
	case "FeiShu":
		err := sendFeishuTextMessage(text)
		if err != nil {
			autoLog.Sugar.Error("通知-飞书文本发送失败:", err)
		}

	default:
		autoLog.Sugar.Errorf("通知-文本未知通知类型:%s", config.Cfg.Notice.Type)
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
			return fmt.Errorf("通知-微信图片发送失败:%v", err)
		}
		autoLog.Sugar.Info("通知-微信图片发送成功")
		return nil
	case "oneBot":
		err = oneBot.SendPrivateWithImage(path)
		if err != nil {
			autoLog.Sugar.Error("通知-OneBot图片发送失败:", err)
			return fmt.Errorf("通知-OneBot图片发送失败:%v", err)
		}
		autoLog.Sugar.Info("通知-OneBot图片发送成功")
		return nil

	case "FeiShu":
		err = sendFeiShuImageMessage(path)
		if err != nil {
			autoLog.Sugar.Error("通知-飞书图片发送失败:", err)
			return fmt.Errorf("通知-飞书图片发送失败:%v", err)
		}
		autoLog.Sugar.Info("通知-飞书图片发送成功")
		return nil

	default:
		autoLog.Sugar.Errorf("通知-图片未知通知类型:%s", config.Cfg.Notice.Type)
		return fmt.Errorf("通知-图片未知通知类型:%s", config.Cfg.Notice.Type)
	}
}

// 电脑截图
func SendScreenshot() error {

	err := control.ScreenShot("jt.jpg")
	if err != nil {
		return fmt.Errorf("通知-图片发送失败:%v", err)
	}
	switch config.Cfg.Notice.Type {
	case "TG":
		err = sendTGImage("jt.jpg")
		if err != nil {
			autoLog.Sugar.Error("通知-TG图片发送失败:", err)
		}
		return fmt.Errorf("通知-TG图片发送失败:%v", err)
	case "Wechat":
		err = sendWeChatImage("jt.jpg")
		if err != nil {
			autoLog.Sugar.Error("通知-微信图片发送失败:", err)
		}
		return fmt.Errorf("通知-微信图片发送失败:%v", err)
	case "oneBot":
		err = oneBot.SendPrivateWithImage("jt.jpg")
		if err != nil {
			autoLog.Sugar.Error("通知-OneBot图片发送失败:", err)
		}
		return fmt.Errorf("通知-OneBot图片发送失败:%v", err)
	default:
		autoLog.Sugar.Errorf("通知-图片未知通知类型:%s", config.Cfg.Notice.Type)
		return fmt.Errorf("通知-图片未知通知类型:%s", config.Cfg.Notice.Type)
	}

}
