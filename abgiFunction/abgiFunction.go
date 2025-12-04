package abgiFunction

import (
	"auto-bgi/AbgiBot"
	"auto-bgi/OneLong"
	"auto-bgi/TaskCron"
	"auto-bgi/abgiAi"
	"auto-bgi/abgiObs"
	"auto-bgi/auth"
	"auto-bgi/autoLog"
	"auto-bgi/bgiStatus"
	"auto-bgi/config"
	"auto-bgi/internal/mysConfig"
	"auto-bgi/task"
)

var OneLongService OneLong.OneLong

func InitFunction() {
	//检查BGI状态
	go bgiStatus.CheckBetterGIStatus()

	//开启每隔一小时发送截图
	if config.Cfg.Control.SendWeChatImage {
		autoLog.Sugar.Infof("开启每隔一小时发送截图")
		go task.SendWeChatImageTask()
	} else {
		autoLog.Sugar.Infof("关闭每隔一小时发送截图")
	}

	//实时读取文件
	go bgiStatus.LogM()

	if config.Cfg.OneRemote.IsMonitor {
		go bgiStatus.Log1Remote()
		autoLog.Sugar.Infof("1Remote监控开启状态")
	}

	//米游社自动签到
	mysConfig.LoadConfig("mysConfig.yaml")

	//obs是否连接
	if config.Cfg.ScreenRecord.IsRecord {
		err := abgiObs.EnsureConnected()
		if err != nil {
			autoLog.Sugar.Infof("OBS连接失败")
		} else {
			autoLog.Sugar.Infof("OBS连接成功")
		}
	}

	//初始化bgi日志信息
	bgiStatus.InitBgiLogStatus()

	if config.Cfg.CommandBot.TgBOT {
		go func() {
			err := AbgiBot.InitTG(config.Cfg.Notice.TGNotice.TGToken, config.Cfg.Notice.TGNotice.Proxy)
			if err != nil {
				autoLog.Sugar.Infof("TG机器人开启失败")
			} else {
				autoLog.Sugar.Infof("TG机器人开启状态")
			}
		}()
	}

	if config.Cfg.CommandBot.FeiShuBot {
		autoLog.Sugar.Infof("飞书机器人开启状态")
		go AbgiBot.InitFeiShuBot()
	}

	if config.Cfg.AbgiAiConfig.IsAbgiAi {
		autoLog.Sugar.Infof("AI开启状态")
		go abgiAi.InitAi()
	}

	//读取bgi配置
	config.ReadBgiConfig()

	//定时任务
	TaskCron.InitTaskCron()

	//登录验证
	auth.InitAuth()

}
