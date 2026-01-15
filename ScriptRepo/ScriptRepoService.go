package ScriptRepo

import (
	"github.com/gin-gonic/gin"
)

// 查询所有js脚本
func QueryAllScript(context *gin.Context) {

	search := context.Query("search")
	script, err := getAllJsScript(search)
	if err != nil {
		context.JSON(200, gin.H{
			"code": 500,
			"msg":  err.Error(),
		})
		return
	}
	context.JSON(200, gin.H{
		"code": 200,
		"msg":  "success",
		"data": script,
	})

}
