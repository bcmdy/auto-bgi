package JsAPI

import (
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"auto-bgi/control"
	"auto-bgi/tools"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"time"
)

// QueryHistoryVersionService 查询脚本所有历史版本
func QueryHistoryVersionService(context *gin.Context) {

	jsName := context.Query("jsName")
	versions, err := QueryHistoryVersion(jsName)
	if err != nil {
		context.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"success": true, "versions": versions})
}

// RollbackHistoryVersionService 回滚脚本历史版本
func RollbackHistoryVersionService(context *gin.Context) {
	type RollbackVersionRequest struct {
		Version string `json:"version"`
		JSName  string `json:"jsName"`
	}
	var request RollbackVersionRequest
	if err := context.ShouldBindJSON(&request); err != nil {
		context.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	// 回滚脚本历史版本
	res, err := RollbackHistoryVersion(request.JSName, request.Version)
	if err != nil {
		context.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"success": true, "message": res})
}

// BgiRollbackHistoryVersionService bgi 版本回滚
func BgiRollbackHistoryVersionService(context *gin.Context) {

	type RollbackVersionRequest struct {
		Version string `json:"version"`
		JSName  string `json:"jsName"`
	}
	var request RollbackVersionRequest
	if err := context.ShouldBindJSON(&request); err != nil {
		context.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	// 回滚bgi历史版本
	err := BgiRollbackHistoryVersion(request.JSName, request.Version)
	if err != nil {
		context.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}

	Version := strings.ReplaceAll(request.Version, ".7z", "")

	config.BgiCfg.RunForVersion = Version
	control.OpenSoftware(config.Cfg.BetterGIAddress + "\\BetterGI.exe")

	//重启aBgi
	time.Sleep(10 * time.Second)
	autoLog.Sugar.Infof("准备重启abgi")
	if err := tools.RestartProgram(); err != nil {
		autoLog.Sugar.Errorf("重启程序失败: %v", err)
		autoLog.Sugar.Error(err.Error())
	}

	context.JSON(http.StatusOK, gin.H{"success": true, "message": "回滚成功"})
}
