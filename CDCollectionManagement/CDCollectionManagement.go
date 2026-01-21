package CDCollectionManagement

import (
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"auto-bgi/tools"
	"fmt"
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
		return nil
	}
	return directories
}

// 遍历文件，获取路径和json文件
func ReadAllPathing() (*FileNode, error) {
	tree, err := GenerateTree(filepath.Join(config.Cfg.BetterGIAddress, "User", "JsScript", "采集cd管理", "pathing"), ".js", ".txt", ".ini", ".ico", ".md")
	if err != nil {
		autoLog.Sugar.Errorf("ReadAllPathing:%s", err.Error())
		return nil, err
	}
	return tree, nil
}

// 读取record.json
func ReadRecord(name string) *FileNode {

	data, err := os.ReadFile(filepath.Join(config.Cfg.BetterGIAddress, "User", "JsScript", "采集cd管理", "record", name, "record.json"))
	if err != nil {
		autoLog.Sugar.Errorf("读取record.json失败: %v", err)
		return nil
	}

	result := gjson.Parse(string(data))

	var records []Record

	recordMap := make(map[string]Record)

	result.ForEach(func(key, value gjson.Result) bool {
		fileName := gjson.Get(value.String(), "fileName")

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
		parse, _ := time.Parse(time.RFC3339Nano, value.Get("cdTime").String())
		localTime := parse.Local()
		record.CdTime = localTime.Format("2006-01-02 15:04:05")
		//判断收集时间是否到达
		if time.Now().After(localTime) {
			record.Status = "可采集"
		} else {
			record.Status = "冷却中"
		}

		records = append(records, record)
		recordMap[fileName.String()] = record
		return true
	})

	pathing, err := ReadAllPathing()
	if err != nil {
		return nil
	}

	PrintTree(pathing, recordMap)

	return pathing

}

func PrintTree(node *FileNode, recordMap map[string]Record) {
	if node == nil {
		return
	}
	//判断key是否存在
	if _, ok := recordMap[node.Name]; ok {
		node.Record = recordMap[node.Name]
	}
	// 3. 递归遍历子节点
	for _, child := range node.Children {
		PrintTree(child, recordMap)
	}
	return
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

type FileNode struct {
	Name     string      `json:"name"`               // 文件名
	Path     string      `json:"path"`               // 完整路径
	IsDir    bool        `json:"is_dir"`             // 是否是目录
	Size     int64       `json:"size"`               // 大小 (字节)
	Children []*FileNode `json:"children,omitempty"` // 子节点 (如果是文件则为空)
	Record   Record      `json:"record,omitempty"`
}

// GenerateTree 递归生成文件树
func GenerateTree(path string, exclude ...string) (*FileNode, error) {
	// 1. 获取当前路径的文件信息
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	// 2. 创建当前节点
	node := &FileNode{
		Name:  info.Name(),
		Path:  path,
		IsDir: info.IsDir(),
		Size:  info.Size(),
	}

	// 3. 如果是文件，直接返回（没有子节点）
	if !info.IsDir() {
		return node, nil
	}

	// 4. 如果是目录，读取目录下所有内容
	entries, err := os.ReadDir(path)
	if err != nil {
		// 如果读取目录失败（例如权限问题），这里可以选择返回错误，或者记录日志并返回空目录
		// 这里选择返回错误
		return nil, err
	}

	// 5. 遍历目录内容，递归构建子节点
	for _, entry := range entries {

		if tools.DetermineFileType(entry.Name(), exclude...) {
			//fmt.Println(entry.Name())
			continue
		}

		// 构建子文件的完整路径
		childPath := filepath.Join(path, entry.Name())

		// 递归调用
		childNode, err := GenerateTree(childPath, exclude...)
		if err != nil {
			// 这里可以根据需求处理子节点的错误，例如跳过权限不足的文件
			fmt.Printf("Error processing %s: %v\n", childPath, err)
			continue
		}

		node.Children = append(node.Children, childNode)
	}

	return node, nil
}
