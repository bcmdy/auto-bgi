package OneLong

import (
	"auto-bgi/abgiObs"
	"auto-bgi/autoLog"
	"auto-bgi/bgiStatus"
	"auto-bgi/config"
	"auto-bgi/control"
)

// 启动一条龙
func (o *OneLong) OneLongTask(longName string) {
	autoLog.Sugar.Info("开始执行一条龙任务")

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

	o.StartOneLong(longName)

	autoLog.Sugar.Info("一条龙任务执行完成")
}
