package bgiStatus

import (
	"auto-bgi/config"
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-vgo/robotgo"
	"io"
	"net/http"
	"os"
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
		WebhookURL:   config.Cfg.WebhookURL,
		ScanInterval: interval,
		stopChan:     make(chan struct{}),
	}
}

func (m *LogMonitor) validateConfig() error {
	if _, err := os.Stat(m.LogFile); err != nil {
		return fmt.Errorf("日志文件不存在: %v", err)
	}
	if !(strings.HasPrefix(m.WebhookURL, "http://") || strings.HasPrefix(m.WebhookURL, "https://")) {
		return fmt.Errorf("Webhook URL 格式不正确")
	}
	return nil
}

func (m *LogMonitor) sendAlert(content string, isTest bool) bool {
	prefix := ""
	if isTest {
		prefix = "[TEST] "
	}
	payload := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]interface{}{
			"content":               prefix + content,
			"mentioned_mobile_list": []string{"@all"},
		},
	}
	data, _ := json.Marshal(payload)

	resp, err := http.Post(m.WebhookURL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		fmt.Println("[!] 告警发送失败:", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("[!] 企业微信返回状态码:", resp.StatusCode)
		return false
	}
	return true
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
	if err := m.validateConfig(); err != nil {
		fmt.Println("[!] 配置错误:", err)
		return
	}

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

	for {
		select {
		case <-ticker.C:
			lines, err := m.scanLog()
			if err != nil {
				fmt.Println("[!] 读取日志错误:", err)
				m.sendAlert(fmt.Sprintf("日志监控服务异常: %v", err), false)
				return
			}

			for _, line := range lines {
				for _, kw := range m.Keywords {
					if strings.Contains(strings.ToLower(line), strings.ToLower(kw)) {
						msg := fmt.Sprintf("⚠️ 日志告警\n关键词: %s\n内容: %s", kw, strings.TrimSpace(line))
						m.sendAlert(msg, false)
						fmt.Printf("[%s] 检测到关键词: %s\n", time.Now().Format("2006-01-02 15:04:05"), kw)
					}
				}
				if strings.Contains(line, "一条龙和配置组任务结束") {
					ArchiveConfig()
					m.sendAlert("一条龙和配置组任务结束，所有配置组已归档", false)
				}
				if strings.Contains(line, "OnRdpClientDisconnected") {
					m.sendAlert("RDP 客户端断开连接", false)
					aaa()

				}
			}

		case <-m.stopChan:
			fmt.Println("[i] 日志监控已退出:", m.LogFile)
			return
		}
	}
}

func (m *LogMonitor) ManualTest() {
	msg := fmt.Sprintf("测试告警\n时间: %s\n状态: 监控系统正常", time.Now().Format("2006-01-02 15:04:05"))
	if m.sendAlert(msg, true) {
		fmt.Println("[√] 测试消息发送成功")
	} else {
		fmt.Println("[×] 测试消息发送失败")
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
	fmt.Println("正在执行会话关闭后操作...")
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

	fmt.Println("操作完成！")
}
