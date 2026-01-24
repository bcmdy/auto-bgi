package Mihoyo

import (
	"auto-bgi/autoLog"
	"github.com/gin-gonic/gin"
)

func UpdatePushConfig(context *gin.Context) {

	NoticeType := context.Query("NoticeType")
	server, err := pushServer(NoticeType)
	if err != nil {
		autoLog.Sugar.Infof("配置失败:%s", err)
		context.JSON(200, gin.H{
			"status":  500,
			"message": "配置失败",
		})
		return
	}
	context.JSON(200, gin.H{
		"status":  200,
		"message": server,
	})
	return
}
