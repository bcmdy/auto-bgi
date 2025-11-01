package AbgiBot

import (
	"auto-bgi/Notice"
	"auto-bgi/abgiObs"
	"auto-bgi/abgiSSE"
	"auto-bgi/autoLog"
	"auto-bgi/bgiStatus"
	"auto-bgi/config"
	"auto-bgi/control"
	"auto-bgi/task"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	lark "github.com/larksuite/oapi-sdk-go/v3"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	larkws "github.com/larksuite/oapi-sdk-go/v3/ws"
	"strings"
)

var larkWsClient *larkws.Client
var larkClient *lark.Client

func InitFeiShuBot() {

	larkClient = lark.NewClient(config.Cfg.Notice.FeiShu.AppID, config.Cfg.Notice.FeiShu.AppSecret)

	// 注册事件 Register event
	eventHandler := dispatcher.NewEventDispatcher("", "").
		OnP2MessageReceiveV1(func(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
			prettify := larkcore.Prettify(event.Event.Message.Content)

			var inner string
			if err := json.Unmarshal([]byte(prettify), &inner); err != nil {
				panic(err)
			}

			var m map[string]string
			if err := json.Unmarshal([]byte(inner), &m); err != nil {
				panic(err)
			}
			data := strings.ReplaceAll(m["text"], "@_user_1 ", "")

			fmt.Printf("收到消息: %s\n", data)
			fmt.Println("回话id", event.Event.Message.MessageId)
			command := FeiShuBotCommand(data)
			send(*event.Event.Message.MessageId, command)

			return nil
		})

	larkWsClient = larkws.NewClient(config.Cfg.Notice.FeiShu.AppID, config.Cfg.Notice.FeiShu.AppSecret,
		larkws.WithEventHandler(eventHandler),
		//larkws.WithLogLevel(larkcore.LogLevelDebug),
	)

	// 建立长连接 Establish persistent connection
	err := larkWsClient.Start(context.Background())
	if err != nil {
		autoLog.Sugar.Errorf("飞书机器人启动失败")
		panic(err)
	}
}

func FeiShuBotCommand(command string) string {
	commandMap := map[string]func() string{
		"联机上线": func() string {
			abgiSSE.OnStart()
			return "上线成功"
		},
		"联机下线": func() string {
			abgiSSE.Close()
			return "下线成功"
		},
		"情况": func() string {
			info := bgiStatus.BgiLogStatusInfo
			return fmt.Sprintf("⚠️通知：💗💗\n配置组：%s\n路线：%s💗\n%s", info.Group, info.MapTrackingLine, info.Timestamp)
		},
		"截图": func() string {
			Notice.SendScreenshot()
			return "发送成功"
		},
		"开始录屏": func() string {
			err := abgiObs.StartRecording()
			if err != nil {
				return fmt.Sprintf("启动失败：%v", err.Error())
			}
			return "录屏成功"
		},
		"停止录屏": func() string {
			info := bgiStatus.BgiLogStatusInfo
			err := abgiObs.StopRecording(info.Group)
			if err != nil {
				return fmt.Sprintf("停止失败：%v", err.Error())
			}
			return "停止成功"
		},
		"关闭原神": func() string {
			control.CloseYuanShen()
			return "关闭成功"
		},
		"关闭bgi": func() string {
			control.CloseSoftware()
			return "关闭成功"
		},
	}

	if response, exists := commandMap[command]; exists {
		return response()
	}

	if strings.Contains(command, "启动一条龙") {
		oneLong.StartOneLong(strings.Replace(command, "启动一条龙", "", -1))
		return "启动一条龙成功"
	} else if strings.Contains(command, "启动配置组") {
		groups := strings.Replace(command, "启动配置组", "", -1)
		data := strings.Split(groups, " ")
		err := task.StartGroups(data)
		if err != nil {
			return fmt.Sprintf("启动失败：%v", err.Error())
		}
		return "启动配置组成功"
	}

	return "指令错误"
}

// send 函数用于发送回复消息
// 参数:
//   - MessageId: 需要回复的消息ID
func send(MessageId, content string) {

	s := `{"text":"content"}`
	res := strings.ReplaceAll(s, "content", content)

	// 使用构建器模式创建回复消息请求
	req := larkim.NewReplyMessageReqBuilder().
		// 设置要回复的消息ID
		MessageId(MessageId).
		// 设置消息请求体
		Body(larkim.NewReplyMessageReqBodyBuilder().
			// 设置消息内容，使用JSON格式
			Content(res).
			// 设置消息类型为文本
			MsgType(`text`).
			// 设置为在原消息线程中回复
			ReplyInThread(true).
			// 设置请求的唯一标识符
			Uuid(uuid.NewString()).
			// 构建消息请求体
			Build()).
		// 构建完整的请求对象
		Build()
	// 发起请求，调用IM服务的消息回复接口
	resp, err := larkClient.Im.V1.Message.Reply(context.Background(), req)

	// 处理错误
	if err != nil {
		fmt.Println(err)
		return
	}

	// 服务端错误处理
	if !resp.Success() {
		fmt.Printf("logId: %s, error response: \n%s", resp.RequestId(), larkcore.Prettify(resp.CodeError))
		return
	}

	//// 业务处理
	//fmt.Println(larkcore.Prettify(resp))
}
