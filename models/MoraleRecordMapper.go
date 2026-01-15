package models

import (
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"time"
)

// 新增摩拉记录
func AddMoraleRecord(time string, num int, action string) {
	var moraleRecord MoraleRecord
	moraleRecord.Time = time
	moraleRecord.Num = num
	moraleRecord.Action = action
	//根据时间判断有没有新增过
	if err := DB.Where("time = ?", time).First(&moraleRecord).Error; err != nil {
		DB.Create(&moraleRecord)
	}
}

//// 批量新增摩拉记录（根据 time 去重）
//func BatchAddMoraleRecords(records []MoraleRecord) error {
//	autoLog.Sugar.Infof("共%d条新记录", len(records))
//	if len(records) == 0 {
//		return nil
//	}
//	return DB.
//		Clauses(clause.OnConflict{
//			Columns:   []clause.Column{{Name: "time"}}, // 冲突字段
//			DoNothing: true,                            // 已存在则忽略
//		}).
//		Create(&records).Error
//}

// 批量新增摩拉记录（不进行任何去重检查）
func BatchAddMoraleRecords(records []MoraleRecord) error {
	autoLog.Sugar.Infof("共%d条新记录", len(records))
	if len(records) == 0 {
		return nil
	}
	return DB.Create(&records).Error
}

// 查询出最后一条摩拉记录的时间
func GetLastMoraleRecord() string {
	var moraleRecord MoraleRecord
	DB.Where("uid = ?", config.GameRoles.Data.List[0].GameId).Order("time desc").First(&moraleRecord)
	if moraleRecord.Time == "" {
		yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
		return yesterday + " 00:00:00"
	}
	return moraleRecord.Time
}

// 今天是否已经发送过摩拉排行榜:
func IsSendMoraleRank() bool {
	//判断ck有没有填
	if config.BgiCfg.MiYouSheConfigCookie == "" {
		autoLog.Sugar.Errorf("ck没有填")
		return false
	}
	if len(config.GameRoles.Data.List) == 0 {
		autoLog.Sugar.Errorf("没有绑定角色")
		return false
	}
	today := time.Now().Format("2006-01-02")
	value, err := GetAutoBgiValue("MoraleRank")
	if err != nil {
		autoLog.Sugar.Errorf("MoraleRank查询失败: %v", err)
		return true
	}
	//判断今天是否已经发送
	if value == today {
		return false
	} else {
		return true
	}
}
