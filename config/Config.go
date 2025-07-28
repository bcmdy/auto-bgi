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
	JsName          []string `json:"jsName" comment:"需要更新的js名称"`
	Control         Control  `json:"Control" comment:"控制配置"`
}

type Control struct {
	IsCloseYuanShen bool `json:"IsCloseYuanShen" comment:"bgi关闭需要是否关闭原神"`
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

	fmt.Println("配置文件加载成功")
	return nil
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
