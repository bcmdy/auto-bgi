package MCP

import (
	"auto-bgi/Notice"
	"auto-bgi/OneLong"
	"auto-bgi/autoLog"
	"auto-bgi/bgiStatus"
	"auto-bgi/config"
	"auto-bgi/control"
	taskTask "auto-bgi/task"
	"fmt"
	"github.com/iancoleman/orderedmap"
	"slices"
	"strings"
	"time"
)

type Res struct {
	Time    string
	Msg     string
	MapData []*orderedmap.OrderedMap
}

var OneLongService OneLong.OneLong

func GetBgiIndex() (Res, error) {

	info := bgiStatus.BgiLogStatusInfo
	data := make([]*orderedmap.OrderedMap, 8)
	data[0] = orderedmap.New()
	data[0].Set("当前配置组：", info.Group+" ["+info.GroupProgress+"]")
	data[0].Set("预计：", info.Timestamp)
	data[0].Set("当前路线：", info.MapTrackingLine)
	data[0].Set("进度：", info.ConfigurationGroupExecutionProgress)
	data[0].Set("bgi运行状态：", info.Running)
	data[0].Set("js运行进度：", info.JSProgress)
	var res Res
	res.Time = time.Now().Format("2006-01-02")
	res.Msg = "bgi运行情况"
	res.MapData = data
	return res, nil
}

// 启动一条龙
func StartOneLong(name string) (Res, error) {

	OneLongService.StartOneLong(name)
	var res Res
	res.Time = time.Now().Format("2006-01-02")
	res.Msg = "启动一条龙" + name
	return res, nil
}

var taskOneLong OneLong.OneLong

// 定时任务列表
var TaskMcp = make(map[string]func(string))

func InitTaskCron() {
	TaskMcp["启动一条龙"] = func(data string) {
		//参数校验
		longAllName := OneLongService.OneLongAllName()
		//判断参数是否包含
		if !slices.Contains(longAllName, data) {
			autoLog.Sugar.Errorf("没有这个一条龙，你的一条龙有：%s", strings.Join(longAllName, "、"))
		}

		taskOneLong.OneLongTask(data)
		go func() {
			Notice.SentText(fmt.Sprintf("MCP定时任务启动：一条龙-现在时间:%s 参数:[%s]", time.Now().Format("15:04:05"), data))
			Notice.SendScreenshot()
		}()
		autoLog.Sugar.Infof("MCP定时任务启动：一条龙-现在时间:%s 参数:[%s]", time.Now().Format("15:04:05"), data)
	}
	TaskMcp["关闭原神和关闭bgi"] = func(data string) {
		autoLog.Sugar.Infof("MCP定时任务启动：关闭原神和关闭bgi-现在时间:%s 参数:[%s]", time.Now().Format("15:04:05"), data)
		control.CloseSoftware()
		control.CloseYuanShen()
	}
	TaskMcp["米游社签到"] = func(data string) {
		autoLog.Sugar.Infof("MCP定时任务启动：米游社签到-现在时间:%s 参数:[%s]", time.Now().Format("15:04:05"), data)
		go func() {
			control.CallPython()
		}()
	}
	TaskMcp["启动配置组"] = func(data string) {
		autoLog.Sugar.Infof("MCP定时任务启动：配置组-现在时间:%s 参数:[%s]", time.Now().Format("15:04:05"), data)
		split := strings.Split(data, " ")
		taskTask.StartGroups(split)
	}
	TaskMcp["备份user"] = func(data string) {
		autoLog.Sugar.Infof("MCP定时任务启动：备份user-现在时间:%s 参数:[%s]", time.Now().Format("15:04:05"), data)
		err4 := bgiStatus.ZipDir(config.Cfg.BetterGIAddress+"\\User\\", "Users\\User"+time.Now().Format("2006-01-02-15-04-05")+".zip", true)
		if err4 != nil {
			autoLog.Sugar.Errorf("备份失败: %v", err4)
			return
		}
		autoLog.Sugar.Info("备份成功")
	}
}

func TaskMcpKeys() []string {
	var keys []string
	for k := range TaskMcp {
		keys = append(keys, k)
	}
	return keys
}

// 新增定时任务
func AddTaskMcp(taskName string, delay int, data string) (Res, error) {

	fn := TaskMcp[taskName]
	if fn == nil {
		autoLog.Sugar.Errorf("没有这个定时任务，你的定时任务有：%s", strings.Join(TaskMcpKeys(), "、"))
		return Res{}, fmt.Errorf("没有这个定时任务，你的定时任务有：%s", strings.Join(TaskMcpKeys(), "、"))
	}

	//校验参数
	err := parameterValidation(taskName, data)
	if err != nil {
		return Res{
			Time: time.Now().Format("2006-01-02"),
			Msg:  err.Error(),
		}, err
	}

	if delay > 0 {
		// 新增定时任务
		time.AfterFunc(time.Duration(delay)*time.Second, func() {
			fn(data)
		})

	} else {
		go fn(data)
	}
	autoLog.Sugar.Infof("新增定时任务：%s-现在时间:%s 参数:[%s] 延迟:%d秒，任务类型：%s", taskName, time.Now().Format("15:04:05"), data, delay, taskName)
	return Res{
		Time: time.Now().Format("2006-01-02"),
		Msg:  fmt.Sprintf("新增定时任务：%s-现在时间:%s 参数:[%s] 延迟:%d秒，任务类型：%s", taskName, time.Now().Format("15:04:05"), data, delay, taskName),
	}, nil
}

// 参数校验
func parameterValidation(taskName, data string) error {
	switch taskName {
	case "启动一条龙":
		longAllName := OneLongService.OneLongAllName()
		//判断参数是否包含
		if !slices.Contains(longAllName, data) {
			autoLog.Sugar.Errorf("没有这个一条龙，你的一条龙有：%s", strings.Join(longAllName, "、"))
			return fmt.Errorf("没有这个一条龙，你的一条龙有：%s", strings.Join(longAllName, "、"))
		}
		OneLongService.StartOneLong(data)
		return nil
	case "启动配置组":
		groups, _ := taskTask.ListGroups()
		//判断参数是否包含
		if !slices.Contains(groups, data) {
			autoLog.Sugar.Errorf("没有这个配置组，你的配置组有：%s", strings.Join(groups, "、"))
			return fmt.Errorf("没有这个配置组，你的配置组有：%s", strings.Join(groups, "、"))
		}
		taskTask.StartGroups([]string{data})
		return nil
	default:
		return nil
	}

}
