package Mihoyo

import (
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"errors"
	"fmt"
	"gopkg.in/ini.v1"
	"strconv"
)

// 一键配置米游社通知
func pushServer(NoticeType string) (string, error) {

	switch NoticeType {
	case "TG":
		res := make(map[string]interface{})
		res["push_server"] = "telegram"
		res["bot_token"] = config.Cfg.Notice.TGNotice.TGToken
		res["chat_id"] = config.Cfg.Notice.TGNotice.ChatID
		res["http_proxy"] = config.Cfg.Notice.TGNotice.Proxy
		return update(res)

	case "Wechat":
		res := make(map[string]interface{})
		res["push_server"] = "wecomrobot"
		res["url"] = config.Cfg.Notice.Wechat
		return update(res)
	case "oneBot":
		res := make(map[string]interface{})
		res["push_server"] = "cqhttp"
		res["cqhttp_url"] = config.Cfg.Notice.OneBot.APIBase
		res["cqhttp_qq"] = config.Cfg.Notice.OneBot.QQNum
		res["cqhttp_group"] = config.Cfg.Notice.OneBot.GroupNum
		return update(res)
	case "FeiShu":
		res := make(map[string]interface{})
		res["push_server"] = "feishubot"
		res["webhook"] = config.Cfg.Notice.FeiShu.FeiShuWebhookURL
		return update(res)
	default:
		return "未知的推送方式", errors.New("未知的推送方式")
	}
}

func update(res map[string]interface{}) (string, error) {

	cfg, err := ini.Load("MihoyoBBSTools/config/push.ini")
	if err != nil {
		fmt.Printf("无法读取文件: %v\n", err)
		autoLog.Sugar.Errorf("无法读取文件: %v\n", err)
		return "配置文件读取失败", err
	}

	cfg.Section("setting").Key("push_server").SetValue(res["push_server"].(string))

	for key, value := range res {

		if key == "push_server" {
			continue
		}
		//判断value是否为字符串
		if _, ok := value.(string); ok {
			cfg.Section(res["push_server"].(string)).Key(key).SetValue(value.(string))
		} else {
			cfg.Section(res["push_server"].(string)).Key(key).SetValue(strconv.FormatInt(value.(int64), 10))
		}
	}

	err = cfg.SaveTo("MihoyoBBSTools/config/push.ini")
	if err != nil {
		fmt.Printf("无法保存配置文件: %v\n", err)
		autoLog.Sugar.Errorf("无法保存配置文件: %v\n", err)
		return "配置文件保存失败", err
	}
	return "配置文件保存成功", nil
}
