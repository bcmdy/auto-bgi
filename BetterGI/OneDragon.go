package BetterGI

import (
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type OneDragonStruct struct {
	Name            string
	TaskEnabledList []TaskEnabled
}

type TaskEnabled struct {
	Name    string
	Enabled bool
	Index   string
}

// 修改一条龙配置
func ModifyOneDragonConfig(path string, oneDragonStruct OneDragonStruct) {

	// 拼接完整的配置文件路径
	path = config.Cfg.BetterGIAddress + "\\User\\OneDragon\\" + path
	// 获取当前运行版本
	version := config.BgiCfg.RunForVersion
	// 判断版本是否包含"lcb"关键字
	if strings.Contains(version, "lcb") {
		modifyChaBao(path, oneDragonStruct)
	} else {
		modifyYaDan(path, oneDragonStruct)
	}

}

// 读取一条龙配置
// readOneDragonConfig 读取一条龙配置文件并返回OneDragonStruct结构体
// 参数path为配置文件名
func readOneDragonConfig(path string) OneDragonStruct {

	// 拼接完整的配置文件路径
	path = config.Cfg.BetterGIAddress + "\\User\\OneDragon\\" + path

	// 获取当前运行版本
	version := config.BgiCfg.RunForVersion
	// 判断版本是否包含"lcb"关键字
	if strings.Contains(version, "lcb") {
		// 如果是lcb版本，调用chaBao函数处理配置
		oneDragonStruct := chaBao(path)
		return oneDragonStruct
	} else {
		// 如果不是lcb版本，调用yaDan函数处理配置
		oneDragonStruct := yaDan(path)
		return oneDragonStruct
	}
}

func chaBao(path string) OneDragonStruct {

	// 读取配置文件内容，忽略可能出现的错误
	configData, _ := os.ReadFile(path)

	data := string(configData)

	//获取名字
	name := gjson.Get(data, "Name").String()

	// 回去任务详情
	taskEnabledList := gjson.Get(data, "TaskEnabledList")

	var taskEnAbleDs []TaskEnabled
	// 遍历取消任务列表
	taskEnabledList.ForEach(func(key, value gjson.Result) bool {

		var taskEnabled TaskEnabled

		item1 := value.Get("Item1").Bool()
		item2 := value.Get("Item2").String()

		taskEnabled.Enabled = item1
		taskEnabled.Name = item2
		taskEnabled.Index = key.String()
		taskEnAbleDs = append(taskEnAbleDs, taskEnabled)

		return true
	})

	// 排序，防止读取时乱序（例如1, 10, 2）
	sort.Slice(taskEnAbleDs, func(i, j int) bool {
		id1, _ := strconv.Atoi(taskEnAbleDs[i].Index)
		id2, _ := strconv.Atoi(taskEnAbleDs[j].Index)
		return id1 < id2
	})

	return OneDragonStruct{
		Name:            name,
		TaskEnabledList: taskEnAbleDs,
	}

}

func yaDan(path string) OneDragonStruct {
	// 读取配置文件内容，忽略可能出现的错误
	configData, _ := os.ReadFile(path)

	data := string(configData)
	//获取名字
	name := gjson.Get(data, "Name").String()

	var taskEnAbleDs []TaskEnabled
	// Reading TaskEnabledList
	taskEnabledList := gjson.Get(data, "TaskEnabledList")

	taskEnabledList.ForEach(func(key, value gjson.Result) bool {
		var taskEnabled TaskEnabled
		taskName := key.String()
		isEnabled := value.Bool()
		taskEnabled.Enabled = isEnabled
		taskEnabled.Name = taskName
		taskEnAbleDs = append(taskEnAbleDs, taskEnabled)

		return true
	})

	return OneDragonStruct{
		Name:            name,
		TaskEnabledList: taskEnAbleDs,
	}

}

func modifyChaBao(path string, oneDragonStruct OneDragonStruct) {
	configData, _ := os.ReadFile(path)
	data := configData
	//修改任务
	var err error

	// 使用Map一次性替换，避免残留旧Key
	taskListMap := make(map[string]interface{})
	for _, task := range oneDragonStruct.TaskEnabledList {
		taskListMap[task.Index] = map[string]interface{}{
			"Item1": task.Enabled,
			"Item2": task.Name,
		}
	}

	data, err = sjson.SetBytes(data, "TaskEnabledList", taskListMap)
	if err != nil {
		autoLog.Sugar.Errorf("修改TaskEnabledList失败:%d", err)
	}

	// 写回文件
	if err := os.WriteFile(path, data, 0644); err != nil {

		autoLog.Sugar.Errorf("写入 狗粮联机配置组[%s]失败:%d", config.Cfg.Account.GouLangGroupName, err)

	}
}

func modifyYaDan(path string, oneDragonStruct OneDragonStruct) {
	configData, _ := os.ReadFile(path)
	data := configData
	//修改任务
	var err error
	for _, task := range oneDragonStruct.TaskEnabledList {
		data, err = sjson.SetBytes(data, "TaskEnabledList."+task.Name, task.Enabled)
		if err != nil {
			autoLog.Sugar.Errorf("修改失败:%d", err)
		}
	}

	// 写回文件
	if err := os.WriteFile(path, data, 0644); err != nil {

		autoLog.Sugar.Errorf("写入 狗粮联机配置组[%s]失败:%d", config.Cfg.Account.GouLangGroupName, err)

	}
}
