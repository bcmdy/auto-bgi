package JsAPI

import (
	"auto-bgi/ScriptRepo"
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"auto-bgi/tools"
	"fmt"
	abgiCopy "github.com/otiai10/copy"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// 备份历史版本
func BackupHistoryVersion(jsPath string, version string) string {
	//捕获异常
	defer func() {
		if err := recover(); err != nil {
			autoLog.Sugar.Errorf("备份历史版本失败: %v", err)
		}
	}()

	jsName := filepath.Base(jsPath)

	err := abgiCopy.Copy(jsPath, "./historyVersion/"+jsName+"-version-"+version)
	if err != nil {
		return ""
	}

	// 清理旧版本，仅保留最近5个 ---
	files, err := os.ReadDir("./historyVersion/")
	if err == nil {
		type fileInfo struct {
			name string
			time time.Time
		}
		var historyFiles []fileInfo

		// 1. 筛选出属于当前 jsName 的备份文件
		for _, f := range files {
			if !f.IsDir() && strings.HasPrefix(f.Name(), jsName+"-version-") {
				info, err := f.Info()
				if err == nil {
					historyFiles = append(historyFiles, fileInfo{f.Name(), info.ModTime()})
				}
			}
		}

		// 2. 按修改时间从新到旧排序
		sort.Slice(historyFiles, func(i, j int) bool {
			return historyFiles[i].time.After(historyFiles[j].time)
		})

		// 3. 如果超过5个，删除最早的
		if len(historyFiles) > 5 {
			for i := 5; i < len(historyFiles); i++ {
				oldFile := filepath.Join("./historyVersion/", historyFiles[i].name)
				os.Remove(oldFile)
				autoLog.Sugar.Infof("已清理过期的历史版本: %s", historyFiles[i].name)
			}
		}
	}
	// ---------------------------------------

	autoLog.Sugar.Infof("历史版本备份成功：%s-%s", jsName, version)
	return "历史版本备份成功"

}

// 根据脚本名称查询历史版本
func QueryHistoryVersion(jsName string) ([]string, error) {

	entries, err := os.ReadDir("./historyVersion")
	if err != nil {
		fmt.Printf("读取目录失败: %v\n", err)
		return nil, err
	}

	var matchedFiles []string
	for _, entry := range entries {

		// 确保是文件而不是目录，且匹配前缀
		if strings.HasPrefix(entry.Name(), jsName) {
			version := strings.ReplaceAll(entry.Name(), jsName+"-version-", "")
			matchedFiles = append(matchedFiles, version)
		}
	}

	return matchedFiles, nil
}

// 版本回滚
func RollbackHistoryVersion(jsName string, version string) (string, error) {

	ScriptRepo.RepoLock.Lock()
	defer ScriptRepo.RepoLock.Unlock()

	// 仓库中 js 脚本目录
	subFolderPath, err := findSubFolder("./historyVersion", jsName+"-version-"+version)
	if err != nil {
		autoLog.Sugar.Errorf("查找子文件夹失败: %v", err)
		return fmt.Sprintf("未找到子文件夹: %s", jsName), err
	}

	// 本地 js 脚本目录
	targetPath := filepath.Join(config.Cfg.BetterGIAddress, "User", "JsScript", jsName)

	// manifest 中指定的待备份文件或目录
	manifest, err := config.ReadManifest(subFolderPath)
	if err != nil {
		return err.Error(), err
	}
	files := manifest.SavedFiles

	// 备份路径
	backupRoot := filepath.Join("backups", jsName)

	//先删除旧备份文件
	err = os.RemoveAll(backupRoot)
	if err != nil {
		autoLog.Sugar.Errorf("旧文件删除失败：%s,错误：%s", jsName, err)
		return "", err
	}
	autoLog.Sugar.Infof("旧文件删除成功")

	// 开始备份
	for _, pattern := range files {
		fullPattern := filepath.Join(targetPath, pattern)
		matches, err := filepath.Glob(fullPattern)
		if err != nil {
			autoLog.Sugar.Warnf("路径匹配失败: %s, 错误: %v", fullPattern, err)
			continue
		}

		for _, match := range matches {
			relPath, _ := filepath.Rel(targetPath, match)
			dstPath := filepath.Join(backupRoot, relPath)

			err := abgiCopy.Copy(match, dstPath)
			if err != nil {
				autoLog.Sugar.Warnf("备份失败: %s -> %s, 错误: %v", match, dstPath, err)
			} else {
				autoLog.Sugar.Infof("备份成功: %s -> %s", match, dstPath)
			}
		}
	}

	// 删除原 js 脚本目录
	os.RemoveAll(targetPath)

	// 拷贝更新的 js 脚本目录
	err = abgiCopy.Copy(subFolderPath, targetPath)
	if err != nil {
		return err.Error(), err
	}

	// 4. 还原备份内容到新脚本目录
	for _, pattern := range files {
		backupPattern := filepath.Join(backupRoot, pattern)
		matches, err := filepath.Glob(backupPattern)
		if err != nil {
			autoLog.Sugar.Warnf("还原匹配失败: %s, 错误: %v", backupPattern, err)
			continue
		}

		for _, backupItem := range matches {
			relPath, _ := filepath.Rel(backupRoot, backupItem)
			restorePath := filepath.Join(targetPath, relPath)

			_ = os.MkdirAll(filepath.Dir(restorePath), os.ModePerm)

			if err := abgiCopy.Copy(backupItem, restorePath); err != nil {
				autoLog.Sugar.Warnf("还原失败: %s -> %s, 错误: %v", backupItem, restorePath, err)
			} else {
				autoLog.Sugar.Infof("还原成功: %s -> %s", backupItem, restorePath)
			}
		}
	}

	autoLog.Sugar.Infof("Js脚本: %s 版本回滚到: %s", jsName, version)
	return fmt.Sprintf("版本回滚到: %s", version), nil

}

// 查找 repo 目录下是否存在名为 targetFolder 的子文件夹
func findSubFolder(root string, targetFolder string) (string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return "", err
	}

	for _, entry := range entries {
		if entry.IsDir() && entry.Name() == targetFolder {
			return filepath.Join(root, entry.Name()), nil
		}
	}

	return "", fmt.Errorf("未找到子文件夹: %s", targetFolder)
}

// bgi 版本回滚
func BgiRollbackHistoryVersion(jsName string, version string) error {

	now := time.Now().Format("2006-01-02-15-04-05")

	//1、备份user文件
	err4 := tools.ZipDir(config.Cfg.BetterGIAddress+"\\User\\", "Users\\User"+now+".zip", true)
	if err4 != nil {
		autoLog.Sugar.Errorf("备份失败: %v", err4)
		return fmt.Errorf("备份失败: %v", err4)
	}
	autoLog.Sugar.Infof("备份user文件")

	//备份log
	err2 := abgiCopy.Copy(config.Cfg.BetterGIAddress+"\\log\\", "backups\\log\\")
	if err2 != nil {
		autoLog.Sugar.Errorf("备份log失败: %v", err2)
		return fmt.Errorf("备份log失败: %v", err2)
	}
	autoLog.Sugar.Infof("备份log")

	//2、删除bgi文件夹
	err := os.RemoveAll(config.Cfg.BetterGIAddress)
	if err != nil {
		autoLog.Sugar.Errorf("删除bgi文件夹失败: %v", err)
		return fmt.Errorf("删除bgi文件夹失败: %v", err)
	}
	autoLog.Sugar.Infof("删除bgi文件夹")

	//3、解压bgi压缩包
	base := filepath.Base(config.Cfg.BetterGIAddress)
	path := strings.ReplaceAll(config.Cfg.BetterGIAddress, base, "")
	err4 = tools.Un7z("./historyVersion/"+jsName+"-version-"+version, path)
	if err4 != nil {
		autoLog.Sugar.Errorf("解压bgi压缩包失败: %v", err)
		return fmt.Errorf("解压bgi压缩包失败: %v", err)
	}

	autoLog.Sugar.Infof("解压bgi压缩包")

	//4、删除user文件夹
	err2 = os.RemoveAll(config.Cfg.BetterGIAddress + "\\User\\")
	if err2 != nil {
		autoLog.Sugar.Errorf("删除user文件夹失败: %v", err2)
		return fmt.Errorf("删除user文件夹失败: %v", err2)
	}
	autoLog.Sugar.Infof("删除user文件夹")

	//5、复制压缩包到User
	err3 := abgiCopy.Copy("./Users/User"+now+".zip", config.Cfg.BetterGIAddress+"\\User.zip")
	if err3 != nil {
		autoLog.Sugar.Errorf("复制压缩包到User失败: %v", err3)
		return fmt.Errorf("复制压缩包到User失败: %v", err3)
	}
	autoLog.Sugar.Infof("复制压缩包到User")

	// 6、解压user压缩包
	err4 = tools.Unzip(config.Cfg.BetterGIAddress+"\\User.zip", config.Cfg.BetterGIAddress)
	if err4 != nil {
		autoLog.Sugar.Errorf("解压user压缩包失败: %v", err4)
		return fmt.Errorf("解压user压缩包失败: %v", err4)
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

	//err5 = os.Remove("./uploads/BetterGI.7z")
	//if err5 != nil {
	//	autoLog.Sugar.Errorf("删除BetterGI压缩包失败: %v", err5)
	//}
	//autoLog.Sugar.Infof("删除BetterGI压缩包")

	return nil
}
