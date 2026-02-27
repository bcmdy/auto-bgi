package menu

import (
	"auto-bgi/abgiSSE"
	"auto-bgi/config"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func GetAppInfo(context *gin.Context) {

	data := make(map[string]any)
	//上线次数
	data["上线次数"] = abgiSSE.RunCount
	//查询当前aBgi版本
	//去除空格，\r\n
	Version := strings.TrimSpace(CurrentVersion)
	data["aBgi版本"] = Version
	//最新的aBgi版本
	data["最新aBgi版本"], _ = GetVersion()

	if !strings.Contains(config.BgiCfg.RunForVersion, "lcb") {
		//当前bgi版本
		data["bgi版本"] = "不能在线更新"
		//获取最新的bgi版本
		data["最新bgi版本"] = "不能在线更新"
	} else {
		//当前bgi版本
		data["bgi版本"] = config.BgiCfg.RunForVersion
		//获取最新的bgi版本
		data["最新bgi版本"], _ = GetVersion()
	}

	context.JSON(http.StatusOK, gin.H{"status": "success", "data": data})

}
