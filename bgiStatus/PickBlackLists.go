package bgiStatus

import (
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"os"
	"path/filepath"
	"strings"
)

// PickBlackLists 黑名单
type PickBlackLists struct {
	BlackLists []string `json:"BlackLists" comment:"黑名单"`
}

var manifestPath = filepath.Join(config.Cfg.BetterGIAddress, "User", "pick_black_lists.txt")

// 读取PickBlackLists.json
func (P *PickBlackLists) ReadPickBlackLists() (PickBlackLists, error) {

	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		// 创建空文件
		file, err := os.Create(manifestPath)
		if err != nil {
			return PickBlackLists{}, err
		}
		defer file.Close()
		return PickBlackLists{BlackLists: []string{}}, nil
	}

	// 读取文件内容
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return PickBlackLists{}, err
	}

	split := strings.Split(string(data), "\r\n")

	PickBlackLists := PickBlackLists{BlackLists: split}

	return PickBlackLists, nil
}

// 添加黑名单
func (P *PickBlackLists) AddPickBlackLists(blackList []string) error {
	BlackData, err := P.ReadPickBlackLists()
	if err != nil {
		autoLog.Sugar.Errorf("读取黑名单失败：%v", err)
	}

	// 原有黑名单
	blackString := BlackData.BlackLists

	// 合并新旧黑名单
	blackString = append(blackString, blackList...)

	// 去重
	unique := make(map[string]struct{})
	var dedup []string
	for _, item := range blackString {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := unique[item]; !ok {
			unique[item] = struct{}{}
			dedup = append(dedup, item)
		}
	}

	// 拼接回字符串
	newBlackString := strings.Join(dedup, "\r\n")

	// 写入文件
	err = os.WriteFile(manifestPath, []byte(newBlackString), 0644)
	if err != nil {
		autoLog.Sugar.Infof("写入黑名单失败：%v", err)
		return err
	}
	return nil
}

// 删除黑名单
func (P *PickBlackLists) DeletePickBlackLists(blackName string) error {
	BlackData, err := P.ReadPickBlackLists()
	if err != nil {
		autoLog.Sugar.Errorf("读取黑名单失败：%v", err)
		return err
	}

	blackString := BlackData.BlackLists

	var newList []string
	for _, item := range blackString {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		// 跳过要删除的
		if item == blackName {
			continue
		}
		newList = append(newList, item)
	}

	// 拼接成字符串
	newBlackString := strings.Join(newList, "\r\n")

	// 写回文件
	err = os.WriteFile(manifestPath, []byte(newBlackString), 0644)
	if err != nil {
		autoLog.Sugar.Errorf("写入黑名单失败：%v", err)
		return err
	}

	return nil
}
