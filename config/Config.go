package config

import (
	"auto-bgi/autoLog"
	"encoding/json"
	"fmt"
	"github.com/robfig/cron/v3"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	OneLong         oneLong  `json:"OneLong" comment:"一条龙配置"`
	BetterGIAddress string   `json:"BetterGIAddress" comment:"BetterGI地址"`
	WebhookURL      string   `json:"webhookURL" comment:"webhook地址"`
	Content         string   `json:"content" comment:"通知内容"`
	ConfigNames     []string `json:"ConfigNames" comment:"一条龙配置名称"`
	BagStatistics   string   `json:"BagStatistics" comment:"需要统计的物品"`
	Post            string   `json:"post" comment:"post地址"`
	MySign          MySign   `json:"MySign" comment:"米游社签到"`
	Backups         []string `json:"backups" comment:"需要的备份文件"`
	Cookie          string   `json:"cookie"`
	BasePath        string   `json:"basePath"`
	Control         Control  `json:"Control" comment:"控制配置"`
	LogKeywords     []string `json:"LogKeywords" comment:"日志关键词"`
}

type Control struct {
	IsCloseYuanShen bool `json:"IsCloseYuanShen" comment:"bgi关闭需要是否关闭原神"`
	SendWeChatImage bool `json:"SendWeChatImage" comment:"是否开启每隔一小时发送截图"`
}

type oneLong struct {
	IsStartTimeLong bool `json:"isStartTimeLong" comment:"是否开启一条龙"`
	OneLongHour     int  `json:"OneLongHour" comment:"一条龙小时"`
	OneLongMinute   int  `json:"OneLongMinute" comment:"一条龙分钟"`
}

type MySign struct {
	IsMySignIn bool   `json:"isMysSignIn" comment:"是否开启我的签到"`
	Url        string `json:"url" comment:"我的签到url"`
}

var Cfg Config
var Parser = cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)

func init() {
	err := ReloadConfig()
	if err != nil {
		//autoLog.Sugar.Fatalf("首次加载配置失败: %v", err)
		fmt.Println("首次加载配置失败: %v", err)
	}
}

// ReloadConfig 重新加载配置文件
func ReloadConfig() error {
	file, err := os.Open("main.json")
	if err != nil {
		fmt.Println("ReloadConfig打开配置文件失败: %v", err)
		return err
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {

		fmt.Println("ReloadConfig读取配置文件失败: %v", err)
		return err
	}

	if err := json.Unmarshal(bytes, &Cfg); err != nil {

		fmt.Println("ReloadConfig解析配置文件失败: %v", err)
		return err
	}

	// 更新 BasePath 为当前可执行文件目录
	ex, err := os.Executable()
	if err != nil {
		log.Printf("无法获取可执行文件路径: %v", err)
		// 不返回错误，继续执行
	} else {
		Cfg.BasePath = filepath.Dir(ex)
	}

	// 💡 填充默认值
	fillDefaults(&Cfg)

	// 💾 写回 main.json（已填充默认值）
	err = writeConfig("main.json", &Cfg)
	if err != nil {
		fmt.Println("写回配置文件失败:", err)
	} else {
		fmt.Println("配置文件加载并更新默认值成功")
	}

	return nil
}

func writeConfig(filename string, cfg *Config) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // 美化格式
	return encoder.Encode(cfg)
}

func fillDefaults(cfg *Config) {
	if cfg.BetterGIAddress == "" {
		cfg.BetterGIAddress = "D:\\BetterGI"
	}
	if cfg.WebhookURL == "" {
		cfg.WebhookURL = "https://qyapi.weixin.qq.com"
	}
	if cfg.Content == "" {
		cfg.Content = "可以填autobgi的网页链接"
	}
	if cfg.ConfigNames == nil || len(cfg.ConfigNames) < 7 {
		cfg.ConfigNames = []string{"星期日", "星期一", "星期二", "星期三", "星期四", "星期五", "星期六"}
	}
	if cfg.BagStatistics == "" {
		cfg.BagStatistics = "晶核,大英雄的经验,水晶块,竹笋,螃蟹,劫波莲,兽肉,萃凝晶,紫晶块,星银矿石"
	}
	if cfg.Post == "" {
		cfg.Post = ":8082"
	}
	if cfg.Cookie == "" {
		cfg.Cookie = ""
	}
	if cfg.Backups == nil {
		cfg.Backups = []string{}
	}
	if cfg.LogKeywords == nil {
		cfg.LogKeywords = []string{
			"未识别到突发任务",
			"OCR 识别失败",
			"此路线出现3次卡死，重试一次路线或放弃此路线！",
			"检测到复苏界面，存在角色被击败",
			"执行路径时出错",
			"传送点未激活或不存在"}
	}
	if !cfg.OneLong.IsStartTimeLong {
		cfg.OneLong.IsStartTimeLong = false
	}
	if cfg.OneLong.OneLongHour == 0 {
		cfg.OneLong.OneLongHour = 4
	}
	if cfg.OneLong.OneLongMinute == 0 {
		cfg.OneLong.OneLongMinute = 10
	}
	if !cfg.MySign.IsMySignIn {
		cfg.MySign.IsMySignIn = false
	}
	if cfg.MySign.Url == "" {
		cfg.MySign.Url = "http://localhost:8883"
	}
	if cfg.Control.IsCloseYuanShen {
		cfg.Control.IsCloseYuanShen = false
	}
	if cfg.Control.SendWeChatImage {
		cfg.Control.SendWeChatImage = false
	}

}

// 获取今天启动的一条龙名字
func GetTodayOneLongName() string {
	var oneLongs = Cfg.ConfigNames
	now := time.Now()
	weekdayNum := int(now.Weekday())
	autoLog.Sugar.Infof("今天是: 星期%d", weekdayNum)
	oneLongName := oneLongs[weekdayNum]
	return oneLongName
}
