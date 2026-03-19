package bgiStatus

import (
	"auto-bgi/JsAPI"
	"auto-bgi/Notice"
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"auto-bgi/tools"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// 读取js的md文件
func ReadMd(filePath string) string {

	path := ""
	split := strings.Split(filePath, "/")

	if strings.Contains(filePath, "js/") {
		path = split[0] + "/" + split[1]
	} else if strings.Contains(filePath, "combat/") || strings.Contains(filePath, "tcg/") {
		filename := filepath.Clean(fmt.Sprintf("%s\\Repos\\bettergi-scripts-list-git\\repo\\%s", config.Cfg.BetterGIAddress, filePath))
		// 读取文件内容
		data, err := os.ReadFile(filename)
		if err != nil {
			autoLog.Sugar.Errorf("ReadMd读取文件失败: %v", err)
			return "*该作者没有写说明文档，无脑直接用就完事了*"
		}
		return string(data)
	} else {
		for i := range len(split) - 1 {
			path += split[i] + "/"
		}
	}

	filename := filepath.Clean(fmt.Sprintf("%s\\Repos\\bettergi-scripts-list-git\\repo\\%s\\README.md", config.Cfg.BetterGIAddress, path))

	// 读取文件内容
	data, err := os.ReadFile(filename)
	if err != nil {
		autoLog.Sugar.Errorf("ReadMd读取文件失败: %v", err)
		return "*该作者没有写说明文档，无脑直接用就完事了*"
	}

	return string(data)

}

// 指定脚本更新
func SpecifyUpdateJs(subDirs []string) (string, error) {
	GitPull()
	time.Sleep(1)

	// 获取本地所有订阅脚本目录
	scriptDir := filepath.Join(config.Cfg.BetterGIAddress, "User", "JsScript")

	for _, name := range subDirs {

		nowVersion, _ := GetJsNowVersion(scriptDir, name)
		newVersion, chineseName, err := GetJsNewVersion(name)
		if err != nil {
			continue
		}

		if tools.CompareVersion(nowVersion, newVersion) == -1 {
			if name == "CD-Aware-AutoGather" {
				autoLog.Sugar.Infof("脚本更新：带CD管理的自动采集有更新，版本号是：" + newVersion + "。如需更新，请手动更新")
				Notice.SentText("带CD管理的自动采集有更新，版本号是：" + newVersion + "。如需更新，请手动更新")
				continue
			} else if name == "采集cd管理" {
				autoLog.Sugar.Infof("脚本更新：采集cd管理有更新，版本号是：" + newVersion + "。如需更新，请手动更新")
				Notice.SentText("采集cd管理有更新，版本号是：" + newVersion + "。如需更新，请手动更新")
			}

			// 开始更新
			_, err := UpdateJs(name)
			if err != nil {
				autoLog.Sugar.Errorf("更新脚本失败: %v", err)
				continue
			}
			autoLog.Sugar.Infof("更新脚本成功: %s", chineseName)
			Notice.SentText(fmt.Sprintf("脚本 %s 已更新,版本号:%s @所有人", chineseName, newVersion))
		}

	}
	return "全部更新成功", nil
}

// 批量更新脚本
func BatchUpdateScript() string {
	GitPull()
	time.Sleep(1)

	// 获取本地所有订阅脚本目录
	scriptDir := filepath.Join(config.Cfg.BetterGIAddress, "User", "JsScript")
	subDirs, err := tools.ListSubDirsOnly(scriptDir)
	if err != nil {
		autoLog.Sugar.Errorf("获取本地脚本失败: %v", err)
		return "获取本地脚本失败"
	}

	for _, name := range subDirs {

		nowVersion, _ := GetJsNowVersion(scriptDir, name)
		newVersion, chineseName, err := GetJsNewVersion(name)
		if err != nil {
			continue
		}

		if tools.CompareVersion(nowVersion, newVersion) == -1 {
			if name == "CD-Aware-AutoGather" {
				autoLog.Sugar.Infof("脚本更新：带CD管理的自动采集有更新，版本号是：" + newVersion + "。如需更新，请手动更新")
				Notice.SentText("带CD管理的自动采集有更新，版本号是：" + newVersion + "。如需更新，请手动更新")
				continue
			}

			repoDir := filepath.Join(config.Cfg.BetterGIAddress, "Repos", "bettergi-scripts-list-git", "repo", "js", name)
			JsAPI.BackupHistoryVersion(repoDir, newVersion)

			// 开始更新
			_, err := UpdateJs(name)
			if err != nil {
				autoLog.Sugar.Errorf("更新脚本失败: %v", err)
				continue
			}
			autoLog.Sugar.Infof("更新脚本成功: %s", chineseName)
			Notice.SentText(fmt.Sprintf("脚本 %s 已更新,版本号:%s @所有人", chineseName, newVersion))
		}

	}
	return ""
}
