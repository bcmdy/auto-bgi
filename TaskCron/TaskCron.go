package TaskCron

import (
	"auto-bgi/BetterGI"
	"auto-bgi/Notice"
	"auto-bgi/OneLong"
	"auto-bgi/abgiObs"
	"auto-bgi/abgiSSE"
	"auto-bgi/autoLog"
	"auto-bgi/bgiStatus"
	"auto-bgi/config"
	"auto-bgi/control"
	"auto-bgi/models"
	taskTask "auto-bgi/task"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type TaskManager struct {
	c    *cron.Cron
	jobs map[cron.EntryID]models.TaskCron
	db   *gorm.DB
}

func NewTaskManager(db *gorm.DB) *TaskManager {
	return &TaskManager{
		c:    cron.New(cron.WithSeconds()),
		jobs: make(map[cron.EntryID]models.TaskCron),
		db:   db,
	}
}

// 启动调度器
func (tm *TaskManager) Start() {
	tm.c.Start()
}

// var Task map[string]func(data string)
var Task = make(map[string]func(string))

var Tm *TaskManager

var taskOneLong OneLong.OneLong
var AudioService control.Audio

func InitTaskCron() {
	//Task = make(map[string]func(string))
	Task["一条龙"] = func(data string) {
		autoLog.Sugar.Infof("定时任务启动：一条龙-现在时间:%s 参数:[%s]", time.Now().Format("15:04:05"), data)
		taskOneLong.OneLongTask(data)
	}
	Task["计划表(茶包)"] = func(data string) {
		autoLog.Sugar.Infof("定时任务启动：计划表-现在时间:%s 参数:[%s]", time.Now().Format("15:04:05"), data)
		taskOneLong.StartOneLongPlan(data)
	}
	Task["配置组"] = func(data string) {
		autoLog.Sugar.Infof("定时任务启动：配置组-现在时间:%s 参数:[%s]", time.Now().Format("15:04:05"), data)
		split := strings.Split(data, " ")
		taskTask.StartGroups(split)
	}
	Task["米游社签到"] = func(data string) {
		autoLog.Sugar.Infof("定时任务启动：米游社签到-现在时间:%s 参数:[%s]", time.Now().Format("15:04:05"), data)
		//taskTask.MiYouSheSign()
		go func() {
			control.CallPython()
		}()
	}
	Task["关闭原神和关闭bgi"] = func(data string) {
		autoLog.Sugar.Infof("定时任务启动：关闭原神和关闭bgi-现在时间:%s 参数:[%s]", time.Now().Format("15:04:05"), data)
		control.CloseSoftware()
		control.CloseYuanShen()

	}
	Task["备份user"] = func(data string) {
		autoLog.Sugar.Infof("定时任务启动：备份user-现在时间:%s 参数:[%s]", time.Now().Format("15:04:05"), data)
		err4 := bgiStatus.ZipDir(config.Cfg.BetterGIAddress+"\\User\\", "Users\\User"+time.Now().Format("2006-01-02-15-04-05")+".zip", true)
		if err4 != nil {
			autoLog.Sugar.Errorf("备份失败: %v", err4)
			return
		}
		autoLog.Sugar.Info("备份成功")
	}
	Task["定时关机"] = func(data string) {
		abgiObs.Shutdown()
		autoLog.Sugar.Infof("定时任务启动：定时关机-现在时间:%s 参数:[%s]", time.Now().Format("15:04:05"), data)
		cmd := exec.Command("shutdown", "/s", "/t", "60")
		err := cmd.Run()
		if err != nil {
			autoLog.Sugar.Errorf("定时任务启动：定时关机失败:[%s]", err)
		} else {
			autoLog.Sugar.Infof("定时任务启动：定时关机成功:[%s]", err)
		}
	}
	Task["定时重启"] = func(data string) {
		autoLog.Sugar.Infof("定时任务启动：定时重启-现在时间:%s 参数:[%s]", time.Now().Format("15:04:05"), data)
		cmd := exec.Command("shutdown", "/r", "/t", "60")
		err := cmd.Run()
		if err != nil {
			autoLog.Sugar.Errorf("定时任务启动：定时重启失败:[%s]", err)
		} else {
			autoLog.Sugar.Infof("定时任务启动：定时重启成功:[%s]", err)
		}
	}
	Task["狗粮联机上线"] = func(data string) {
		autoLog.Sugar.Infof("定时任务启动：联系上线-现在时间:%s 参数:[%s]", time.Now().Format("15:04:05"), data)
		abgiSSE.OnStart()
	}

	Task["狗粮联机调试上线"] = func(data string) {
		autoLog.Sugar.Infof("定时任务启动：联系上线-调试-现在时间:%s 参数:[%s]", time.Now().Format("15:04:05"), data)
		abgiSSE.OnStartDebug()
	}
	Task["联机下线"] = func(data string) {
		autoLog.Sugar.Infof("定时任务启动：联机下线-现在时间:%s 参数:[%s]", time.Now().Format("15:04:05"), data)
		abgiSSE.Close()
	}
	Task["联机更换房间"] = func(data string) {
		autoLog.Sugar.Infof("定时任务启动：联机更换房间-现在时间:%s 参数:[%s]", time.Now().Format("15:04:05"), data)
		data = strings.ReplaceAll(data, "ABGI启动更换房间：", "")
		split := strings.Split(data, "-")
		if len(split) == 2 {
			abgiSSE.ModifyRoom(split[0], split[1])
		} else {
			autoLog.Sugar.Errorf("更换房间参数错误")
		}
	}
	Task["电脑静音"] = func(data string) {
		autoLog.Sugar.Infof("定时任务启动：电脑静音-现在时间:%s 参数:[%s]", time.Now().Format("15:04:05"), data)
		AudioService.Mute()
	}
	Task["脚本自动更新"] = func(data string) {
		autoLog.Sugar.Infof("定时任务启动：脚本自动更新-现在时间:%s 参数:[%s]", time.Now().Format("15:04:05"), data)
		if err := bgiStatus.BatchUpdateScript(); err != "" {
			autoLog.Sugar.Errorf("批量更新脚本失败: %v", err)
		} else {
			autoLog.Sugar.Infof("批量更新脚本成功")
		}
	}
	Task["指定脚本更新"] = func(data string) {
		autoLog.Sugar.Infof("定时任务启动：指定脚本更新-现在时间:%s 参数:[%s]", time.Now().Format("15:04:05"), data)
		split := strings.Split(data, " ")
		js, err := bgiStatus.SpecifyUpdateJs(split)
		if err != nil {
			autoLog.Sugar.Errorf("指定脚本更新失败: %v", err)
		}
		autoLog.Sugar.Infof("指定脚本更新成功: %s", js)
		Notice.SentText("指定脚本更新成功: " + js)

	}
	Task["启动bat脚本"] = func(data string) {
		autoLog.Sugar.Infof("定时任务启动：启动bat脚本-现在时间:%s 参数:[%s]", time.Now().Format("15:04:05"), data)
		taskTask.CallBat(data)
	}
	Task["今日配置组执行情况通知"] = func(data string) {
		autoLog.Sugar.Infof("定时任务启动：今日配置组执行情况通知-现在时间:%s 参数:[%s]", time.Now().Format("15:04:05"), data)
		bgiStatus.TodayGroupsInfo()
	}
	Task["开始obs录制"] = func(data string) {
		autoLog.Sugar.Infof("定时任务启动：开始obs录制-现在时间:%s 参数:[%s]", time.Now().Format("15:04:05"), data)
		err := abgiObs.StartRecording()
		if err != nil {
			autoLog.Sugar.Errorf("启动obs失败: %v", err)
		}
	}
	Task["结束obs录制"] = func(data string) {
		autoLog.Sugar.Infof("定时任务启动：结束obs录制-现在时间:%s 参数:[%s]", time.Now().Format("15:04:05"), data)
		err := abgiObs.StopRecording(bgiStatus.BgiLogStatusInfo.Group)
		if err != nil {
			autoLog.Sugar.Errorf("结束obs录制失败: %v", err)
			return
		}
		autoLog.Sugar.Infof("结束obs录制成功,视频:" + bgiStatus.BgiLogStatusInfo.Group)
	}
	Task["删除obs视频"] = func(data string) {
		autoLog.Sugar.Infof("定时任务启动：删除obs视频-现在时间:%s 参数:[%s]", time.Now().Format("15:04:05"), data)
		//转成数字
		num, err := strconv.Atoi(data)
		if err != nil {
			autoLog.Sugar.Errorf("删除obs视频失败: %v", err)
			return
		}
		err = abgiObs.DeleteVideosByAge(num)
		if err != nil {
			autoLog.Sugar.Errorf("删除obs视频失败: %v", err)
		}
	}
	Task["键鼠操作"] = func(data string) {
		autoLog.Sugar.Infof("定时任务启动：键鼠操作-现在时间:%s 参数:[%s]", time.Now().Format("15:04:05"), data)
		control.MouseAndKeyboardClicks(data)
	}
	Task["关闭obs"] = func(data string) {
		autoLog.Sugar.Infof("定时任务启动：关闭obs-现在时间:%s 参数:[%s]", time.Now().Format("15:04:05"), data)
		abgiObs.Shutdown()
	}
	Task["开启obs回放缓存"] = func(data string) {
		autoLog.Sugar.Infof("定时任务启动：开启obs回放缓存-现在时间:%s 参数:[%s]", time.Now().Format("15:04:05"), data)
		err := abgiObs.StartReplayBuffer()
		if err != nil {
			autoLog.Sugar.Errorf("开启obs回放缓存失败: %v", err)
		}
	}
	Task["关闭obs回放缓存"] = func(data string) {
		autoLog.Sugar.Infof("定时任务启动：关闭obs回放缓存-现在时间:%s 参数:[%s]", time.Now().Format("15:04:05"), data)
		err := abgiObs.StopReplayBuffer()
		if err != nil {
			autoLog.Sugar.Errorf("关闭obs回放缓存失败: %v", err)
		}
	}
	Task["一条龙全部开启/关闭"] = func(data string) {
		//备份
		err4 := bgiStatus.ZipDir(config.Cfg.BetterGIAddress+"\\User\\", "Users\\User"+time.Now().Format("2006-01-02-15-04-05")+".zip", true)
		if err4 != nil {
			autoLog.Sugar.Errorf("备份失败: %v", err4)
			return
		}
		autoLog.Sugar.Info("备份成功")

		autoLog.Sugar.Infof("定时任务启动：一条龙全部开启/关闭-现在时间:%s 参数:[%s]", time.Now().Format("15:04:05"), data)
		split := strings.Split(data, "-")
		if len(split) == 2 {
			BetterGI.SetOneDragonAllStatus(split[0]+".json", split[1] == "开启")
			autoLog.Sugar.Infof("一条龙全部开启/关闭成功")
		} else {
			autoLog.Sugar.Errorf("一条龙全部开启/关闭参数错误")
		}

	}
	Task["更新aBgi"] = func(data string) {

	}

}

// 数据库任务初始化
func TmStart() {
	Tm = NewTaskManager(models.DB)
	// 从数据库恢复任务
	Tm.loadTasksFromDB()
	Tm.Start()
}
