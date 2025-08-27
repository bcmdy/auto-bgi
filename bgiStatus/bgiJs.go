package bgiStatus

import (
	"auto-bgi/autoLog"
	"auto-bgi/config"
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

	split := strings.Split(filePath, "/")

	path := ""

	for i := range len(split) - 1 {
		path += split[i] + "/"
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
