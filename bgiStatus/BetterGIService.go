package bgiStatus

import (
	"auto-bgi/config"
	"github.com/gin-gonic/gin"
	"os"
	"strings"
)

func ReadBgiConfig(context *gin.Context) {
	configName := context.Query("configName")
	oneDragonStruct := readOneDragonConfig(configName + ".json")
	context.JSON(200, gin.H{"success": true, "msg": oneDragonStruct})
}

func ModifyBgiConfig(context *gin.Context) {
	var oneDragonStruct OneDragonStruct
	err := context.ShouldBindJSON(&oneDragonStruct)
	if err != nil {
		context.JSON(200, gin.H{"success": false, "data": err.Error()})
		return
	}
	ModifyOneDragonConfig(oneDragonStruct.Name+".json", oneDragonStruct)
	context.JSON(200, gin.H{"success": true, "msg": "修改成功"})

}

func ReadBgiConfigAll(context *gin.Context) {
	//读取所有配置
	entries, err := os.ReadDir(config.Cfg.BetterGIAddress + "\\User\\OneDragon")
	if err != nil {
		context.JSON(200, gin.H{"success": false, "msg": err.Error()})
		return
	}
	var oneLongInfo []string
	for _, entry := range entries {

		//去除后缀：.json
		name := strings.ReplaceAll(entry.Name(), ".json", "")

		oneLongInfo = append(oneLongInfo, name)

	}
	context.JSON(200, gin.H{"success": true, "msg": oneLongInfo})
	return

}
