package config

import (
	"auto-bgi/autoLog"
	"encoding/json"
	"github.com/robfig/cron/v3"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

type config struct {
	OneLong         oneLong  `json:"OneLong"`
	BetterGIAddress string   `json:"BetterGIAddress" comment:"BetterGI地址"`
	WebhookURL      string   `json:"webhookURL" comment:"webhook地址"`
	Content         string   `json:"content" comment:"通知内容"`
	ConfigNames     []string `json:"ConfigNames" comment:"一条龙配置名称"`
	BagStatistics   string   `json:"BagStatistics" comment:"需要统计的物品"`
	Post            string   `json:"post" comment:"post地址"`
	MySign          MySign   `json:"MySign"`
	Backups         []string `json:"backups"`
	Cookie          string   `json:"cookie"`
	BasePath        string   `json:"basePath"`
	JsName          []string `json:"jsName" comment:"需要更新的js名称"`
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

var Cfg config
var Parser = cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)

func init() {
	file, err := os.Open("main.json")
	if err != nil {
		autoLog.Sugar.Fatalf("打开配置文件失败: %v", err)
		return
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return
	}

	if err := json.Unmarshal(bytes, &Cfg); err != nil {
		return
	}
	// 获取程序的绝对路径
	ex, err := os.Executable()
	if err != nil {
		log.Fatalf("无法获取可执行文件路径: %v", err)
	}
	// 获取包含可执行文件的目录
	Cfg.BasePath = filepath.Dir(ex)
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
