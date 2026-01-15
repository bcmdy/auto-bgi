package task

import (
	"auto-bgi/abgiConstant"
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"auto-bgi/models"
	"fmt"
	"time"
)

// 发送摩拉排行榜:
func SendMoraleRank() {
	//捕获异常
	defer func() {
		if r := recover(); r != nil {
			autoLog.Sugar.Errorf("摩拉排行榜异常:%v", r)
			return
		}
	}()

	//判断时间是否在5点之后
	hour := time.Now().Hour()
	if hour <= 4 {
		autoLog.Sugar.Infof("不在发送摩拉排行榜的时间范围内")
		return
	}

	//今天是否已经发送
	if !models.IsSendMoraleRank() {
		autoLog.Sugar.Infof("今天已经发送摩拉排行榜")
		return
	}

	//获取摩拉排行榜
	var TotalMorale int
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	err := models.DB.Model(&models.MoraleRecord{}).
		Select("COALESCE(SUM(num), 0)").
		Where("time LIKE ? and uid = ?", yesterday+"%", config.GameRoles.Data.List[0].GameId).
		Row().Scan(&TotalMorale)
	if err != nil {
		autoLog.Sugar.Errorf("获取摩拉排行榜失败:%v", err)
		return
	}
	autoLog.Sugar.Infof("昨天摩拉排行榜总收益:%v", TotalMorale)
	if TotalMorale == 0 {
		return
	}
	//发送摩拉排行榜
	EarningsRecord := make(map[string]interface{})
	EarningsRecord["uid"] = config.GameRoles.Data.List[0].GameId
	EarningsRecord["name"] = config.GameRoles.Data.List[0].NicName
	EarningsRecord["revenue"] = TotalMorale

	jsonResp, jsonStatus, jsonErr := abgiConstant.PostJSON(abgiConstant.ABgiInfoUrl+"/api/MoraRank/Add", EarningsRecord, nil)
	if jsonErr != nil {
		fmt.Printf("JSON 请求失败: %v\n", jsonErr)
	} else {
		fmt.Printf("JSON 请求状态码: %d\n", jsonStatus)
		fmt.Printf("JSON 响应内容: %s\n\n", jsonResp)
		//更新发送摩拉排行榜状态
		models.UpdateAutoBgiValue("MoraleRank", time.Now().Format("2006-01-02"))

	}
}
