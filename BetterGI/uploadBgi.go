package BetterGI

import (
	"auto-bgi/autoLog"
	"auto-bgi/bgiStatus"
	"auto-bgi/config"
	"auto-bgi/control"
	"auto-bgi/tools"
	"github.com/gin-gonic/gin"
	abgiCopy "github.com/otiai10/copy"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// 上传bgi压缩包
func UploadBgi(c *gin.Context) {

	control.CloseSoftware()

	// 单文件上传
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 保存文件到指定目录
	dst := filepath.Join("./uploads", "BetterGI.zip")
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	time.Sleep(1 * time.Second)

	UpdateBgi()

	//更新仓库
	go func() {
		bgiStatus.GitPull()
	}()

	c.JSON(http.StatusOK, gin.H{
		"message":  "更新成功",
		"filename": file.Filename,
		"size":     file.Size,
	})
}

// 更新bgi
func UpdateBgi() {

	now := time.Now().Format("2006-01-02-15-04-05")

	//1、备份user文件
	err4 := bgiStatus.ZipDir(config.Cfg.BetterGIAddress+"\\User\\", "Users\\User"+now+".zip", true)
	if err4 != nil {
		autoLog.Sugar.Errorf("备份失败: %v", err4)
		return
	}
	autoLog.Sugar.Infof("备份user文件")

	//备份log
	err2 := abgiCopy.Copy(config.Cfg.BetterGIAddress+"\\log\\", "backups\\log\\")
	if err2 != nil {
		autoLog.Sugar.Errorf("备份log失败: %v", err2)
		return
	}
	autoLog.Sugar.Infof("备份log")

	//2、删除bgi文件夹
	err := os.RemoveAll(config.Cfg.BetterGIAddress)
	if err != nil {
		autoLog.Sugar.Errorf("删除bgi文件夹失败: %v", err)
		return
	}
	autoLog.Sugar.Infof("删除bgi文件夹")

	//3、解压bgi压缩包
	base := filepath.Base(config.Cfg.BetterGIAddress)
	path := strings.ReplaceAll(config.Cfg.BetterGIAddress, base, "")
	err4 = tools.Un7z("./uploads/BetterGI.zip", path)
	if err4 != nil {
		autoLog.Sugar.Errorf("解压bgi压缩包失败: %v", err)
		return
	}

	autoLog.Sugar.Infof("解压bgi压缩包")

	//4、删除user文件夹
	err2 = os.RemoveAll(config.Cfg.BetterGIAddress + "\\User\\")
	if err2 != nil {
		autoLog.Sugar.Errorf("删除user文件夹失败: %v", err2)
		return
	}
	autoLog.Sugar.Infof("删除user文件夹")

	//5、复制压缩包到User
	err3 := abgiCopy.Copy("./Users/User"+now+".zip", config.Cfg.BetterGIAddress+"\\User.zip")
	if err3 != nil {
		autoLog.Sugar.Errorf("复制压缩包到User失败: %v", err3)
		return
	}
	autoLog.Sugar.Infof("复制压缩包到User")

	// 6、解压user压缩包
	err4 = tools.Unzip(config.Cfg.BetterGIAddress+"\\User.zip", config.Cfg.BetterGIAddress)
	if err4 != nil {
		autoLog.Sugar.Errorf("解压user压缩包失败: %v", err4)
		return
	}
	autoLog.Sugar.Infof("解压user压缩包")

	//7、删除user压缩包
	err5 := os.Remove(config.Cfg.BetterGIAddress + "\\User.zip")
	if err5 != nil {
		autoLog.Sugar.Errorf("删除user压缩包失败: %v", err5)
	}

	//8、还原log
	err5 = abgiCopy.Copy("backups\\log\\", config.Cfg.BetterGIAddress+"\\log\\")
	if err5 != nil {
		autoLog.Sugar.Errorf("还原log失败: %v", err5)
	}
	autoLog.Sugar.Infof("还原log")

	err5 = os.Remove("./uploads/BetterGI.zip")
	if err5 != nil {
		autoLog.Sugar.Errorf("删除BetterGI压缩包失败: %v", err5)
	}
	autoLog.Sugar.Infof("删除BetterGI压缩包")

	//删除log
	err5 = os.RemoveAll("backups\\log\\")
	if err5 != nil {
		autoLog.Sugar.Errorf("删除log失败: %v", err5)
	}
	autoLog.Sugar.Infof("删除log")

}
