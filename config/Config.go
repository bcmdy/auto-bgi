package config

import (
	"auto-bgi/autoLog"
	"encoding/json"
	"github.com/robfig/cron/v3"
	"io/ioutil"
	"os"
	"time"
)

type config struct {
	OneLongHour     int      `json:"OneLongHour"`
	OneLongMinute   int      `json:"OneLongMinute"`
	BetterGIAddress string   `json:"BetterGIAddress"`
	WebhookURL      string   `json:"webhookURL"`
	Content         string   `json:"content"`
	ConfigNames     []string `json:"ConfigNames"`
	BagStatistics   string   `json:"BagStatistics"`
	LongX           int      `json:"longX"`
	LongY           int      `json:"longY"`
	ExecuteX        int      `json:"executeX"`
	ExecuteY        int      `json:"executeY"`
	Post            string   `json:"post"`
	IsStartTimeLong bool     `json:"isStartTimeLong"`
	IsMysSignIn     bool     `json:"isMysSignIn"`
	Backups         []string `json:"backups"`
	Cookie          string   `json:"cookie"`
}

var Cfg config
var Parser = cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)

func init() {
	file, err := os.Open("main.json")
	if err != nil {
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
