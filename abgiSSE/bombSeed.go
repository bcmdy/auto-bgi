package abgiSSE

import (
	"auto-bgi/abgiConstant"
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"fmt"
	"github.com/gin-gonic/gin"
)

type BombRecord struct {
	//炸弹人名字
	BombName string `json:"BombName"`
	//炸弹人人行为
	BombAction string `json:"BombAction"`
	//举报人uid和名字
	ReportUidNickName string `json:"ReportUidNickName"`
}

// AddBombRecord 新增举报炸弹
func AddBombRecord(bombRecord BombRecord) {
	// 设置炸弹记录的报告用户ID和昵称
	// 格式为 "UID-昵称"
	bombRecord.ReportUidNickName = config.Cfg.Account.Uid + "-" + config.Cfg.Account.Name
	autoLog.Sugar.Infof("炸弹人名字: %s,行为:%s", bombRecord.BombName, bombRecord.BombAction)

}

func ReportBomb(context *gin.Context) {
	var bombRecord BombRecord
	if err := context.ShouldBindJSON(&bombRecord); err != nil {
		context.JSON(400, gin.H{"error": err.Error()})
		return
	}
	// 设置炸弹记录的报告用户ID和昵称
	// 格式为 "UID-昵称"
	bombRecord.ReportUidNickName = config.Cfg.Account.Uid + "-" + config.Cfg.Account.Name
	autoLog.Sugar.Infof("炸弹人名字: %s,行为:%s", bombRecord.BombName, bombRecord.BombAction)

	jsonResp, jsonStatus, jsonErr := abgiConstant.PostJSON(abgiConstant.ABgiInfoUrl+"/api/bombRecord/Add", bombRecord, nil)
	if jsonErr != nil {
		fmt.Printf("JSON 请求失败: %v\n", jsonErr)
		context.JSON(500, gin.H{"message": "举报失败"})
	} else {
		fmt.Printf("JSON 请求状态码: %d\n", jsonStatus)
		fmt.Printf("JSON 响应内容: %s\n\n", jsonResp)
		context.JSON(200, gin.H{"message": string(jsonResp)})
	}

}
