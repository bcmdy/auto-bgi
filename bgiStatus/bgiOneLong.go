package bgiStatus

import (
	"time"
)

// 根据启动配置组计算一条龙时间
func GetTodayOneLongTime(oneLongGroup []string) {
	archive := ListArchive()
	var groupMap = make(map[string]string)
	for _, item := range archive {
		groupMap[item.Title] = item.ExecuteTime
	}

	var start = time.Now()

	for _, s := range oneLongGroup {

		executeTime := groupMap[s]
		if executeTime != "" {
			// 将执行时长字符串 "HH:MM:SS" 转为 Duration
			duration, err := time.ParseDuration(executeTime)
			if err != nil {
				continue
			}

			// 计算预计结束时间
			start = start.Add(duration)
		} else if s == "领取邮件" {
			// 领取邮件 10 分钟
			start = start.Add(4 * time.Minute)
		} else if s == "领取尘歌壶奖励" {
			// 领取尘歌壶奖励 10 分钟
			start = start.Add(2 * time.Minute)
		} else if s == "自动秘境" {
			start = start.Add(4 * time.Minute)
		}

	}
	SentText("一条龙预计结束时间:" + start.Format("2006-01-02 15:04:05"))
}
