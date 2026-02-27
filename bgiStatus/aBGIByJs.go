package bgiStatus

import (
	"auto-bgi/Notice"
	"auto-bgi/abgiObs"
	"auto-bgi/abgiSSE"
	"auto-bgi/autoLog"
	"auto-bgi/control"
	"auto-bgi/task"
	"auto-bgi/tools"
	"strings"
)

//cmd测试
//chcp 65001 >nul
//echo ABGI启动联机调试： >> better-genshin-impact20260226.log

func JsLogHandler(line string) {
	//ABGI启动重启ABGI:
	if strings.HasPrefix(line, "ABGI启动重启ABGI：") {
		autoLog.Sugar.Infof("js日志调用ABGI启动重启ABGI")
		err := tools.RestartProgram()
		if err != nil {
			autoLog.Sugar.Errorf("js日志调用ABGI启动重启ABGI失败: %v", err)
		}
	}

	//关闭原神
	if strings.HasPrefix(line, "ABGI启动关闭原神：") {
		autoLog.Sugar.Infof("js日志调用ABGI启动关闭原神")
		control.CloseYuanShen()
		data := strings.ReplaceAll(line, "ABGI启动关闭原神：", "")
		if strings.Contains(data, "配置组-") {
			//配置组
			groupName := strings.ReplaceAll(data, "配置组-", "")
			split := strings.Split(groupName, " ")
			err := task.StartGroups(split)
			if err != nil {
				autoLog.Sugar.Errorf("js日志调用abgi启动配置组失败: %v", err)
			} else {
				autoLog.Sugar.Infof("js日志调用abgi关闭原神启动配置组成功: %s", groupName)
			}
		} else if strings.Contains(data, "一条龙-") {
			//一条龙
			oneLongName := strings.ReplaceAll(data, "一条龙-", "")
			autoLog.Sugar.Infof("js日志调用abgi关闭原神启动一条龙" + oneLongName)
			//// 关闭软件
			control.CloseSoftware()
			task.StartOneDragon(oneLongName)
		}

	}

	//js日志调用abgi联机换号
	if strings.HasPrefix(line, "ABGI启动联机换号：") {
		data := strings.ReplaceAll(line, "ABGI启动联机换号：", "")
		abgiSSE.ChangeAccount(data)
	}

	//js日志调用abgi启动一条龙
	if strings.HasPrefix(line, "ABGI启动一条龙：") {

		oneLongName := strings.ReplaceAll(line, "ABGI启动一条龙：", "")
		autoLog.Sugar.Infof("js日志调用abgi启动一条龙" + oneLongName)
		//// 关闭软件
		control.CloseSoftware()
		task.StartOneDragon(oneLongName)
	}

	//js日志调用abgi启动配置组
	if strings.HasPrefix(line, "ABGI启动配置组：") {

		autoLog.Sugar.Infof("js日志调用abgi启动配置组")
		groupName := strings.ReplaceAll(line, "ABGI启动配置组：", "")
		split := strings.Split(groupName, " ")
		err := task.StartGroups(split)
		if err != nil {
			autoLog.Sugar.Errorf("js日志调用abgi启动配置组失败: %v", err)
		}
	}
	//js日志调用abgi联机上线
	if strings.HasPrefix(line, "ABGI启动联机上线：") {

		autoLog.Sugar.Infof("js日志调用abgi联机上线")
		abgiSSE.OnStart()
	}

	//js日志调用abgi联机下线
	if strings.HasPrefix(line, "ABGI启动联机下线：") {
		autoLog.Sugar.Infof("js日志调用abgi联机下线")
		abgiSSE.Close()
	}

	//ABGI启动联机调试：
	if strings.HasPrefix(line, "ABGI启动联机调试：") {
		abgiSSE.OnStartDebug()
		autoLog.Sugar.Infof("js日志调用ABGI启动联机调试")
	}
	//ABGI启动BAT脚本：
	if strings.HasPrefix(line, "ABGI启动BAT脚本：") {
		data := strings.ReplaceAll(line, "ABGI启动BAT脚本：", "")
		task.CallBat(data)

	}

	//ABGI启动脚本更新
	if strings.HasPrefix(line, "ABGI启动脚本更新：") {
		names := strings.ReplaceAll(line, "ABGI启动脚本更新：", "")
		split := strings.Split(names, " ")

		js, err := SpecifyUpdateJs(split)
		if err != nil {
			autoLog.Sugar.Errorf("指定脚本更新失败: %v", err)
		}
		autoLog.Sugar.Infof("指定脚本更新成功: %s", js)
		Notice.SentText("指定脚本更新成功: " + js)
	}

	//ABGI启动今日配置组执行情况通知
	if strings.HasPrefix(line, "ABGI启动今日配置组执行情况通知：") {
		TodayGroupsInfo()
		autoLog.Sugar.Infof("js日志调用ABGI启动今日配置组执行情况通知")
	}

	//ABGI启动关闭原神和关闭bgi：
	if strings.HasPrefix(line, "ABGI启动关闭原神和关闭bgi：") {
		control.CloseSoftware()
		control.CloseYuanShen()
		autoLog.Sugar.Infof("js日志调用ABGI启动关闭原神和关闭bgi")
	}

	//ABGI启动电脑静音：
	if strings.HasPrefix(line, "ABGI启动电脑静音：") {
		AudioService.Mute()
		autoLog.Sugar.Infof("js日志调用ABGI启动电脑静音")
	}

	//js日志调用abgi启动关闭obs
	if strings.HasPrefix(line, "ABGI启动obs：") {

		autoLog.Sugar.Infof("ABGI启动obs")
		data := strings.ReplaceAll(line, "ABGI启动obs：", "")
		if data == "启动" {
			err := abgiObs.StartRecording()
			if err != nil {
				autoLog.Sugar.Errorf("js日志调用abgi启动obs失败: %v", err)
			}
		} else if data == "关闭" {
			err := abgiObs.StopRecording(BgiLogStatusInfo.Group)
			if err != nil {
				autoLog.Sugar.Errorf("js日志调用abgi关闭obs失败: %v", err)
			}
		}

	}

	//js日志调用abgi米游社签到
	if strings.HasPrefix(line, "ABGI启动米游社签到：") {
		autoLog.Sugar.Infof("ABGI启动米游社签到")
		go func() {
			control.CallPython()
		}()
	}

	//js日志调用abgi更换房间
	if strings.HasPrefix(line, "ABGI启动更换房间：") {
		autoLog.Sugar.Infof(line)
		data := strings.ReplaceAll(line, "ABGI启动更换房间：", "")
		split := strings.Split(data, "-")
		if len(split) == 2 {
			abgiSSE.ModifyRoom(split[0], split[1])
		} else {
			autoLog.Sugar.Errorf("更换房间参数错误")
		}

	}

}
