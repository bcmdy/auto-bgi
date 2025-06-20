package config

import (
	"encoding/json"
	"github.com/robfig/cron/v3"
	"io/ioutil"
	"os"
)

type config struct {
	OneLongHour     int      `json:"OneLongHour"`
	OneLongMinute   int      `json:"OneLongMinute"`
	BetterGIAddress string   `json:"BetterGIAddress"`
	WebhookURL      string   `json:"webhookURL"`
	Content         string   `json:"content"`
	ConfigName      string   `json:"ConfigName"`
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
