package ScriptGroup

import (
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"auto-bgi/control"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func listGroups() ([]string, error) {
	// 指定要读取的文件夹路径
	//自定义配置路径
	folderPath := config.Cfg.BetterGIAddress + "\\User\\ScriptGroup"

	var groupNames []string

	// 遍历文件夹
	err := filepath.WalkDir(folderPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// 检查是否是 JSON 文件
		if filepath.Ext(d.Name()) == ".json" {
			// 打印 JSON 文件名（相对于文件夹的路径）
			relativePath, err := filepath.Rel(folderPath, path)
			if err != nil {
				return err
			}

			name := strings.Replace(relativePath, ".json", "", -1)

			groupNames = append(groupNames, name)

		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return groupNames, nil
}

// StartGroups 启动配置组
func startGroups(names []string) error {
	control.CloseSoftware()
	time.Sleep(5 * time.Second)

	betterGIPath := filepath.Join(config.Cfg.BetterGIAddress, "BetterGI.exe")

	// 检查文件是否存在
	if _, err := os.Stat(betterGIPath); err != nil {
		autoLog.Sugar.Errorf("BetterGI.exe 不存在: %v", err)
		return err
	}

	args := append([]string{"--startGroups"}, names...) // 每个组名单独参数

	cmd := exec.Command(betterGIPath, args...)

	cmd.Stdout = nil
	cmd.Stderr = nil
	err := cmd.Start()
	if err != nil {
		autoLog.Sugar.Errorf("启动配置组失败: %v", err)
		return err
	}
	autoLog.Sugar.Infof("启动配置组成功: %v", names)
	return nil
}
