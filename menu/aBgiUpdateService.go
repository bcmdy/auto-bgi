package menu

import (
	"auto-bgi/autoLog"
	"github.com/gin-gonic/gin"
	"os"
	"strings"
	"time"
)

func GetCurrentVersion(c *gin.Context) {

	//去除空格，\r\n
	Version := strings.TrimSpace(CurrentVersion)

	c.JSON(200, gin.H{"version": Version})

}

// 获取最新版本
func GetLastVersion(c *gin.Context) {
	version, err := GetVersion()
	if err != nil {
		autoLog.Sugar.Error(err.Error())
		c.JSON(200, gin.H{"version": CurrentVersion, "msg": "获取版本失败"})
	}
	c.JSON(200, gin.H{"version": version, "msg": "获取版本成功"})

}

// 更新abgi
func UpdateABgi(c *gin.Context) {
	err := Update()
	if err != nil {
		autoLog.Sugar.Error(err.Error())
		c.JSON(200, gin.H{"msg": "更新失败"})
	}
	c.JSON(200, gin.H{"msg": "更新成功！请重新启动abgi，5秒后自动退出"})

	go func() {
		//等待5秒后关闭程序
		<-time.After(5 * time.Second)
		//关闭程序
		os.Exit(0)
	}()
}
