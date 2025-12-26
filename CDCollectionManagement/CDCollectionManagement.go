package CDCollectionManagement

import (
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"auto-bgi/tools"
	"github.com/tidwall/gjson"
	"os"
	"path/filepath"
	"time"
)

type Record struct {
	FileName string //文件名
	CdTime   string //收集时间
	Status   string
	History  []History
}

type History struct {
	Item        map[string]int
	DurationSec int
}

// 读取所有多用户
func ReadAllUser() []string {

	directories, err := tools.ListDirectories(filepath.Join(config.Cfg.BetterGIAddress, "User", "JsScript", "采集cd管理", "record"))
	if err != nil {
		autoLog.Sugar.Errorf("读取record.json失败: %v", err)
	}
	return directories
}

// 遍历文件，获取路径和json文件
func ReadAllPathing() map[string]string {
	mapData := make(map[string]string)
	directories, err := tools.FindJSONFiles(filepath.Join(config.Cfg.BetterGIAddress, "User", "JsScript", "采集cd管理", "pathing"))
	if err != nil {
		autoLog.Sugar.Errorf("读取pathing.json失败: %v", err)
		return nil
	}
	rootPrefix := filepath.Join(config.Cfg.BetterGIAddress, "User", "JsScript", "采集cd管理", "pathing")

	for _, fullPath := range directories {

		baseName := filepath.Base(fullPath)
		dirPath := filepath.Dir(fullPath)
		relativePath, err := filepath.Rel(rootPrefix, dirPath)

		if err != nil {
			relativePath = dirPath
		}
		//finalPath := strings.ReplaceAll(relativePath, "\\", "-")

		finalPath := filepath.Base(relativePath)

		mapData[baseName] = finalPath
	}

	return mapData

}

type TreeNode struct {
	Label    string      `json:"label"`              // 显示的名称（文件夹名）
	FullPath string      `json:"fullPath,omitempty"` // 完整路径标识（可选，方便前端做key）
	Children []*TreeNode `json:"children,omitempty"` // 子文件夹
	Records  []Record    `json:"records,omitempty"`  // 该文件夹下的文件记录
}

// 读取record.json
func ReadRecord(name string) map[string]*[]Record {

	data, err := os.ReadFile(filepath.Join(config.Cfg.BetterGIAddress, "User", "JsScript", "采集cd管理", "record", name, "record.json"))
	if err != nil {
		autoLog.Sugar.Errorf("读取record.json失败: %v", err)
		return nil
	}

	result := gjson.Parse(string(data))

	collectionMap := make(map[string]*[]Record)

	pathingMap := ReadAllPathing()
	result.ForEach(func(key, value gjson.Result) bool {
		fileName := gjson.Get(value.String(), "fileName")
		standardName := pathingMap[fileName.String()]
		if _, ok := collectionMap[standardName]; !ok {
			collectionMap[standardName] = &[]Record{}
		}
		records := collectionMap[standardName]

		//历史收集
		var record Record
		value.Get("history").ForEach(func(_, hVal gjson.Result) bool {
			hist := History{
				DurationSec: int(hVal.Get("durationSec").Int()),
				Item:        make(map[string]int),
			}
			hVal.Get("items").ForEach(func(k, v gjson.Result) bool {
				hist.Item[k.String()] = int(v.Int())

				return true
			})

			record.History = append(record.History, hist)
			return true
		})

		record.FileName = fileName.String()
		parse, _ := time.Parse(time.RFC3339, value.Get("cdTime").String())
		//加8个小时
		parse = parse.Add(8 * time.Hour)
		record.CdTime = parse.Format("2006-01-02 15:04:05")
		//判断收集时间是否到达
		if time.Now().After(parse) {
			record.Status = "可采集"
		} else {
			record.Status = "冷却中"
		}
		*records = append(*records, record)
		collectionMap[standardName] = records
		return true
	})

	return collectionMap

}

type PickupRecord struct {
	Date string
	Item map[string]int
}

// 读取拾取记录
func ReadPickupRecord(name string) []PickupRecord {
	data, err := os.ReadFile(filepath.Join(config.Cfg.BetterGIAddress, "User", "JsScript", "采集cd管理", "record", name, "拾取记录.json"))
	if err != nil {
		autoLog.Sugar.Errorf("读取record.json失败: %v", err)
		return nil
	}

	var pickupRecords []PickupRecord
	result := gjson.Parse(string(data))
	result.ForEach(func(key, value gjson.Result) bool {
		pickupRecord := PickupRecord{
			Date: gjson.Get(value.String(), "date").String(),
			Item: make(map[string]int),
		}
		items := gjson.Get(value.String(), "items")
		items.ForEach(func(k, v gjson.Result) bool {
			pickupRecord.Item[k.String()] = int(v.Int())
			return true
		})
		pickupRecords = append(pickupRecords, pickupRecord)
		return true

	})
	return pickupRecords
}
