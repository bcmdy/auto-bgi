package menu

import (
	"auto-bgi/BetterGI"
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"auto-bgi/control"
	"auto-bgi/tools"
	"github.com/gin-gonic/gin"
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
		return
	}
	c.JSON(888, gin.H{"msg": "更新成功！请重新启动abgi，5秒后自动退出"})

	time.Sleep(3 * time.Second)

	version, _ := GetVersion()

	control.GetMachineFingerprint("autobgi", version)

	go func() {
		// 调用 run_auto_bgi.vbs 脚本来启动新的 auto-bgi.exe 程序
		if err := tools.RestartProgram(); err != nil {
			autoLog.Sugar.Error(err.Error())
			//return fmt.Errorf("重启程序失败: %v", err)
		}
	}()

}

// GetBgiVersion 获取当前bgi版本和最新的bgi版本
func GetBgiVersion(c *gin.Context) {
	if !strings.Contains(config.BgiCfg.RunForVersion, "lcb") {

		versionInfo := BetterGI.GetYDVersionInfo()

		c.JSON(200, gin.H{"currentVersion": config.BgiCfg.RunForVersion, "lastVersion": versionInfo.Version, "msg": "获取版本成功"})
		return
	}
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
