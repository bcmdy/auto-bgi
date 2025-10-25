package OneLong

import (
	"auto-bgi/abgiObs"
	"auto-bgi/autoLog"
	"auto-bgi/bgiStatus"
	"auto-bgi/config"
	"auto-bgi/control"
	"database/sql"
	"fmt"
	"github.com/robfig/cron/v3"
	"time"
)

func (o *OneLong) StartOneLongTask() {

	cronTab := cron.New(cron.WithSeconds())

	// 定时任务,cron表达式
	spec := fmt.Sprintf("0 %d %d * * *", config.Cfg.OneLong.OneLongMinute, config.Cfg.OneLong.OneLongHour)

	// 定义定时器调用的任务函数
	task := func() {

		autoLog.Sugar.Infof("一条龙服务启动 %v", time.Now().Format("2006-01-02 15:04:05"))

		o.OneLongTask()

		time.Sleep(1000 * time.Millisecond)

		schedule, err := config.Parser.Parse(spec)
		if err != nil {
			autoLog.Sugar.Error("解析失败:", err)
			return
		}

		autoLog.Sugar.Infof("一条龙服务启动完毕 %v", schedule.Next(time.Now()).Format("2006-01-02 15:04:05"))
	}

	// 添加定时任务
	cronTab.AddFunc(spec, task)
	// 启动定时器
	cronTab.Start()
	// 阻塞主线程停止
	select {}

}

// 启动一条龙
func (o *OneLong) OneLongTask() {
	autoLog.Sugar.Info("开始执行一条龙任务")

	// 2. 并行执行用户目录备份
	go func() {
		autoLog.Sugar.Info("开始备份 User 目录")
		o.backupUsers()
	}()

	// 3. 关闭软件（同步，后续任务依赖此步骤）
	control.CloseSoftware()
	autoLog.Sugar.Info("软件已关闭")

	// 4. 批量更新脚本
	if config.Cfg.OneLong.AutoUpdateJs {
		autoLog.Sugar.Info("开始批量更新脚本")
		if err := bgiStatus.BatchUpdateScript(); err != "" {
			autoLog.Sugar.Errorf("批量更新脚本失败: %v", err)

		}
	} else {
		autoLog.Sugar.Info("自动更新js已关闭")
	}

	//是否开启obs回放缓存
	if config.Cfg.Control.OBSReplayBuffer {
		autoLog.Sugar.Info("开启回放obs缓存")
		err := abgiObs.StartReplayBuffer()
		if err != nil {
			autoLog.Sugar.Errorf("回放obs缓存失败: %v", err)
		}
	} else {
		autoLog.Sugar.Info("不开启回放obs缓存")
	}

	// 6. 启动今日一条龙
	longName := config.GetTodayOneLongName()
	autoLog.Sugar.Infof("今日启动一条龙: %s", longName)

	o.StartOneLong(longName)

	autoLog.Sugar.Info("一条龙任务执行完成")
}

const interval = 72 * time.Hour

// 每周一备份users文件夹
func (o *OneLong) backupUsers() {

	var lastBackupStr string
	err := config.DB.QueryRow(`SELECT autobgi_value FROM autoBgi_config WHERE autobgi_key = 'BackupUserTime'`).Scan(&lastBackupStr)
	if err != nil && err != sql.ErrNoRows {
		autoLog.Sugar.Errorf("查询 BackupUserTime 失败: %v", err)
		return
	}
	// 解析上次时间
	var lastBackup time.Time
	if lastBackupStr != "" {
		parsed, per := time.ParseInLocation("2006-01-02 15:04:05", lastBackupStr, time.Local)
		if per == nil {
			lastBackup = parsed
		} else {
			autoLog.Sugar.Warnf("时间解析失败(%v)，使用默认时间", per)
			lastBackup = time.Now().Add(-interval)
		}
	}

	now := time.Now()

	if now.Sub(lastBackup) >= interval {
		autoLog.Sugar.Info("🟢 满足条件，开始备份 users 文件夹...")
		autoLog.Sugar.Infof("开始备份user文件夹")
		err4 := bgiStatus.ZipDir(config.Cfg.BetterGIAddress+"\\User\\", "Users\\User"+time.Now().Format("2006100215020405")+".zip", true)
		if err4 != nil {
			autoLog.Sugar.Errorf("备份失败: %v", err4)
			return
		}

		autoLog.Sugar.Info("备份成功")

		// 更新数据库记录
		_, err = config.DB.Exec(`UPDATE autoBgi_config SET autobgi_value = ? WHERE autobgi_key = 'BackupUserTime'`, time.Now().Format("2006-01-02 15:04:05"))
		if err != nil {
			autoLog.Sugar.Errorf("更新 BackupUserTime 失败: %v", err)
		} else {
			autoLog.Sugar.Info("✅ 备份完成，时间已更新")
		}
	} else {
		autoLog.Sugar.Infof("⏳ 未满足条件（上次：%v，下次至少需等待：%.0f小时）", lastBackup, (interval - now.Sub(lastBackup)).Hours())
	}
}
