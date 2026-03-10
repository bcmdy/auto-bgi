package bgiStatus

import (
	"auto-bgi/Notice"
	"auto-bgi/autoLog"
	"auto-bgi/control"
	"auto-bgi/task"
	"fmt"
	"strings"
	"time"
)

// JsLogRestart js日志触发重启
func JsLogRestart(line string) {

	if strings.HasPrefix(line, "OpenCV内存异常, 重试1次") {
		autoLog.Sugar.Errorf("bgi严重错误：%s", line)

		// 构造通知内容
		content := fmt.Sprintf(
			"🚨 【BGI严重错误】\n"+
				"------------------\n"+
				"⚠️ 报警项：OpenCV 内存异常\n"+
				"🔍 详情：%s\n"+
				"⏰ 时间：%s\n"+
				"🛠 abgi解决方法：将在3秒后重启bgi，并且继续跑剩余的没有跑完的配置组",
			line,
			time.Now().Format("2006-01-02 15:04:05"),
		)
		Notice.SentText(content)

		//等待3秒
		time.Sleep(3 * time.Second)

		// 获取没有跑完的配置组
		noRunningOneLongProgress := NoRunningOneLongProgress()
		if len(noRunningOneLongProgress) == 0 {
			// 所有配置组都跑完了，关闭原神和bgi
			Notice.SentText("所有配置组都跑完了，关闭原神和bgi")
			autoLog.Sugar.Infof("所有配置组都跑完了，关闭原神和bgi")
			// 关闭原神和bgi
			control.CloseSoftware()
			control.CloseYuanShen()

		} else {
			// 有配置组没有跑完，继续跑
			Notice.SentText(fmt.Sprintf("有配置组没有跑完，继续跑：%s", strings.Join(noRunningOneLongProgress, "、")))
			autoLog.Sugar.Infof("有配置组没有跑完，继续跑：%s", strings.Join(noRunningOneLongProgress, "、"))
			task.StartGroups(noRunningOneLongProgress)
		}

	}

}
