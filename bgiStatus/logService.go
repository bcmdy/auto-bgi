package bgiStatus

import (
	"auto-bgi/config"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"path/filepath"
	"time"
)

// 获取所有日志文件
func GetLogFiles(context *gin.Context) {
	filePath := filepath.Clean(fmt.Sprintf("%s\\log", config.Cfg.BetterGIAddress)) // 本地日志路径
	files, err := findLogFiles(filePath)
	if err != nil {
		return
	}
	context.JSON(http.StatusOK, gin.H{"files": files})
}

// 获取指定日志文件内容
func GetLogFileContent(context *gin.Context) {
	filePath := filepath.Clean(fmt.Sprintf("%s\\log\\better-genshin-impact%s.log", config.Cfg.BetterGIAddress, time.Now().Format("20060102"))) // 本地日志路径
	logInfo, err := GetLogInfo(filePath)
	if err != nil {
		context.String(http.StatusInternalServerError, "读取日志失败")
	}

	context.String(http.StatusOK, logInfo)
}

// 首页内容
func GetIndex(context *gin.Context) {
	info := BgiLogStatusInfo
	data := make(map[string]interface{})
	data["group"] = info.Group + " [" + info.GroupProgress + "]"
	data["ExpectedToEnd"] = info.Timestamp
	data["line"] = info.MapTrackingLine
	data["scriptName"] = info.ScriptName
	data["progress"] = info.ConfigurationGroupExecutionProgress
	data["running"] = info.Running
	data["jsProgress"] = info.JSProgress
	//data["version"] = Version
	//data["isUpdate"] = Version != ABgi.GetCurrentVersion()

	context.JSON(http.StatusOK, data)
}
