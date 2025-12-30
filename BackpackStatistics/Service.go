package BackpackStatistics

import "github.com/gin-gonic/gin"

func ListService(context *gin.Context) {

}
func AddService(context *gin.Context) {
	value := context.Query("name")
	err := Add(value)
	if err != nil {
		context.JSON(200, gin.H{
			"code": 400,
			"msg":  err.Error(),
		})
	}
	context.JSON(200, gin.H{
		"code": 200,
		"msg":  "ok",
	})

}
func DeleteService(context *gin.Context) {

	value := context.Query("name")
	err := Delete(value)
	if err != nil {
		context.JSON(200, gin.H{
			"code": 400,
			"msg":  err.Error(),
		})
	}
	context.JSON(200, gin.H{
		"code": 200,
		"msg":  "ok",
	})

}

func ClearAllService(context *gin.Context) {
	err := ClearAll()
	if err != nil {
		context.JSON(200, gin.H{
			"code": 400,
			"msg":  err.Error(),
		})
	}
	context.JSON(200, gin.H{
		"code": 200,
		"msg":  "ok",
	})
}
