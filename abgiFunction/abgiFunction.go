package abgiFunction

import (
	"auto-bgi/AbgiBot"
	"auto-bgi/OneLong"
	"auto-bgi/TaskCron"
	"auto-bgi/abgiObs"
	"auto-bgi/abgiSSE"
	"auto-bgi/auth"
	"auto-bgi/autoLog"
	"auto-bgi/bgiStatus"
	"auto-bgi/config"
	"auto-bgi/control"
	"auto-bgi/frpc"
	"auto-bgi/task"
	"os"
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

	if config.Cfg.Account.SecretKey != "" {
		//每天0点清空联机次数
		go task.ClearRunCount()
	}

	//实时读取文件
	go bgiStatus.LogM()

	if config.Cfg.OneRemote.IsMonitor {
		go bgiStatus.Log1Remote()
		autoLog.Sugar.Infof("1Remote监控开启状态")
	}

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

	//if config.Cfg.AbgiAiConfig.IsAbgiAi {
	//	autoLog.Sugar.Infof("AI开启状态")
	//	go abgiAi.InitAi()
	//}

	// 检查目标目录是否存在，若不存在则创建
	if _, err := os.Stat("Users"); os.IsNotExist(err) {
		autoLog.Sugar.Infof("目录不存在，正在创建：%s", "Users")
		err := os.MkdirAll("Users", os.ModePerm) // 创建多级目录
		if err != nil {
			autoLog.Sugar.Errorf("创建Users失败: %v", err)

		}
		autoLog.Sugar.Infof("目录Users成功：%s", "Users")
	}

	//读取bgi配置
	config.ReadBgiConfig()

	//定时任务
	TaskCron.InitTaskCron()

	//登录验证
	auth.InitAuth()

	//判断根目录也没有frpc_user.toml
	if _, err := os.Stat("frpc_user.toml"); os.IsNotExist(err) {
		autoLog.Sugar.Infof("===启动成功===")
	} else {
		go frpc.InitFrp()
	}
	//检查重启文件
	check()

	go func() {
		gameRolesRes := config.InitA()
		if len(gameRolesRes.Data.List) != 0 {
			autoLog.Sugar.Infof("游戏角色列表：%v", gameRolesRes.Data.List)
			task.SendMoraleRankTask()
		}
	}()

	abgiSSE.Status()
	control.GetSysTemUser()

}
