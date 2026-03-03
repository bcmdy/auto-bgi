package MCP

import (
	"auto-bgi/Notice"
	"auto-bgi/OneLong"
	"auto-bgi/autoLog"
	"auto-bgi/bgiStatus"
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
			Notice.SentText(fmt.Sprintf("定时任务启动：一条龙-现在时间:%s 参数:[%s]", time.Now().Format("15:04:05"), data))
			Notice.SendScreenshot()
		}()
		autoLog.Sugar.Infof("定时任务启动：一条龙-现在时间:%s 参数:[%s]", time.Now().Format("15:04:05"), data)
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
		return nil

	default:
		return fmt.Errorf("没有这个定时任务，你的定时任务有：%s", strings.Join(TaskMcpKeys(), "、"))
	}

}
