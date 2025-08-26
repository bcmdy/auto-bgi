package bgiStatus

import (
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"fmt"
	"os"
	"path/filepath"
)

// 读取js的md文件
func ReadMd(group string, name string) string {

	filename := filepath.Clean(fmt.Sprintf("%s\\Repos\\bettergi-scripts-list-git\\repo\\%s\\%s\\README.md", config.Cfg.BetterGIAddress, group, name))

	// 读取文件内容
	data, err := os.ReadFile(filename)
	if err != nil {
		autoLog.Sugar.Errorf("ReadMd读取文件失败: %v", err)
		return ""
	}

	return string(data)

}
