package bgiStatus

import (
	"auto-bgi/config"
	"fmt"
	"os"
	"path/filepath"
)

// 读取js的md文件
func ReadJsMd(jsName string) string {

	filename := filepath.Clean(fmt.Sprintf("%s\\Repos\\bettergi-scripts-list-git\\repo\\js\\%s\\READEME.md", config.Cfg.BetterGIAddress, jsName))

	// 读取文件内容
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("读取文件失败: %v\n", err)
		return ""
	}

	return string(data)

}
