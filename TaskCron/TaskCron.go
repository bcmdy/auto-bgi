package TaskCron

import (
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

var task map[string]func(data string)
var Tm *TaskManager

var taskOneLong OneLong.OneLong
var AudioService control.Audio

func InitTaskCron() {
	task = make(map[string]func(string))
	task["一条龙"] = func(data string) {
		autoLog.Sugar.Infof("定时任务启动：一条龙-现在时间:%s 参数:[%s]", time.Now().Format("15:04:05"), data)
		taskOneLong.OneLongTask(data)
	}
	task["配置组"] = func(data string) {
		autoLog.Sugar.Infof("定时任务启动：配置组-现在时间:%s 参数:[%s]", time.Now().Format("15:04:05"), data)
		split := strings.Split(data, " ")
		taskTask.StartGroups(split)
	}
	task["米游社签到"] = func(data string) {
		autoLog.Sugar.Infof("定时任务启动：米游社签到-现在时间:%s 参数:[%s]", time.Now().Format("15:04:05"), data)
		//taskTask.MiYouSheSign()
		go func() {
			control.CallPython()
		}()
	}
	task["关闭原神和关闭bgi"] = func(data string) {
		autoLog.Sugar.Infof("定时任务启动：关闭原神和关闭bgi-现在时间:%s 参数:[%s]", time.Now().Format("15:04:05"), data)
		control.CloseSoftware()
		control.CloseYuanShen()

	}
	task["备份user"] = func(data string) {
		autoLog.Sugar.Infof("定时任务启动：备份user-现在时间:%s 参数:[%s]", time.Now().Format("15:04:05"), data)
		err4 := bgiStatus.ZipDir(config.Cfg.BetterGIAddress+"\\User\\", "Users\\User"+time.Now().Format("2006-01-02-15-04-05")+".zip", true)
		if err4 != nil {
			autoLog.Sugar.Errorf("备份失败: %v", err4)
			return
		}
		autoLog.Sugar.Info("备份成功")
	}
	task["定时关机"] = func(data string) {
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
	task["定时重启"] = func(data string) {
		autoLog.Sugar.Infof("定时任务启动：定时重启-现在时间:%s 参数:[%s]", time.Now().Format("15:04:05"), data)
		cmd := exec.Command("shutdown", "/r", "/t", "60")
		err := cmd.Run()
		if err != nil {
			autoLog.Sugar.Errorf("定时任务启动：定时重启失败:[%s]", err)
		} else {
			autoLog.Sugar.Infof("定时任务启动：定时重启成功:[%s]", err)
		}
	}
	task["狗粮联机上线"] = func(data string) {
		autoLog.Sugar.Infof("定时任务启动：联系上线-现在时间:%s 参数:[%s]", time.Now().Format("15:04:05"), data)
		abgiSSE.OnStart()
	}

	task["狗粮联机调试上线"] = func(data string) {
		autoLog.Sugar.Infof("定时任务启动：联系上线-调试-现在时间:%s 参数:[%s]", time.Now().Format("15:04:05"), data)
		abgiSSE.OnStartDebug()
	}
	task["联机下线"] = func(data string) {
		autoLog.Sugar.Infof("定时任务启动：联机下线-现在时间:%s 参数:[%s]", time.Now().Format("15:04:05"), data)
		abgiSSE.Close()
	}
	task["电脑静音"] = func(data string) {
		autoLog.Sugar.Infof("定时任务启动：电脑静音-现在时间:%s 参数:[%s]", time.Now().Format("15:04:05"), data)
		AudioService.Mute()
	}
	task["脚本自动更新"] = func(data string) {
		autoLog.Sugar.Infof("定时任务启动：脚本自动更新-现在时间:%s 参数:[%s]", time.Now().Format("15:04:05"), data)
		if err := bgiStatus.BatchUpdateScript(); err != "" {
			autoLog.Sugar.Errorf("批量更新脚本失败: %v", err)
		} else {
			autoLog.Sugar.Infof("批量更新脚本成功")
		}
	}
	task["指定脚本更新"] = func(data string) {
		autoLog.Sugar.Infof("定时任务启动：指定脚本更新-现在时间:%s 参数:[%s]", time.Now().Format("15:04:05"), data)
		split := strings.Split(data, " ")
		js, err := bgiStatus.SpecifyUpdateJs(split)
		if err != nil {
			autoLog.Sugar.Errorf("指定脚本更新失败: %v", err)
		}
		autoLog.Sugar.Infof("指定脚本更新成功: %s", js)
		Notice.SentText("指定脚本更新成功: " + js)

	}
	task["启动bat脚本"] = func(data string) {
		autoLog.Sugar.Infof("定时任务启动：启动bat脚本-现在时间:%s 参数:[%s]", time.Now().Format("15:04:05"), data)
		taskTask.CallBat(data)
	}
	task["今日配置组执行情况通知"] = func(data string) {
		autoLog.Sugar.Infof("定时任务启动：今日配置组执行情况通知-现在时间:%s 参数:[%s]", time.Now().Format("15:04:05"), data)
		bgiStatus.TodayGroupsInfo()
	}
	task["开始obs录制"] = func(data string) {
		autoLog.Sugar.Infof("定时任务启动：开始obs录制-现在时间:%s 参数:[%s]", time.Now().Format("15:04:05"), data)
		err := abgiObs.StartRecording()
		if err != nil {
			autoLog.Sugar.Errorf("启动obs失败: %v", err)
		}
	}
	task["结束obs录制"] = func(data string) {
		autoLog.Sugar.Infof("定时任务启动：结束obs录制-现在时间:%s 参数:[%s]", time.Now().Format("15:04:05"), data)
		err := abgiObs.StopRecording(bgiStatus.BgiLogStatusInfo.Group)
		if err != nil {
			autoLog.Sugar.Errorf("结束obs录制失败: %v", err)
			return
		}
		autoLog.Sugar.Infof("结束obs录制成功,视频:" + bgiStatus.BgiLogStatusInfo.Group)
	}
	task["删除obs视频"] = func(data string) {
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
	task["键鼠操作"] = func(data string) {
		autoLog.Sugar.Infof("定时任务启动：键鼠操作-现在时间:%s 参数:[%s]", time.Now().Format("15:04:05"), data)
		control.MouseAndKeyboardClicks(data)
	}
	task["关闭obs"] = func(data string) {
		autoLog.Sugar.Infof("定时任务启动：关闭obs-现在时间:%s 参数:[%s]", time.Now().Format("15:04:05"), data)
		abgiObs.Shutdown()
	}
	Tm = NewTaskManager(models.DB)
	// 从数据库恢复任务
	Tm.loadTasksFromDB()
	Tm.Start()

}
