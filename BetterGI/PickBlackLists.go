package BetterGI

import (
	"auto-bgi/config"
	"os"
	"path/filepath"
)

// PickBlackLists 黑名单
type PickBlackLists struct {
	BlackLists string `json:"BlackLists" comment:"黑名单"`
}

// 读取PickBlackLists.json
func (P *PickBlackLists) ReadPickBlackLists() (PickBlackLists, error) {
	manifestPath := filepath.Join(config.Cfg.BetterGIAddress, "User", "pick_black_lists.txt")
	//没有就创建

	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		// 创建空文件
		file, err := os.Create(manifestPath)
		if err != nil {
			return PickBlackLists{}, err
		}
		defer file.Close()
		return PickBlackLists{BlackLists: ""}, nil
	}

	// 读取文件内容
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return PickBlackLists{}, err
	}

	PickBlackLists := PickBlackLists{BlackLists: string(data)}

	return PickBlackLists, nil
}
