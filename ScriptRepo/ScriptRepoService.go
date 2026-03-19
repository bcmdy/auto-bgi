package ScriptRepo

import (
	"auto-bgi/abgiConstant"
	"github.com/gin-gonic/gin"
	"os"
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

// 删除js脚本
func DeleteScript(context *gin.Context) {

	type Data struct {
		Name string `json:"name"`
	}
	var data Data
	err := context.ShouldBindJSON(&data)
	if err != nil {
		context.JSON(200, gin.H{
			"code": 500,
			"msg":  err.Error(),
		})
		return
	}
	jsPath := abgiConstant.JsPath(data.Name)
	if err := os.RemoveAll(jsPath); err != nil {
		context.JSON(200, gin.H{
			"code": 500,
			"msg":  err.Error(),
		})
		return

	}
	context.JSON(200, gin.H{
		"code": 200,
		"msg":  "success",
	})

}
