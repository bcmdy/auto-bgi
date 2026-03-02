package bgiStatus

import (
	"auto-bgi/Notice"
	"auto-bgi/abgiObs"
	"auto-bgi/abgiSSE"
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"auto-bgi/control"
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/go-vgo/robotgo"
)

var AudioService control.Audio

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

	info, err := f.Stat()
	if err == nil && m.lastPosition > info.Size() {
		m.lastPosition = 0
	}

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
	autoLog.Sugar.Infof("当前监控的日志文件:%s", m.LogFile)
	fmt.Println("关键词:", strings.Join(m.Keywords, ", "))
	fmt.Println("=========================")

	ticker := time.NewTicker(time.Duration(m.ScanInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			lines, err := m.scanLog()
			if err != nil {
				if os.IsNotExist(err) {
					logDir := filepath.Dir(m.LogFile)
					files, ferr := findLogFiles(logDir)
					if ferr != nil || len(files) == 0 {
						fmt.Println("日志文件被删除，且未找到新的日志文件")
						continue
					}

					newLogFile := filepath.Join(logDir, files[0])
					fmt.Println("日志文件被删除，切换到新的日志文件:", newLogFile)

					m.LogFile = newLogFile

					if f2, err2 := os.Open(newLogFile); err2 == nil {
						pos2, _ := f2.Seek(0, io.SeekEnd)
						m.lastPosition = pos2
						f2.Close()
					} else {
						m.lastPosition = 0
					}

					currentLogFile = newLogFile
					go func() {
						config.Cfg.BgiLog = newLogFile
						InitBgiLogStatus()
					}()

					fmt.Println("====== 日志监控切换 ======")
					fmt.Println("文件:", m.LogFile)
					autoLog.Sugar.Infof("正在监控的文件:%s", m.LogFile)
					fmt.Println("关键词:", strings.Join(m.Keywords, ", "))
					fmt.Println("=========================")

					continue
				}

				fmt.Println("[!] 读取日志错误:", err)
				Notice.SentText(fmt.Sprintf("日志监控服务异常: %v", err))
				return
			}

			if len(lines) == 0 {
				continue
			}

			var lastLine string

			for _, line := range lines {
				//关键字告警
				for _, kw := range m.Keywords {
					if strings.Contains(strings.ToLower(line), strings.ToLower(kw)) {
						msg := fmt.Sprintf("⚠️通知：💗%s💗\n时间:[%s]\n配置组：%s\n路线：%s", strings.TrimSpace(line), time.Now().Format("2006-01-02 15:04:05"), BgiLogStatusInfo.Group, BgiLogStatusInfo.MapTrackingLine)
						//m.sendAlert(msg, false)
						Notice.SentText(msg)
						//fmt.Printf("[%s] 检测到关键词: %s\n", time.Now().Format("2006-01-02 15:04:05"), kw)
						if config.Cfg.ScreenRecord.IsRecord {
							autoLog.Sugar.Infof("[%s] 检测到关键词: %s\n", time.Now().Format("2006-01-02 15:04:05"), kw)
							go func() {
								abgiObs.SaveReplayBuffer(fmt.Sprintf("%s-%s-%s", BgiLogStatusInfo.Group, BgiLogStatusInfo.MapTrackingLine, time.Now().Format("2006-01-02-15-04-05")))
							}()
						}

					}
				}

				if strings.HasPrefix(line, "ABGI") {
					JsLogHandler(line)
				}

				//角色死亡自动吃药
				if strings.Contains(line, "发现角色死亡") && config.Cfg.Control.IsRedBlood {
					//自动吃药
					go func() {
						//等待2秒自动吃药
						time.Sleep(2 * time.Second)
						for range config.Cfg.Control.RedBloodCount {
							control.PressKey("z")
							CheckRedBloodSum++
							Notice.SentText(fmt.Sprintf("角色死亡，自动按Z键，第%d次", CheckRedBloodSum))
							autoLog.Sugar.Infof("角色死亡 自动按Z键，第%d次", CheckRedBloodSum)
							time.Sleep(3 * time.Second)
						}
					}()
				}

				//一条龙结束操作
				if strings.Contains(line, "一条龙和配置组任务结束") {

					archiveConfig, ArchiveRecordMap := ArchiveConfig()
					data := "一条龙和配置组任务结束，所有配置组已归档@所有人\n"
					sumExecuteTime, _ := time.ParseDuration("0s")
					for _, groupMap := range archiveConfig {

						executeTime, _ := time.ParseDuration("0s")
						for _, segment := range groupMap.Segments {
							if segment.Consuming != "" {
								duration, err := time.ParseDuration(segment.Consuming)
								if err != nil {
									autoLog.Sugar.Errorf("解析时间失败: %v", err)
									continue
								}
								executeTime += duration

							}

						}

						//计算差值
						var diff time.Duration
						if archiveRecord, ok := ArchiveRecordMap[groupMap.GroupName]; ok {
							duration, err := time.ParseDuration(archiveRecord)
							if err != nil {
								fmt.Println("解析错误:", err)
								return
							}
							diff = executeTime - duration
						}

						data += fmt.Sprintf("【%s--%s】(%s)\n", groupMap.GroupName, executeTime, diff)
						sumExecuteTime += executeTime
					}
					data += fmt.Sprintf("【%s--%s】\n", "合计", sumExecuteTime)

					Notice.SentText(data)

					autoLog.Sugar.Infof("一条龙和配置组任务结束，所有配置组已归档")
				}

				if strings.Contains(line, "OnRdpClientDisconnected") {
					Notice.SentText("RDP 客户端断开连接")
					autoLog.Sugar.Infof("RDP 客户端断开连接")
					aaa()
				}

				if strings.HasPrefix(line, "通知发送成功：") && strings.Contains(line, "联机狗粮分解获得经验") {
					//提取数字
					re := regexp.MustCompile(`(\d+)$`)
					num := re.FindString(line)

					abgiSSE.WriteDogFoodNum(num)
					abgiSSE.ABgiSeeStatus = "联机结束"
				}

				if strings.Contains(line, "如果你已经在游戏内的其他界面，请自行退出当前界面（ESC）") {
					Notice.SentText("如果你已经在游戏内的其他界面，请自行退出当前界面（ESC）")
					autoLog.Sugar.Infof("如果你已经在游戏内的其他界面，请自行退出当前界面（ESC）")
					Notice.SendScreenshot()
					control.PressEsc()
				}

				//首页相关

				//一条龙初始化
				if strings.Contains(line, "参数指定的一条龙配置：") {
					name := strings.ReplaceAll(line, "参数指定的一条龙配置：", "")
					//去除空格
					name = strings.TrimSpace(name)
					InitialOneLongProgress(name)
				}

				if BgiGroupEnd[line] != "" {
					if _, ok := OneLongProgress.Details[BgiGroupEnd[line]]; ok {
						OneLongProgress.Details[BgiGroupEnd[line]] = true
					}
				}

				//配置组名称
				if strings.Contains(line, "配置组") && strings.Contains(line, "开始执行") {
					re := regexp.MustCompile(`"(.*?)"`)
					matches := re.FindStringSubmatch(line)
					if len(matches) > 1 {
						BgiLogStatusInfo.Group = matches[1]
						autoLog.Sugar.Infof("配置组：【%s】开始执行", matches[1])
						//提取开始时间
						BgiLogStatusInfo.Timestamp = BgiLogTime(lastLine)

						BgiLogStatusInfo.MapTrackingLine = ""
						BgiLogStatusInfo.ScriptName = ""
						BgiLogStatusInfo.JSProgress = ""

						readConfig := Group.ReadConfig(BgiLogStatusInfo.Group)
						Projects = readConfig.Projects
						BgiLogStatusInfo.ConfigurationGroupExecutionProgress = fmt.Sprintf("%d/%d", 0, len(Projects))

					} else {
						BgiLogStatusInfo.Group = "未找到配置组"
					}
				}

				if strings.Contains(line, "配置组") && strings.Contains(line, "执行结束") {
					re := regexp.MustCompile(`"(.*?)"`)
					matches := re.FindStringSubmatch(line)
					if len(matches) > 1 {
						BgiLogStatusInfo.Group = matches[1]
						autoLog.Sugar.Infof("配置组：【%s】结束执行", matches[1])
						BgiLogStatusInfo.MapTrackingLine = "已经结束"
						BgiLogStatusInfo.ScriptName = "已经结束"
						BgiLogStatusInfo.JSProgress = "已经结束"
						BgiLogStatusInfo.ConfigurationGroupExecutionProgress = "已经结束"

						//更新进度
						if _, ok := OneLongProgress.Details[BgiLogStatusInfo.Group]; ok {
							OneLongProgress.Details[BgiLogStatusInfo.Group] = true
						}

					} else {
						BgiLogStatusInfo.Group = "未找到配置组"
					}
				}

				//查找配置组任务执行
				if strings.Contains(line, "配置组任务执行: ") {
					gp := strings.ReplaceAll(line, "配置组任务执行: ", "")
					autoLog.Sugar.Infof("一条龙进度：【%s】", gp)
					BgiLogStatusInfo.GroupProgress = gp
				}

				if strings.HasPrefix(line, "→ 开始执行JS脚本: ") {
					re := regexp.MustCompile(`"(.*?)"`)
					matches := re.FindStringSubmatch(line)
					if len(matches) > 1 {
						BgiLogStatusInfo.ScriptName = matches[1]
					} else {
						BgiLogStatusInfo.ScriptName = "未找到脚本名称"
					}
					index := GetProjectIndex(BgiLogStatusInfo.ScriptName)
					BgiLogStatusInfo.ConfigurationGroupExecutionProgress = fmt.Sprintf("%d/%d", index, len(Projects))
					autoLog.Sugar.Infof("配置组执行进度:%s", fmt.Sprintf("%d/%d", index, len(Projects)))
				}

				//当前运行路线
				if strings.Contains(line, "开始执行地图追踪任务") {
					re := regexp.MustCompile(`"(.*?)"`)
					matches := re.FindStringSubmatch(line)
					if len(matches) > 1 {
						BgiLogStatusInfo.MapTrackingLine = matches[1]
						index := GetProjectIndex(BgiLogStatusInfo.MapTrackingLine)
						if index != 0 {
							BgiLogStatusInfo.ConfigurationGroupExecutionProgress = fmt.Sprintf("%d/%d", index, len(Projects))
							autoLog.Sugar.Infof("配置组执行进度:%s", fmt.Sprintf("%d/%d", index, len(Projects)))

						}
					} else {
						BgiLogStatusInfo.MapTrackingLine = "未找到地图追踪路线"
					}
				}

				//js进度
				if strings.Contains(line, "当前进度") || strings.Contains(line, ": 开始执行") {
					BgiLogStatusInfo.JSProgress = line
					autoLog.Sugar.Infof("js进度:%s", line)
				}

				//原神闪退检测
				if strings.Contains(line, "当前获取焦点的窗口不是原神，尝试恢复窗口") {
					Notice.SentText("当前获取焦点的窗口不是原神，尝试恢复窗口")
					autoLog.Sugar.Infof("当前获取焦点的窗口不是原神，尝试恢复窗口")

				}

				lastLine = line
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

// 读取文件
func GetLogInfo(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}

	return strings.Join(lines, "\n"), nil

}
