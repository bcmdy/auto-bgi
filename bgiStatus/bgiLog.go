package bgiStatus

import (
	"auto-bgi/ArtifactsBulkSupply"
	"auto-bgi/Notice"
	"auto-bgi/abgiObs"
	"auto-bgi/abgiSSE"
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"auto-bgi/control"
	"auto-bgi/task"
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

var YuanShenNum int

var dogFood = ArtifactsBulkSupply.DogFood{}

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

	for {
		select {
		case <-ticker.C:
			lines, err := m.scanLog()
			if err != nil {
				fmt.Println("[!] 读取日志错误:", err)
				Notice.SentText(fmt.Sprintf("日志监控服务异常: %v", err))
				return
			}

			var lastLine string

			for _, line := range lines {
				//关键字告警
				for _, kw := range m.Keywords {
					if strings.Contains(strings.ToLower(line), strings.ToLower(kw)) {
						msg := fmt.Sprintf("⚠️通知：💗%s💗\n配置组：%s\n路线：%s", strings.TrimSpace(line), BgiLogStatusInfo.Group, BgiLogStatusInfo.MapTrackingLine)
						//m.sendAlert(msg, false)
						Notice.SentText(msg)
						//fmt.Printf("[%s] 检测到关键词: %s\n", time.Now().Format("2006-01-02 15:04:05"), kw)
						if config.Cfg.ScreenRecord.IsRecord {
							autoLog.Sugar.Infof("[%s] 检测到关键词: %s\n", time.Now().Format("2006-01-02 15:04:05"), kw)
							go func() {
								abgiObs.SaveReplayBuffer(fmt.Sprintf("%s-%s-%s", BgiLogStatusInfo.Group, BgiLogStatusInfo.MapTrackingLine))
							}()
						}

					}
				}
				//一条龙结束操作
				if strings.Contains(line, "一条龙和配置组任务结束") {
					archiveConfig := ArchiveConfig()
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
						data += fmt.Sprintf("【%s--%s】\n", groupMap.GroupName, executeTime)
						sumExecuteTime += executeTime
					}
					data += fmt.Sprintf("【%s--%s】\n", "合计", sumExecuteTime)
					//Notice.SentText("一条龙和配置组任务结束，所有配置组已归档@所有人")
					Notice.SentText(data)
					autoLog.Sugar.Infof("一条龙和配置组任务结束，所有配置组已归档")
				}
				if strings.Contains(line, "OnRdpClientDisconnected") {
					Notice.SentText("RDP 客户端断开连接")
					autoLog.Sugar.Infof("RDP 客户端断开连接")
					aaa()
				}

				////录屏操作
				//if config.Cfg.ScreenRecord.IsRecord {
				//
				//	if strings.Contains(line, config.Cfg.ScreenRecord.StartScreen) {
				//
				//		Notice.SentText("关键词触发录屏 " + config.Cfg.ScreenRecord.StartScreen + "开始录屏")
				//		Notice.SendScreenshot()
				//		// 开始录屏
				//		control.StartRecord()
				//		autoLog.Sugar.Infof("录屏监控文件 %s", m.LogFile)
				//		autoLog.Sugar.Infof("关键词触发录屏 【" + config.Cfg.ScreenRecord.StartScreen + "】\n开始录屏")
				//	}
				//	if strings.Contains(line, config.Cfg.ScreenRecord.EndScreen) {
				//
				//		Notice.SentText("关键词触发录屏 " + config.Cfg.ScreenRecord.EndScreen + "结束录屏")
				//		Notice.SendScreenshot()
				//		// 开始录屏
				//		control.StopRecord()
				//		autoLog.Sugar.Infof("录屏监控文件 %s", m.LogFile)
				//		autoLog.Sugar.Infof("关键词触发录屏 【" + config.Cfg.ScreenRecord.StartScreen + "】\n结束录屏")
				//	}
				//}

				//关键字触发执行一条龙
				if config.Cfg.Account.OnlineAfterKeyword != "" && strings.Contains(line, config.Cfg.Account.OnlineAfterKeyword) {
					Notice.SentText("关键词触发执行一条龙:" + config.Cfg.Account.OnlineAfterKeyword)
					if config.Cfg.Account.OnlineAfterOneLong == "" {
						Notice.SentText("关键词触发执行一条龙，但是没有配置OnlineAfterOneLong")
					} else {
						task.StartOneDragon(config.Cfg.Account.OnlineAfterOneLong)
					}

				}

				//上线操作
				if config.Cfg.Account.OnlineKeyword != "" && strings.Contains(line, config.Cfg.Account.OnlineKeyword) {

					decrypt, err := abgiSSE.Decrypt(config.Cfg.Account.SecretKey, config.Cfg.Account.AccountKey)
					if err != nil {
						autoLog.Sugar.Infof("密钥错误")
						return
					}

					runDebug := false
					abgiType := "noDebug"
					//查询今天狗粮批发是什么线路
					dogFoodLine := dogFood.DogFoodIsAOrB()
					if dogFoodLine == "" {
						autoLog.Sugar.Errorf("查询今天狗粮批发线路失败")
						runDebug = true
						abgiType = "debug"
					} else if dogFoodLine == "B" {
						runDebug = true
						abgiType = "debug"
					} else if runDebug {
						abgiType = "debug"
					}

					autoLog.Sugar.Infof("当前狗粮批发线路为：%s，是否调试：%t，上线类型：%s", dogFoodLine, runDebug, abgiType)

					ConnectErr := abgiSSE.Connect(fmt.Sprintf("ws://%s/api/abgiWs/%s/%s/%s", decrypt, abgiType, config.Cfg.Account.Uid, config.Cfg.Account.Name), runDebug, nil)
					if ConnectErr != nil {
						autoLog.Sugar.Infof("上线失败")

					}
					autoLog.Sugar.Infof("上线成功")
				}

				//联机狗粮和批发狗粮数量合并
				if strings.Contains(line, "分解可获得经验: ") {
					num := strings.Replace(line, "分解可获得经验: ", "", -1)

					dogFood.WriteDogFoodNum(num)
				}

				//联机狗粮
				if strings.Contains(line, "所有队友已就绪") {
					Notice.SendScreenshot()
				}

				if strings.Contains(line, "如果你已经在游戏内的其他界面，请自行退出当前界面（ESC）") {
					control.PressEsc()
				}

				//首页相关
				//配置组名称
				if strings.Contains(line, "配置组") && strings.Contains(line, "开始执行") {
					re := regexp.MustCompile(`"(.*?)"`)
					matches := re.FindStringSubmatch(line)
					if len(matches) > 1 {
						BgiLogStatusInfo.Group = matches[1]
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

				//查找配置组任务执行
				if strings.Contains(line, "配置组任务执行: ") {
					gp := strings.ReplaceAll(line, "配置组任务执行: ", "")
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
						}
					} else {
						BgiLogStatusInfo.MapTrackingLine = "未找到地图追踪路线"
					}
				}

				//js进度
				if strings.Contains(line, "当前进度：") || strings.Contains(line, "当前次数：") || strings.Contains(line, ": 开始执行") {
					BgiLogStatusInfo.JSProgress = line
				}
				YuanShenNum++
				//原神闪退检测
				if strings.Contains(line, "当前获取焦点的窗口不是原神，尝试恢复窗口") {
					YuanShenNum++
					if YuanShenNum >= 30 {
						YuanShenNum = 0
						Notice.SentText("当前获取焦点的窗口不是原神，尝试恢复窗口超过一分钟")
						autoLog.Sugar.Infof("当前获取焦点的窗口不是原神，尝试恢复窗口超过一分钟")
						//判断原神是否启动
						if !IsWechatRunning("YuanShen.exe") {
							Notice.SentText("原神未启动，尝试启动原神")
							autoLog.Sugar.Infof("原神未启动，尝试启动原神")
							genShinExe := control.ReadGenShinExe()
							control.OpenSoftware(genShinExe)
						}
					}
					//睡眠一秒
					time.Sleep(1 * time.Second)
					autoLog.Sugar.Infof("当前获取焦点的窗口不是原神，尝试恢复窗口%d次", YuanShenNum)
				} else {
					YuanShenNum = 0
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
