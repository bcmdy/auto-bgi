package OneLong

import (
	"auto-bgi/abgiObs"
	"auto-bgi/autoLog"
	"auto-bgi/bgiStatus"
	"auto-bgi/config"
	"auto-bgi/control"
	"database/sql"
	"time"
)

// 启动一条龙
func (o *OneLong) OneLongTask(longName string) {
	autoLog.Sugar.Info("开始执行一条龙任务")

	// 3. 关闭软件（同步，后续任务依赖此步骤）
	control.CloseSoftware()
	autoLog.Sugar.Info("软件已关闭")

	// 2. 并行执行用户目录备份
	go func() {
		autoLog.Sugar.Info("开始备份 User 目录")
		o.backupUsers()
	}()

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

	o.StartOneLong(longName)

	autoLog.Sugar.Info("一条龙任务执行完成")
}

// const interval = 72 * time.Hour

// 每周一备份users文件夹
func (o *OneLong) backupUsers() {

	interval := time.Duration(config.Cfg.Control.BackupUsersHour) * time.Hour

	//捕获异常
	defer func() {
		if err := recover(); err != nil {
			autoLog.Sugar.Errorf("backupUsers捕获异常: %v", err)
			return
		}
	}()

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
		err4 := bgiStatus.ZipDir(config.Cfg.BetterGIAddress+"\\User\\", "Users\\User"+time.Now().Format("2006-01-02-15-04-05")+".zip", true)
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
		autoLog.Sugar.Infof("未满足条件：%v，需等待：%.0f小时", lastBackup, (interval - now.Sub(lastBackup)).Hours())
	}
}
