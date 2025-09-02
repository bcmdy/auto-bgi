package bgiStatus

import (
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"auto-bgi/tools"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// 读取js的md文件
func ReadMd(filePath string) string {

	if strings.Contains(filePath, "=>") {
		filePath = strings.TrimSpace(strings.Split(filePath, "=>")[1])
		if strings.Contains(filePath, "archive") {
			return "归档操作"
		}
	}

	path := ""
	split := strings.Split(filePath, "/")

	if strings.Contains(filePath, "/js/") {
		path = split[0] + "/" + split[1] + "/" + split[2]
	} else if strings.Contains(filePath, "/combat/") {
		filename := filepath.Clean(fmt.Sprintf("%s\\Repos\\bettergi-scripts-list-git\\%s", config.Cfg.BetterGIAddress, filePath))
		// 读取文件内容
		data, err := os.ReadFile(filename)
		if err != nil {
			autoLog.Sugar.Errorf("ReadMd读取文件失败: %v", err)
			return "作者没有写说明文档"
		}
		return string(data)
	} else {
		for i := range len(split) - 1 {
			path += split[i] + "/"
		}
	}

	filename := filepath.Clean(fmt.Sprintf("%s\\Repos\\bettergi-scripts-list-git\\%s\\README.md", config.Cfg.BetterGIAddress, path))

	// 读取文件内容
	data, err := os.ReadFile(filename)
	if err != nil {
		autoLog.Sugar.Errorf("ReadMd读取文件失败: %v", err)
		return "作者没有写说明文档"
	}

	return string(data)

}

// 批量更新脚本
func BatchUpdateScript() string {
	if err := GitPull(); err != nil {
		autoLog.Sugar.Errorf("仓库更新失败，再次尝试一下:%s", err.Error())
		if err := GitPull(); err != nil {
			autoLog.Sugar.Errorf("仓库第二次再次更新失败:%s", err.Error())
			if err := GitPull(); err != nil {
				autoLog.Sugar.Errorf("仓库第三次再次更新失败:%s", err.Error())
			}
		}
	}

	// 获取本地所有订阅脚本目录
	scriptDir := filepath.Join(config.Cfg.BetterGIAddress, "User", "JsScript")
	subDirs, err := tools.ListSubDirsOnly(scriptDir)
	if err != nil {
		autoLog.Sugar.Errorf("获取本地脚本失败: %v", err)
		return "获取本地脚本失败"
	}

	for _, name := range subDirs {
		nowVersion := getJsNowVersion(scriptDir, name)
		newVersion, chineseName, err := GetJsNewVersion(name)
		if err != nil {
			continue
		}
		if nowVersion != newVersion {
			// 开始更新
			_, err := UpdateJs(name)
			if err != nil {
				autoLog.Sugar.Errorf("更新脚本失败: %v", err)
				continue
			}
			autoLog.Sugar.Infof("更新脚本成功: %s", chineseName)
			SentText(fmt.Sprintf("脚本 %s 已更新,版本号:%s", chineseName, newVersion))
		}

	}
	return ""
}
