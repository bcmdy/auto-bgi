package TaskCron

import (
	"auto-bgi/autoLog"
	"auto-bgi/config"
	taskOneLong "auto-bgi/task"
	"database/sql"
	"github.com/robfig/cron/v3"
	"strings"
	"time"
)

type TaskCron struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Spec string `json:"spec"`
	Next string `json:"next"`
	Data string `json:"data"` // NEW: 任务参数
}

type TaskManager struct {
	c    *cron.Cron
	jobs map[cron.EntryID]TaskCron
	db   *sql.DB
}

func NewTaskManager(db *sql.DB) *TaskManager {
	return &TaskManager{
		c:    cron.New(cron.WithSeconds()),
		jobs: make(map[cron.EntryID]TaskCron),
		db:   db,
	}
}

// 启动调度器
func (tm *TaskManager) Start() {
	tm.c.Start()
}

var task map[string]func(data string)
var Tm *TaskManager

func InitTaskCron() {
	task = make(map[string]func(string))
	task["一条龙"] = func(data string) {
		autoLog.Sugar.Infof("定时任务启动：一条龙现在时间:%s 参数:%s", time.Now().Format("15:04:05"), data)
		taskOneLong.StartOneDragon(data)
	}
	task["配置组"] = func(data string) {
		autoLog.Sugar.Infof("定时任务启动：配置组现在时间:%s 参数:%s", time.Now().Format("15:04:05"), data)
		split := strings.Split(data, " ")
		taskOneLong.StartGroups(split)
	}

	Tm = NewTaskManager(config.DB)
	// 从数据库恢复任务
	Tm.loadTasksFromDB()
	Tm.Start()

}
