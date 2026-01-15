package AbgiBot

import (
	"auto-bgi/Notice"
	"auto-bgi/OneLong"
	"auto-bgi/ScriptRepo"
	"auto-bgi/abgiObs"
	"auto-bgi/abgiSSE"
	"auto-bgi/autoLog"
	"auto-bgi/bgiStatus"
	"auto-bgi/control"
	"auto-bgi/task"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
)

var bot *tgbotapi.BotAPI

var oneLong OneLong.OneLong

// 初始化机器人（带代理）
// InitTG 函数用于初始化 Telegram Bot
// 参数:
//   - token: Telegram Bot 的 API token
//   - proxy: 代理服务器地址，可以为空
//
// 返回值:
//   - error: 初始化过程中出现的错误
func InitTG(token, proxy string) error {
	// 如果 token 为空，则跳过初始化
	if token == "" {
		return nil // 允许空 token，跳过初始化
	}
	var client *http.Client
	// 如果提供了代理地址，则创建带有代理的 HTTP 客户端
	if proxy != "" {
		pu, err := url.Parse(proxy)
		if err != nil {
			return err
		}
		client = &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(pu)}}
	} else {
		// 否则使用默认的 HTTP 客户端
		client = http.DefaultClient
	}
	var err error
	// 使用提供的 token 和 HTTP 客户端创建 Bot API 实例
	bot, err = tgbotapi.NewBotAPIWithClient(token, tgbotapi.APIEndpoint, client)
	if err != nil {
		fmt.Println("TG配置错误", err)
		return err
	}

	// 配置更新方式：偏移量为0，超时为60秒
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	// 开始接收更新
	updates := bot.GetUpdatesChan(u)
	go func() {
		for update := range updates {
			fmt.Println("=========", update.Message.Text)

			Message := BotCommand(update.Message.Text)
			fmt.Println(Message)

			// 回复消息
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, Message)

			bot.Send(msg)

		}
	}()

	log.Printf("[TG] bot authorized: @%s", bot.Self.UserName)
	return nil
}

func BotCommand(command string) string {
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
			return fmt.Sprintf("⚠通知：💗\n配置组：%s\n路线：%s💗\n%s", info.Group, info.MapTrackingLine, info.Timestamp)
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
		"批量更新脚本": func() string {
			if err := bgiStatus.BatchUpdateScript(); err != "" {
				autoLog.Sugar.Errorf("批量更新脚本失败: %v", err)
				return "批量更新脚本失败"
			} else {
				autoLog.Sugar.Infof("批量更新脚本成功")
				return "批量更新脚本成功"
			}
		},
		"关闭原神": func() string {
			control.CloseYuanShen()
			return "关闭成功"
		},
		"关闭bgi": func() string {
			control.CloseSoftware()
			return "关闭成功"
		},
		"电脑关机": func() string {
			// Windows 关机命令：立即关机
			cmd := exec.Command("shutdown", "/s", "/t", "60")

			err := cmd.Run()
			if err != nil {
				return "💗关机失败💗"
			} else {
				return "💗60秒之后，电脑将会关机💗"
			}
		},
		"取消关机": func() string {
			// Windows 取消关机命令
			cmd := exec.Command("shutdown", "/a")
			err := cmd.Run()
			if err != nil {
				return "💗取消关机失败💗"
			}
			return "💗取消关机成功💗"
		},
	}

	if response, exists := commandMap[command]; exists {
		return response()
	}
	if strings.HasPrefix(command, "老王") {
		//conversation, err := abgiAi.Conversation(command)
		//if err != nil {
		//	return "老王不想回答你"
		//}
		//return conversation

	} else if strings.HasPrefix(command, "启动一条龙") {
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
	} else if strings.Contains(command, "订阅脚本") {
		ScriptName := strings.Replace(command, "订阅脚本", "", -1)
		script, err := ScriptRepo.SubscribeScript(ScriptName)
		if err != nil {
			return fmt.Sprintf("订阅失败：%v", err.Error())
		}
		return script
	}
	return "老王不想回答你的无礼问题..."
}
