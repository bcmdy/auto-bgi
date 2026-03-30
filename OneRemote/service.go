package OneRemote

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 获取所有多用户
func GetMultiUsers(context *gin.Context) {

	sessions, err := GetDetailedSessions()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"status": "success", "data": sessions})
}

// 启动1Remote用户
func StartMultiUser(context *gin.Context) {

	value := context.Query("launcher")
	if value == "" {
		value = context.Query("username")
	}
	if value == "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "launcher is empty"})
		return
	}

	if err := startOneRemoteUser(value); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"message": "success"})
}

// 注销1Remote用户
func LogoutMultiUser(context *gin.Context) {
	sessionID := context.Query("id")
	if sessionID == "" {
		sessionID = context.Query("sessionID")
	}
	if sessionID == "" {
		username := context.Query("username")
		if username == "" {
			context.JSON(http.StatusBadRequest, gin.H{"error": "id/sessionID/username is empty"})
			return
		}
		id, err := FindSessionIDByUsername(username)
		if err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		sessionID = id
	}

	if err := ForceLogoff(sessionID); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"message": "success"})
}
