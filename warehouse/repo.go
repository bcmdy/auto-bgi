package warehouse

import (
	"auto-bgi/config"
	"github.com/gin-gonic/gin"
	"os"
	"path/filepath"
)

//

func RepoReset(context *gin.Context) {
	reposPath := filepath.Join(config.Cfg.BetterGIAddress, "Repos", "bettergi-scripts-list-git")
	err := os.RemoveAll(reposPath)
	if err != nil {
		context.JSON(500, gin.H{"error": err.Error()})
		return
	}
	context.JSON(200, gin.H{"message": "Repos reset successfully"})
}
