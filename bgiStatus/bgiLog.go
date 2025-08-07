package bgiStatus

import (
	"auto-bgi/config"
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

type LogMonitor struct {
	LogFile      string
	Keywords     []string
	WebhookURL   string
	ScanInterval int
	lastPosition int64
	stopChan     chan struct{} // ✅ 新增: 停止信号
}

// 初始化监控器，复用全局配置
func NewLogMonitor(logFile string, keywords []string, interval int) *LogMonitor {
	return &LogMonitor{
		LogFile:    logFile,
		Keywords:   keywords,
		WebhookURL: config.Cfg.WebhookURL,
		stopChan:   make(chan struct{}), // ✅ 初始化通道
	}
}

// 验证配置
func (m *LogMonitor) validateConfig() error {
	if _, err := os.Stat(m.LogFile); err != nil {
		return fmt.Errorf("日志文件不存在: %v", err)
	}
	if !(strings.HasPrefix(m.WebhookURL, "http://") || strings.HasPrefix(m.WebhookURL, "https://")) {
		return fmt.Errorf("Webhook URL 格式不正确")
	}
	return nil
}

// 发送企业微信告警
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

// 扫描新增日志
func (m *LogMonitor) scanLog() ([]string, error) {
	f, err := os.Open(m.LogFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	_, _ = f.Seek(m.lastPosition, io.SeekStart)
	reader := bufio.NewReader(f)
	var newLines []string

	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			newLines = append(newLines, line)
		}
		if err != nil {
			break
		}
	}
	pos, _ := f.Seek(0, io.SeekCurrent)
	m.lastPosition = pos
	return newLines, nil
}

// 监控主循环
func (m *LogMonitor) Monitor() {
	if err := m.validateConfig(); err != nil {
		fmt.Println("[!] 配置错误:", err)
		return
	}

	// 初始化偏移量
	if f, err := os.Open(m.LogFile); err == nil {
		pos, _ := f.Seek(0, io.SeekEnd)
		m.lastPosition = pos
		f.Close()
	}

	fmt.Println("====== 日志监控启动 ======")
	fmt.Println("文件:", m.LogFile)
	fmt.Println("关键词:", strings.Join(m.Keywords, ", "))
	fmt.Println("=========================")

	for {
		lines, err := m.scanLog()
		if err != nil {
			fmt.Println("[!] 读取日志错误:", err)
			m.sendAlert(fmt.Sprintf("日志监控服务异常: %v", err), false)
			break
		}

		for _, line := range lines {
			for _, kw := range m.Keywords {
				if matched, _ := regexp.MatchString("(?i)"+kw, line); matched {
					msg := fmt.Sprintf("⚠️ 日志告警\n关键词: %s\n内容: %s", kw, strings.TrimSpace(line))
					m.sendAlert(msg, false)
					fmt.Printf("[%s] 检测到关键词: %s\n", time.Now().Format("2006-01-02 15:04:05"), kw)
				}
				if line == "一条龙和配置组任务结束" {
					ArchiveConfig()
					m.sendAlert("一条龙和配置组任务结束，所有配置组已归档", false)
				}
			}
		}
		time.Sleep(time.Duration(m.ScanInterval) * time.Second)
	}
}

// 手动测试
func (m *LogMonitor) ManualTest() {
	msg := fmt.Sprintf("测试告警\n时间: %s\n状态: 监控系统正常", time.Now().Format("2006-01-02 15:04:05"))
	if m.sendAlert(msg, true) {
		fmt.Println("[√] 测试消息发送成功")
	} else {
		fmt.Println("[×] 测试消息发送失败")
	}
}

// ✅ 新增: 停止监控
func (m *LogMonitor) Stop() {
	close(m.stopChan) // 关闭通道以通知退出
	fmt.Println("[i] 日志监控已停止:", m.LogFile)
}
