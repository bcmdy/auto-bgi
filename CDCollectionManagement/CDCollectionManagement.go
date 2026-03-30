package CDCollectionManagement

import (
	"auto-bgi/ScriptGroup"
	"auto-bgi/abgiConstant"
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"auto-bgi/tools"
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	copy "github.com/otiai10/copy"
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

func pathingRootDir() string {
	return filepath.Join(config.Cfg.BetterGIAddress, "User", "JsScript", "采集cd管理", "pathing")
}

func repoPathingRootDir() string {
	return filepath.Join(config.Cfg.BetterGIAddress, "Repos", "bettergi-scripts-list-git", "repo", "pathing")
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
	tree, err := GenerateTree(pathingRootDir(), ".js", ".txt", ".ini", ".ico", ".md")
	if err != nil {
		autoLog.Sugar.Errorf("ReadAllPathing:%s", err.Error())
		return nil, err
	}
	return tree, nil
}

func CreatePathingFolder(parentPath, folderName string) (string, error) {
	rootPath := filepath.Clean(pathingRootDir())
	cleanParentPath := filepath.Clean(strings.TrimSpace(parentPath))
	cleanFolderName := strings.TrimSpace(folderName)

	if cleanFolderName == "" {
		return "", fmt.Errorf("文件夹名称不能为空")
	}
	if cleanFolderName == "." || cleanFolderName == ".." {
		return "", fmt.Errorf("文件夹名称不合法")
	}
	if strings.ContainsAny(cleanFolderName, `<>:"/\|?*`) {
		return "", fmt.Errorf("文件夹名称包含非法字符")
	}
	if cleanParentPath == "" {
		cleanParentPath = rootPath
	}

	relativePath, err := filepath.Rel(rootPath, cleanParentPath)
	if err != nil {
		return "", fmt.Errorf("父级目录校验失败: %w", err)
	}
	if relativePath == ".." || strings.HasPrefix(relativePath, ".."+string(os.PathSeparator)) || filepath.IsAbs(relativePath) {
		return "", fmt.Errorf("父级目录不在采集目录范围内")
	}

	parentInfo, err := os.Stat(cleanParentPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("父级目录不存在")
		}
		return "", fmt.Errorf("读取父级目录失败: %w", err)
	}
	if !parentInfo.IsDir() {
		return "", fmt.Errorf("父级路径不是文件夹")
	}

	targetPath := filepath.Join(cleanParentPath, cleanFolderName)
	if _, err = os.Stat(targetPath); err == nil {
		return "", fmt.Errorf("文件夹已存在")
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("检查文件夹是否存在失败: %w", err)
	}

	if err = os.MkdirAll(targetPath, 0o755); err != nil {
		return "", fmt.Errorf("创建文件夹失败: %w", err)
	}

	return targetPath, nil
}

func AddPathingScript(parentPath, scriptPath string) (string, error) {
	cleanParentPath, err := validatePathingTargetDir(parentPath)
	if err != nil {
		return "", err
	}

	cleanScriptPath, err := normalizeRelativePath(scriptPath)
	if err != nil {
		return "", err
	}

	var scriptGroupConfig ScriptGroup.ScriptGroupConfig
	allPathing, err := scriptGroupConfig.ListAllPathing()
	if err != nil {
		return "", fmt.Errorf("读取脚本列表失败: %w", err)
	}

	leafPaths := make(map[string]struct{})
	collectLeafPathings(allPathing, "", leafPaths)
	if _, ok := leafPaths[cleanScriptPath]; !ok {
		return "", fmt.Errorf("请选择最后一级脚本目录")
	}

	scriptName := filepath.Base(cleanScriptPath)
	exists, err := pathingScriptExists(scriptName)
	if err != nil {
		return "", fmt.Errorf("校验脚本是否已存在失败: %w", err)
	}
	if exists {
		return "", fmt.Errorf("脚本已新增，不能重复新增")
	}

	sourcePath := filepath.Join(repoPathingRootDir(), cleanScriptPath)
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("脚本目录不存在")
		}
		return "", fmt.Errorf("读取脚本目录失败: %w", err)
	}
	if !sourceInfo.IsDir() {
		return "", fmt.Errorf("脚本路径不是文件夹")
	}

	targetPath := filepath.Join(cleanParentPath, scriptName)
	if _, err = os.Stat(targetPath); err == nil {
		return "", fmt.Errorf("当前目录已存在同名脚本")
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("检查目标脚本目录失败: %w", err)
	}

	if err = copy.Copy(sourcePath, targetPath); err != nil {
		return "", fmt.Errorf("新增脚本失败: %w", err)
	}

	return targetPath, nil
}

func DeletePathingNode(targetPath string) (string, error) {
	cleanTargetPath, err := validatePathingTargetPath(targetPath)
	if err != nil {
		return "", err
	}

	rootPath := filepath.Clean(pathingRootDir())
	if cleanTargetPath == rootPath {
		return "", fmt.Errorf("根目录不允许删除")
	}

	if fileInfo, statErr := os.Stat(cleanTargetPath); statErr == nil && !fileInfo.IsDir() {
		if err = os.Remove(cleanTargetPath); err != nil {
			return "", fmt.Errorf("删除文件失败: %w", err)
		}
		return cleanTargetPath, nil
	}

	targetInfo, err := os.Stat(cleanTargetPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("目标目录不存在")
		}
		return "", fmt.Errorf("读取目标目录失败: %w", err)
	}
	if !targetInfo.IsDir() {
		return "", fmt.Errorf("目标路径不是文件夹")
	}

	if err = os.RemoveAll(cleanTargetPath); err != nil {
		return "", fmt.Errorf("删除失败: %w", err)
	}

	return cleanTargetPath, nil
}

func validatePathingTargetDir(targetPath string) (string, error) {
	rootPath := filepath.Clean(pathingRootDir())
	cleanTargetPath := filepath.Clean(strings.TrimSpace(targetPath))
	if cleanTargetPath == "" {
		cleanTargetPath = rootPath
	}

	relativePath, err := filepath.Rel(rootPath, cleanTargetPath)
	if err != nil {
		return "", fmt.Errorf("目录校验失败: %w", err)
	}
	if relativePath == ".." || strings.HasPrefix(relativePath, ".."+string(os.PathSeparator)) || filepath.IsAbs(relativePath) {
		return "", fmt.Errorf("目标目录不在采集目录范围内")
	}

	targetInfo, err := os.Stat(cleanTargetPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("目标目录不存在")
		}
		return "", fmt.Errorf("读取目标目录失败: %w", err)
	}
	if !targetInfo.IsDir() {
		return "", fmt.Errorf("目标路径不是文件夹")
	}

	return cleanTargetPath, nil
}

func validatePathingTargetPath(targetPath string) (string, error) {
	rootPath := filepath.Clean(pathingRootDir())
	cleanTargetPath := filepath.Clean(strings.TrimSpace(targetPath))
	if cleanTargetPath == "" {
		cleanTargetPath = rootPath
	}

	relativePath, err := filepath.Rel(rootPath, cleanTargetPath)
	if err != nil {
		return "", fmt.Errorf("目录校验失败: %w", err)
	}
	if relativePath == ".." || strings.HasPrefix(relativePath, ".."+string(os.PathSeparator)) || filepath.IsAbs(relativePath) {
		return "", fmt.Errorf("目标不在采集目录范围内")
	}

	if _, err = os.Stat(cleanTargetPath); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("目标不存在")
		}
		return "", fmt.Errorf("读取目标失败: %w", err)
	}

	return cleanTargetPath, nil
}

func normalizeRelativePath(path string) (string, error) {
	cleanPath := strings.TrimSpace(path)
	cleanPath = strings.ReplaceAll(cleanPath, "/", string(os.PathSeparator))
	cleanPath = strings.Trim(cleanPath, `\/`)
	if cleanPath == "" {
		return "", fmt.Errorf("脚本路径不能为空")
	}

	cleanPath = filepath.Clean(cleanPath)
	if cleanPath == "." || cleanPath == ".." {
		return "", fmt.Errorf("脚本路径不合法")
	}
	if filepath.IsAbs(cleanPath) || strings.HasPrefix(cleanPath, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("脚本路径不合法")
	}
	return cleanPath, nil
}

func collectLeafPathings(items []ScriptGroup.AllPath, parent string, leafPaths map[string]struct{}) {
	for _, item := range items {
		currentPath := item.FileName
		if parent != "" {
			currentPath = filepath.Join(parent, item.FileName)
		}

		if len(item.FileNameChild) == 0 {
			leafPaths[currentPath] = struct{}{}
			continue
		}

		collectLeafPathings(item.FileNameChild, currentPath, leafPaths)
	}
}

func pathingScriptExists(scriptName string) (bool, error) {
	scriptName = strings.TrimSpace(scriptName)
	if scriptName == "" {
		return false, nil
	}

	stopErr := errors.New("script found")
	err := filepath.WalkDir(pathingRootDir(), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == pathingRootDir() || !d.IsDir() {
			return nil
		}
		if d.Name() == scriptName {
			return stopErr
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, stopErr) {
			return true, nil
		}
		return false, err
	}
	return false, nil
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

// 调整后的拾取记录结构：外层map是分类，内层map是物品名-数量
type PickupRecord struct {
	Date string
	Item map[string]map[string]int // 结构：{"采集": {"晶蝶": 5, "铁块": 3}, "怪物掉落": {"史莱姆凝液": 10}}
}

// 读取拾取记录（修复覆盖问题+完善分类逻辑）
func ReadPickupRecord(name string) []PickupRecord {
	// 1. 读取JSON文件
	data, err := os.ReadFile(filepath.Join(config.Cfg.BetterGIAddress, "User", "JsScript", "采集cd管理", "record", name, "拾取记录.json"))
	if err != nil {
		autoLog.Sugar.Errorf("读取record.json失败: %v", err)
		return nil
	}

	var pickupRecords []PickupRecord
	result := gjson.Parse(string(data))

	// 2. 遍历每条拾取记录
	result.ForEach(func(key, value gjson.Result) bool {
		pickupRecord := PickupRecord{
			Date: gjson.Get(value.String(), "date").String(),
			Item: make(map[string]map[string]int), // 初始化外层分类map
		}

		// 3. 解析当前记录的所有物品
		items := gjson.Get(value.String(), "items")
		items.ForEach(func(k, v gjson.Result) bool {
			// 3.1 获取物品名和数量
			itemName := k.String()
			itemCount := int(v.Int())

			// 3.2 获取物品分类（从常量映射表读取，无匹配则为"未知"）
			category := abgiConstant.MaterialCategoryMap[itemName]
			categoryStr := string(category)
			if categoryStr == "" {
				categoryStr = "未知"
				autoLog.Sugar.Warnf("物品[%s]未匹配到分类，默认归为'未知'", itemName)
			}

			if strings.Contains(itemName, "蟹") {
				itemName = "螃蟹"
			} else if strings.Contains(itemName, "晶蝶") {
				itemName = "晶蝶"
			} else if strings.Contains(itemName, "蜥") {
				itemName = "蜥"
			} else if strings.Contains(itemName, "鳗") {
				itemName = "鳗肉"
			} else if strings.Contains(itemName, "鲈") {
				itemName = "鲈鱼"
			}

			// 3.4 添加物品到分类
			// 如果该分类还未初始化内层map，先初始化
			if pickupRecord.Item[categoryStr] == nil {
				pickupRecord.Item[categoryStr] = make(map[string]int)
			}
			// 向该分类的内层map添加物品-数量（支持同一分类多个物品）
			pickupRecord.Item[categoryStr][itemName] = pickupRecord.Item[categoryStr][itemName] + itemCount

			//// 3.3 核心修复：避免同一分类下物品被覆盖
			//// 如果该分类还未初始化内层map，先初始化
			//if pickupRecord.Item[categoryStr] == nil {
			//	pickupRecord.Item[categoryStr] = make(map[string]int)
			//}
			//// 向该分类的内层map添加物品-数量（支持同一分类多个物品）
			//pickupRecord.Item[categoryStr][itemName] = pickupRecord.Item[categoryStr][itemName] + itemCount

			return true
		})

		// 4. 将当前记录加入结果列表
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
