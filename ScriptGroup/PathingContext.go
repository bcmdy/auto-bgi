package ScriptGroup

import (
	"auto-bgi/config"
	"github.com/gin-gonic/gin"
	"net/http"
)

// 更新地图追踪
func (s *ScriptGroupConfig) UpdatePathingNoConfig(context *gin.Context) {
	var updatePath config.UpdatePathing
	if err := context.ShouldBindJSON(&updatePath); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "参数格式错误", "error": err.Error()})
		return
	}

	res, err := s.updatePathingNoConfig(updatePath.FolderName)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"status": "error", "msg": err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{"status": "success", "data": res})

}
