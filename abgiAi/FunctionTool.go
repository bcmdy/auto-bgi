package abgiAi

import (
	"auto-bgi/OneLong"
	"context"
	"fmt"
	"github.com/robfig/cron/v3"
	"strings"
	"time"
	"trpc.group/trpc-go/trpc-agent-go/tool"
	"trpc.group/trpc-go/trpc-agent-go/tool/function"
)

var cronTab *cron.Cron

func init() {
	cronTab = cron.New(cron.WithSeconds())
	cronTab.Start()
}

func createFunctionTool() []tool.Tool {
	jobOneLongTool := function.NewFunctionTool(
		OneLongJob,
		function.WithName("job"),
		function.WithDescription(`定时启动一条龙功能。
参数(JSON)：{"after_seconds": 秒数(可选), "spec": "Cron表达式(可选)"}
只要提供其中一个即可；after_seconds 优先生效。返回创建结果。`),
	)
	return []tool.Tool{jobOneLongTool}
}

// AddJob 添加任务
type JobReq struct {
	AfterSeconds int    `json:"after_seconds" jsonschema:"description=一次性提醒"` // 一次性提醒（秒）
	Spec         string `json:"spec" jsonschema:"description=Cron 表达式"`       // Cron 表达式
} // empty request
type JobRsp struct {
	Text string `json:"text"`
	ID   int    `json:"id,omitempty"`
}

var oneLongJob = OneLong.OneLong{}

func OneLongJob(ctx context.Context, d JobReq) (JobRsp, error) {
	// 一次性提醒优先
	if d.AfterSeconds > 0 {
		time.AfterFunc(time.Duration(d.AfterSeconds)*time.Second, func() {
			oneLongJob.OneLongTask()
		})
		fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "添加了任务", d.Spec, d.AfterSeconds)

		return JobRsp{Text: fmt.Sprintf("好的，%d 秒后提醒你。", d.AfterSeconds)}, nil
	}

	// Cron 调度
	if strings.TrimSpace(d.Spec) == "" {
		return JobRsp{}, fmt.Errorf("缺少参数：after_seconds 或 spec")
	}
	id, err := cronTab.AddFunc(d.Spec, oneLongJob.OneLongTask)
	if err != nil {
		return JobRsp{}, fmt.Errorf("无效的 Cron 表达式: %w", err)
	}
	return JobRsp{Text: "已创建 Cron 定时任务。", ID: int(id)}, nil
}
