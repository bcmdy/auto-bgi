package bgiStatus

import (
	"auto-bgi/auth"
	"auto-bgi/config"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"path/filepath"
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
	fileName := context.Query("fileName")
	filePath := filepath.Clean(fmt.Sprintf("%s\\log\\%s", config.Cfg.BetterGIAddress, fileName)) // 本地日志路径
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

	//// 统一使用UTC时间
	//fmt.Println(time.Now().Format("2006-01-02 15:04:05"))
	//fmt.Println(RunDetail.StartTime.Format("2006-01-02 15:04:05"))
	//
	//hour := time.Now().Hour() - RunDetail.StartTime.Hour()
	//minute := time.Now().Minute() - RunDetail.StartTime.Minute()
	//fmt.Println(hour, minute)

	data["ExpectedToEnd"] = info.Timestamp
	data["line"] = info.MapTrackingLine
	data["scriptName"] = info.ScriptName
	data["progress"] = info.ConfigurationGroupExecutionProgress
	data["running"] = info.Running
	data["jsProgress"] = info.JSProgress
	data["title"] = auth.User.SysConfig.SystemName
	//data["version"] = Version
	//data["isUpdate"] = Version != ABgi.GetCurrentVersion()

	context.JSON(http.StatusOK, data)
}
