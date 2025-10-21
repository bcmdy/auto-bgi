package abgiObs

import (
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"fmt"
	"github.com/andreykaipov/goobs"
	"github.com/andreykaipov/goobs/api/requests/outputs"
	"github.com/andreykaipov/goobs/api/requests/record"
	"os"
	"path/filepath"
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

func SaveReplayBuffer(fileName string) (*outputs.SaveReplayBufferResponse, error) {

	if client == nil {
		return nil, fmt.Errorf("OBS客户端未连接")
	}

	connLock.Lock()
	defer connLock.Unlock()

	fmt.Println("开始保存回放缓冲区")

	//睡眠5秒
	time.Sleep(5 * time.Second)

	res, err := client.Outputs.SaveReplayBuffer(&outputs.SaveReplayBufferParams{})
	if err != nil {
		autoLog.Sugar.Errorf("保存回放缓冲区失败: %v", err)
	}

	go func() {
		file, err := renameFile(fileName)
		if err != nil {
			autoLog.Sugar.Errorf("重命名文件失败: %v", err)
		}
		autoLog.Sugar.Infof("📁 已保存回放缓冲区，输出文件: %s", file)
	}()

	return res, nil

}

// 文件重命名
func renameFile(newName string) (string, error) {
	files, err := os.ReadDir(config.Cfg.ScreenRecord.ObsSavePath)
	if err != nil {
		return "", fmt.Errorf("读取目录失败: %v", err)
	}

	var latest os.DirEntry
	var latestTime time.Time

	for _, f := range files {
		name := f.Name()
		if !strings.Contains(name, "Replay_") {
			continue
		}
		ext := filepath.Ext(name)
		t, err := replayTime(strings.TrimSuffix(name, ext))
		if err != nil {
			autoLog.Sugar.Warnf("解析文件时间失败: %s, 错误: %v", name, err)
			continue
		}
		if t.After(latestTime) {
			latest = f
			latestTime = t
		}
	}

	if latest == nil {
		return "", fmt.Errorf("未找到符合条件的回放文件")
	}

	newName = filepath.Clean(newName)
	if strings.ContainsAny(newName, string(os.PathSeparator)) {
		return "", fmt.Errorf("文件名包含非法字符")
	}

	oldPath := filepath.Join(config.Cfg.ScreenRecord.ObsSavePath, latest.Name())
	newPath := filepath.Join(config.Cfg.ScreenRecord.ObsSavePath, newName)
	err = os.Rename(oldPath, newPath)
	if err != nil {
		return "", fmt.Errorf("重命名文件失败: %v", err)
	}

	return newPath, nil
}

// 格式化日期
func replayTime(data string) (time.Time, error) {
	if !strings.HasPrefix(data, "Replay_") {
		return time.Time{}, fmt.Errorf("文件名格式错误，需以 Replay_ 开头: %s", data)
	}

	parts := strings.Split(data, "_")
	if len(parts) != 3 {
		return time.Time{}, fmt.Errorf("文件名格式错误，期望 Replay_YYYY-MM-DD_HH-MM-SS: %s", data)
	}

	datePart := parts[1]
	timePart := strings.ReplaceAll(parts[2], "-", ":")
	if !regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`).MatchString(datePart) {
		return time.Time{}, fmt.Errorf("日期格式错误，期望 YYYY-MM-DD: %s", datePart)
	}
	if !regexp.MustCompile(`^\d{2}:\d{2}:\d{2}$`).MatchString(timePart) {
		return time.Time{}, fmt.Errorf("时间格式错误，期望 HH:MM:SS: %s", timePart)
	}

	full := datePart + " " + timePart
	t, err := time.Parse("2006-01-02 15:04:05", full)
	if err != nil {
		return time.Time{}, fmt.Errorf("解析时间失败: %s, 错误: %v", data, err)
	}

	return t, nil
}
