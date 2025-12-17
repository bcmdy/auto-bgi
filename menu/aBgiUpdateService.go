package menu

import (
	"auto-bgi/BetterGI"
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"github.com/gin-gonic/gin"
	"strings"
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
		return
	}
	c.JSON(200, gin.H{"msg": "更新成功！请重新启动abgi，5秒后自动退出"})

}

// GetBgiVersion 获取当前bgi版本和最新的bgi版本
func GetBgiVersion(c *gin.Context) {
	//获取当前bgi版本
	LastBgiVersion, err := BetterGI.GetVersion()
	if err != nil {
		autoLog.Sugar.Error(err.Error())
		c.JSON(200, gin.H{"msg": "获取版本失败"})
		return
	}
	autoLog.Sugar.Infof("最新的bgi版本: %s", LastBgiVersion)
	c.JSON(200, gin.H{"currentVersion": config.BgiCfg.RunForVersion, "lastVersion": LastBgiVersion, "msg": "获取版本成功"})

}
