package OneRemote

import "github.com/gin-gonic/gin"

// 获取所有多用户
func GetMultiUsers(context *gin.Context) {

	sessions, err := GetDetailedSessions()
	if err != nil {
		context.JSON(500, gin.H{"error": err.Error()})
		return
	}
	context.JSON(200, sessions)
}

// 启动1Remote用户
func StartMultiUser(context *gin.Context) {

	value := context.Query("username")
	if value == "" {
		context.JSON(500, gin.H{"error": "username is empty"})
		return
	}

	startOneRemoteUser(value)
	context.JSON(200, gin.H{"message": "success"})
}

// 注销1Remote用户
func LogoutMultiUser(context *gin.Context) {
	value := context.Query("username")
	if value == "" {
		context.JSON(500, gin.H{"error": "username is empty"})
		return
	}

	ForceLogoff(value)
	context.JSON(200, gin.H{"message": "success"})
}
