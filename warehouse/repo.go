package warehouse

import (
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"github.com/gin-gonic/gin"
	otiai10Copy "github.com/otiai10/copy"
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

func SubscribeScript(context *gin.Context) {
	ScriptName := context.Query("ScriptName")
	ReposScriptPath := filepath.Join(config.Cfg.BetterGIAddress, "Repos", "bettergi-scripts-list-git", "repo", "js", ScriptName)

	//复制user
	UserScriptPath := filepath.Join(config.Cfg.BetterGIAddress, "User", "JsScript", ScriptName)

	err := otiai10Copy.Copy(ReposScriptPath, UserScriptPath)
	if err != nil {
		autoLog.Sugar.Errorf("复制脚本失败: %v", err)
		context.JSON(500, gin.H{"error": "脚本不存在"})
		return
	}
	context.JSON(200, gin.H{"message": "脚本订阅成功"})

}
