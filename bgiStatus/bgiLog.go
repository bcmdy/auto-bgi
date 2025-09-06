package bgiStatus

import (
	"auto-bgi/Notice"
	"auto-bgi/abgiSSE"
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"auto-bgi/control"
	"bufio"
	"fmt"
	"github.com/go-vgo/robotgo"
	"io"
	"os"
	"regexp"
	"strings"
	"syscall"
	"time"
)

type LogMonitor struct {
	LogFile      string
	Keywords     []string
	WebhookURL   string
	ScanInterval int
	lastPosition int64
	stopChan     chan struct{}
}

func NewLogMonitor(logFile string, keywords []string, interval int) *LogMonitor {
	return &LogMonitor{
		LogFile:      logFile,
		Keywords:     keywords,
		WebhookURL:   config.Cfg.Notice.Wechat,
		ScanInterval: interval,
		stopChan:     make(chan struct{}),
	}
}

func (m *LogMonitor) scanLog() ([]string, error) {
	f, err := os.Open(m.LogFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	_, err = f.Seek(m.lastPosition, io.SeekStart)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(f)
	var newLines []string
	for scanner.Scan() {
		newLines = append(newLines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	pos, _ := f.Seek(0, io.SeekCurrent)
	m.lastPosition = pos
	return newLines, nil
}

func (m *LogMonitor) Monitor() {

	if f, err := os.Open(m.LogFile); err == nil {
		pos, _ := f.Seek(0, io.SeekEnd)
		m.lastPosition = pos
		f.Close()
	}

	fmt.Println("====== 日志监控启动 ======")
	fmt.Println("文件:", m.LogFile)
	fmt.Println("关键词:", strings.Join(m.Keywords, ", "))
	fmt.Println("=========================")

	ticker := time.NewTicker(time.Duration(m.ScanInterval) * time.Second)
	defer ticker.Stop()
	groupName := ""
	JsonName := ""
	for {
		select {
		case <-ticker.C:
			lines, err := m.scanLog()
			if err != nil {
				fmt.Println("[!] 读取日志错误:", err)
				Notice.SentText(fmt.Sprintf("日志监控服务异常: %v", err))
				return
			}

			startRegexp := regexp.MustCompile(`配置组 "(.*?)" 加载完成`)

			for _, line := range lines {

				matches := startRegexp.FindStringSubmatch(line)
				if matches != nil {
					groupName = matches[1]
				}
				if strings.HasPrefix(line, "→ 开始执行地图追踪任务") {
					JsonName = line
				}

				//关键字告警
				for _, kw := range m.Keywords {
					if strings.Contains(strings.ToLower(line), strings.ToLower(kw)) {
						msg := fmt.Sprintf("⚠️ 日志告警\n\n配置组: %s\n脚本名称: %s\n关键词: %s\n内容: %s", groupName, JsonName, kw, strings.TrimSpace(line))
						//m.sendAlert(msg, false)
						Notice.SentText(msg)
						//fmt.Printf("[%s] 检测到关键词: %s\n", time.Now().Format("2006-01-02 15:04:05"), kw)
						autoLog.Sugar.Infof("[%s] 检测到关键词: %s\n", time.Now().Format("2006-01-02 15:04:05"), kw)
					}
				}
				if strings.Contains(line, "一条龙和配置组任务结束") {
					ArchiveConfig()
					Notice.SentText("一条龙和配置组任务结束，所有配置组已归档")
					autoLog.Sugar.Infof("一条龙和配置组任务结束，所有配置组已归档")
				}
				if strings.Contains(line, "OnRdpClientDisconnected") {
					Notice.SentText("RDP 客户端断开连接")
					autoLog.Sugar.Infof("RDP 客户端断开连接")
					aaa()
				}
				if config.Cfg.ScreenRecord.IsRecord {

					if strings.Contains(line, config.Cfg.ScreenRecord.StartScreen) {

						Notice.SentText("关键词触发录屏 " + config.Cfg.ScreenRecord.StartScreen + "开始录屏")
						Notice.SendScreenshot()
						// 开始录屏
						control.StartRecord()
						autoLog.Sugar.Infof("录屏监控文件 %s", m.LogFile)
						autoLog.Sugar.Infof("关键词触发录屏 【" + config.Cfg.ScreenRecord.StartScreen + "】\n开始录屏")
					}
					if strings.Contains(line, config.Cfg.ScreenRecord.EndScreen) {

						Notice.SentText("关键词触发录屏 " + config.Cfg.ScreenRecord.EndScreen + "结束录屏")
						Notice.SendScreenshot()
						// 开始录屏
						control.StopRecord()
						autoLog.Sugar.Infof("录屏监控文件 %s", m.LogFile)
						autoLog.Sugar.Infof("关键词触发录屏 【" + config.Cfg.ScreenRecord.StartScreen + "】\n结束录屏")
					}

					if strings.Contains(line, config.Cfg.Account.OnlineKeyword) {
						Notice.SentText("联机上线")
						decrypt, err := abgiSSE.Decrypt(config.Cfg.Account.SecretKey, config.Cfg.Account.AccountKey)
						if err != nil {
							autoLog.Sugar.Infof("密钥错误: %v", err)
							Notice.SentText("密钥错误")
						}
						ConnectErr := abgiSSE.Connect(fmt.Sprintf("ws://%s/api/abgiWs/%s/%s", decrypt, config.Cfg.Account.Uid, config.Cfg.Account.Name), nil)
						if ConnectErr != nil {
							autoLog.Sugar.Infof("上线失败: %v", err)
							Notice.SentText("上线失败:" + ConnectErr.Error())
						}
						Notice.SentText("上线成功")
					}

				}
			}

		case <-m.stopChan:
			fmt.Println("[i] 日志监控已退出:", m.LogFile)
			return
		}
	}
}

func (m *LogMonitor) Stop() {
	close(m.stopChan)
}

var (
	user32         = syscall.NewLazyDLL("user32.dll")
	procKeybdEvent = user32.NewProc("keybd_event")
)

const (
	VK_LWIN         = 0x5B // 左 Win 键
	VK_D            = 0x44 // D 键
	KEYEVENTF_KEYUP = 0x0002
)

// 调用 Windows API 模拟键盘事件
func keybdEvent(bVk byte, bScan byte, dwFlags uint32, dwExtraInfo uintptr) {
	procKeybdEvent.Call(
		uintptr(bVk),
		uintptr(bScan),
		uintptr(dwFlags),
		dwExtraInfo,
	)
}

// Win+D 返回桌面
func pressWinD() {
	keybdEvent(VK_LWIN, 0, 0, 0)               // 按下 Win
	keybdEvent(VK_D, 0, 0, 0)                  // 按下 D
	time.Sleep(50 * time.Millisecond)          // 稍微延迟
	keybdEvent(VK_D, 0, KEYEVENTF_KEYUP, 0)    // 松开 D
	keybdEvent(VK_LWIN, 0, KEYEVENTF_KEYUP, 0) // 松开 Win
}

func aaa() {
	autoLog.Sugar.Infof("正在执行会话关闭后操作...")
	time.Sleep(2 * time.Second)

	// 返回 Windows 桌面（Win + D）
	pressWinD()

	time.Sleep(1 * time.Second)

	// 按下 Alt + M
	robotgo.KeyDown("alt")
	robotgo.KeyTap("m")
	time.Sleep(500 * time.Millisecond)
	robotgo.KeyUp("alt")
	robotgo.KeyUp("m")

	time.Sleep(100 * time.Millisecond)

	// 按两次 Enter，间隔 0.3 秒
	robotgo.KeyTap("enter")
	time.Sleep(300 * time.Millisecond)
	robotgo.KeyTap("enter")

	autoLog.Sugar.Infof("操作完成！")
}
