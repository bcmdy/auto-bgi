package AbgiBot

import (
	"auto-bgi/autoLog"
	"auto-bgi/config"
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
			command := BotCommand(data)
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
