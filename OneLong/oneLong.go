package OneLong

import (
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"auto-bgi/control"
	"auto-bgi/task"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

type OneLong struct {
}

var oneDragon OneDragon

// StartOneLong 启动指定一条龙
func (o *OneLong) StartOneLong(longName string) {

	// 3. 关闭软件（同步，后续任务依赖此步骤）
	control.CloseSoftware()

	autoLog.Sugar.Infof("启动一条龙: %s", longName)

	task.StartOneDragon(longName)

	autoLog.Sugar.Info("一条龙启动完毕")

}

// 启动计划表
func (o *OneLong) StartOneLongPlan(PlanName string) {
	// 关闭软件（同步，后续任务依赖此步骤）
	control.CloseSoftware()

	autoLog.Sugar.Infof("启动计划表: %s", PlanName)

	task.StartOneDragonPlan(PlanName)

	autoLog.Sugar.Info("计划表启动完毕")
}

// 判断是否是公版还是茶包s
func (o *OneLong) IsChaBaoBgi(longName string) string {
	typ, err := DetectJsonType(longName)
	if err != nil {
		autoLog.Sugar.Errorf("检测 JSON 类型失败: %v", err)
		return ""
	}
	return typ
}

// OneLongAllName 读取所有一条龙配置
func (o *OneLong) OneLongAllName() []string {
	entries, err := os.ReadDir(config.Cfg.BetterGIAddress + "\\User\\OneDragon")
	if err != nil {
		return []string{}
	}
	var oneLongInfo []string
	for _, entry := range entries {

		//去除后缀：.json
		name := strings.ReplaceAll(entry.Name(), ".json", "")

		oneLongInfo = append(oneLongInfo, name)

	}
	return oneLongInfo
}

// 判断任务列表是数字键还是中文名键
func IsNumberKeyTaskList(taskList map[string]interface{}) bool {
	for key := range taskList {
		// 如果 key 的第一个字符是数字，就说明是数字型配置
		if unicode.IsDigit(rune(key[0])) {
			return true
		}
		break // 只检查第一个就够了
	}
	return false
}

// 判断 JSON 文件类型
func DetectJsonType(longName string) (string, error) {

	filename := filepath.Join(config.Cfg.BetterGIAddress, "User", "OneDragon", longName+".json")

	data, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("DetectJsonType读取文件失败: %w", err)
	}

	// 解析为 map
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return "", fmt.Errorf("解析 JSON 失败: %w", err)
	}

	taskRaw, ok := raw["TaskEnabledList"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("JSON 格式不合法，没有 TaskEnabledList")
	}

	if IsNumberKeyTaskList(taskRaw) {
		return "茶包s老师版本", nil
	}
	return "公版", nil
}
