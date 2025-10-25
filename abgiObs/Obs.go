package abgiObs

import (
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"fmt"
	"github.com/andreykaipov/goobs"
	"github.com/andreykaipov/goobs/api/requests/outputs"
	"github.com/andreykaipov/goobs/api/requests/record"
	"github.com/andreykaipov/goobs/api/requests/stream"
	"regexp"
	"strings"
	"sync"
	"time"
)

var (
	client   *goobs.Client
	connLock sync.Mutex
	//address  = "localhost:4455" // OBS WebSocket 地址
	//password = "123456"         // OBS WebSocket 密码（如有）
)

// ========== 确保连接 ==========
func EnsureConnected() error {
	connLock.Lock()
	defer connLock.Unlock()

	if client != nil {
		return nil
	}

	var err error
	client, err = goobs.New(config.Cfg.ScreenRecord.StartScreen, goobs.WithPassword(config.Cfg.ScreenRecord.EndScreen))
	if err != nil {
		return fmt.Errorf("连接到 OBS WebSocket 失败: %v", err)
	}

	fmt.Printf("✅ 已连接到 OBS WebSocket (%s)", config.Cfg.ScreenRecord.StartScreen)
	return nil
}

// ==========================================================================
// 🎬【录制控制】
// ==========================================================================

// 开始录制
func StartRecording() error {
	connLock.Lock()
	defer connLock.Unlock()

	if client == nil {
		return fmt.Errorf("OBS客户端未连接")
	}

	_, err := client.Record.StartRecord(&record.StartRecordParams{})
	if err != nil {
		return fmt.Errorf("启动录制失败: %v", err)
	}
	autoLog.Sugar.Infof("🎬 已开始录制")
	return nil
}

// 停止录制
func StopRecording() error {
	connLock.Lock()
	defer connLock.Unlock()

	if client == nil {
		return fmt.Errorf("OBS客户端未连接")
	}

	resp, err := client.Record.StopRecord(&record.StopRecordParams{})
	if err != nil {
		return fmt.Errorf("停止录制失败: %v", err)
	}

	autoLog.Sugar.Infof("🛑 已停止录制，输出文件: %s", resp.OutputPath)
	return nil
}

// 查询录制状态
func GetRecordingStatus() (*record.GetRecordStatusResponse, error) {
	connLock.Lock()
	defer connLock.Unlock()

	if client == nil {
		return nil, fmt.Errorf("OBS客户端未连接")
	}

	status, err := client.Record.GetRecordStatus(&record.GetRecordStatusParams{})
	if err != nil {
		return status, fmt.Errorf("获取录制状态失败: %v", err)
	}

	return status, nil
}

// ==========================================================================
// 🔌【连接管理】
// ==========================================================================

func Disconnect() {
	connLock.Lock()
	defer connLock.Unlock()

	if client != nil {
		client.Disconnect()
		client = nil
		autoLog.Sugar.Infof("🔌 已断开与 OBS 的连接")
	}
}

var (
	lastSentMap = struct {
		sync.Mutex
		m map[string]time.Time
	}{m: make(map[string]time.Time)}

	noticeTTL   = 60 * time.Second
	maxCacheLen = 5
)

func SaveReplayBuffer(fileName string) (*outputs.SaveReplayBufferResponse, error) {

	lastSentMap.Lock()
	defer lastSentMap.Unlock()

	// 如果存在并且 20 秒内，直接忽略
	if t, ok := lastSentMap.m[fileName]; ok {
		if time.Since(t) < noticeTTL {
			autoLog.Sugar.Infof("最近保存过相同文件名，请等待 20 秒后重试：" + fileName)
			return nil, fmt.Errorf("最近已发送过相同文件名，请等待 20 秒后重试")
		}
	}

	// 如果超过最大缓存大小，移除最旧的一条
	if len(lastSentMap.m) >= maxCacheLen {
		var oldestKey string
		var oldestTime time.Time
		for k, v := range lastSentMap.m {
			if oldestTime.IsZero() || v.Before(oldestTime) {
				oldestKey = k
				oldestTime = v
			}
		}
		delete(lastSentMap.m, oldestKey)
	}

	// 更新为当前时间
	lastSentMap.m[fileName] = time.Now()

	//去除名字中的特殊符号
	fileName = SanitizeFileName(fileName)

	if client == nil {
		return nil, fmt.Errorf("OBS客户端未连接")
	}

	connLock.Lock()
	defer connLock.Unlock()

	fmt.Println("开始保存回放缓冲区")

	_, err2 := setOutputSettings(fileName)
	if err2 != nil {
		autoLog.Sugar.Infof("设置输出设置失败: %v", err2)

	}
	autoLog.Sugar.Infof("回放保存成功: %v", fileName)

	//睡眠5秒
	time.Sleep(5 * time.Second)

	res, err := client.Outputs.SaveReplayBuffer(&outputs.SaveReplayBufferParams{})
	if err != nil {
		autoLog.Sugar.Errorf("保存回放缓冲区失败: %v", err)
	}

	return res, nil

}

// SanitizeFileName 去除字符串中影响文件命名的特殊符号
func SanitizeFileName(name string) string {
	// 替换 Windows / Linux / macOS 不允许的字符
	// 包括: \ / : * ? " < > | 和一些控制字符
	reg := regexp.MustCompile(`[<>:"/\\|?*\x00-\x1F]`)
	cleaned := reg.ReplaceAllString(name, "")

	// 去掉首尾空格与点（防止 Windows 文件名非法）
	cleaned = strings.Trim(cleaned, " .")
	cleaned = strings.Replace(cleaned, ".json", "", -1)

	// 防止文件名为空
	if cleaned == "" {
		cleaned = "untitled"
	}

	return cleaned
}

// 开启回放缓存
func StartReplayBuffer() error {

	buffer, err := client.Outputs.StartReplayBuffer(&outputs.StartReplayBufferParams{})
	if err != nil {
		return fmt.Errorf("开启回放缓冲区失败: %v", err)
	}
	autoLog.Sugar.Infof("📼 已开启回放缓冲区，输出文件: %s", buffer)
	return nil
}

// 停止回放缓存
func StopReplayBuffer() error {
	buffer, err := client.Outputs.StopReplayBuffer(&outputs.StopReplayBufferParams{})
	if err != nil {
		return fmt.Errorf("停止回放缓冲区失败: %v", err)
	}
	autoLog.Sugar.Infof("🛑 已停止回放缓冲区，输出文件: %s", buffer)
	return nil
}

// 获取重放缓冲区状态
func GetReplayBufferStatus() (*outputs.GetReplayBufferStatusResponse, error) {
	if client == nil {
		return nil, fmt.Errorf("OBS客户端未连接")
	}

	status, err := client.Outputs.GetReplayBufferStatus(&outputs.GetReplayBufferStatusParams{})
	if err != nil {
		return status, fmt.Errorf("获取回放缓冲区状态失败: %v", err)
	}

	return status, nil
}

// SetOutputSettings 设置回放缓存输出设置
func setOutputSettings(fileName string) (*outputs.SetOutputSettingsResponse, error) {
	if client == nil {
		return nil, fmt.Errorf("OBS客户端未连接")
	}

	s := "回放缓存"

	settings, err := client.Outputs.GetOutputSettings(&outputs.GetOutputSettingsParams{
		OutputName: &s,
	})
	if err != nil {
		return nil, fmt.Errorf("获取输出设置失败: %v", err)
	}
	settings.OutputSettings["format"] = fileName + "%CCYY-%MM-%DD %hh-%mm-%ss"

	outputSettings, err := client.Outputs.SetOutputSettings(&outputs.SetOutputSettingsParams{
		OutputName:     &s,
		OutputSettings: settings.OutputSettings,
	})
	if err != nil {
		return nil, err
	}

	return outputSettings, nil
}

// 启动流
func StartStream() error {
	if client == nil {
		return fmt.Errorf("OBS客户端未连接")
	}

	_, err := client.Stream.StartStream(&stream.StartStreamParams{})
	if err != nil {
		return fmt.Errorf("启动流失败: %v", err)
	}
	autoLog.Sugar.Infof("🔴 已启动流")
	return nil
}
