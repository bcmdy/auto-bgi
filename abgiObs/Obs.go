package abgiObs

import (
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"fmt"
	"github.com/andreykaipov/goobs"
	"github.com/andreykaipov/goobs/api/requests/outputs"
	"github.com/andreykaipov/goobs/api/requests/record"
	"github.com/andreykaipov/goobs/api/requests/stream"
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
)

// ========== 确保连接 ==========
func EnsureConnected() error {

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
		autoLog.Sugar.Error("OBS客户端未连接,准备重连")
		err := EnsureConnected()
		if err != nil {
			return fmt.Errorf("重连还是不行,OBS客户端未连接:%s", err)
		}
	}

	_, err := client.Record.StartRecord(&record.StartRecordParams{})
	if err != nil {
		return fmt.Errorf("启动录制失败: %v", err)
	}

	autoLog.Sugar.Infof("🎬 已开始录制")
	return nil
}

func updateFileName(OutputPath, videoName string) {
	if videoName == "" {
		return
	}

	videoName = filepath.Join(config.Cfg.ScreenRecord.ObsSavePath, videoName+time.Now().Format("2006-01-02-15-01")+".mkv")

	if OutputPath == "" {
		return
	}

	fileName := filepath.Base(OutputPath)
	fileName = SanitizeFileName(fileName)

	// 等待文件可用
	for i := 0; i < 5; i++ {
		err := os.Rename(OutputPath, videoName)
		if err == nil {
			autoLog.Sugar.Infof("重命名成功：%s -> %s", fileName, videoName)
			return
		}

		if os.IsPermission(err) || strings.Contains(err.Error(), "being used") {
			time.Sleep(2 * time.Second)
			continue
		}

		autoLog.Sugar.Infof("重命名失败：" + err.Error())
		return
	}

	autoLog.Sugar.Infof("重命名失败：文件始终被占用")
}

// 停止录制
func StopRecording(videoName string) error {
	connLock.Lock()
	defer connLock.Unlock()

	if client == nil {
		autoLog.Sugar.Error("OBS客户端未连接,准备重连")
		err := EnsureConnected()
		if err != nil {
			return fmt.Errorf("重连还是不行,OBS客户端未连接:%s", err)
		}
	}

	resp, err := client.Record.StopRecord(&record.StopRecordParams{})
	if err != nil {
		return fmt.Errorf("停止录制失败: %v", err)
	}

	go func() {
		updateFileName(resp.OutputPath, videoName)
	}()

	autoLog.Sugar.Infof("🛑 已停止录制，输出文件: %s", resp.OutputPath)
	return nil
}

// 查询录制状态
func GetRecordingStatus() (*record.GetRecordStatusResponse, error) {
	connLock.Lock()
	defer connLock.Unlock()

	if client == nil {
		autoLog.Sugar.Error("OBS客户端未连接,准备重连")
		err := EnsureConnected()
		if err != nil {
			return nil, fmt.Errorf("重连还是不行,OBS客户端未连接:%s", err)
		}
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

	//捕获错误
	defer func() {
		if r := recover(); r != nil {
			autoLog.Sugar.Errorf("⚠️ 停止回放缓冲区失败: %v", r)

		}
	}()

	if client == nil {
		autoLog.Sugar.Error("OBS客户端未连接,准备重连")
		err := EnsureConnected()
		if err != nil {
			return nil, fmt.Errorf("重连还是不行,OBS客户端未连接:%s", err)
		}
	}

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

	_, err2 := setOutputSettings("回放缓存", fileName)
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
func StartReplayBuffer() (err error) {

	if client == nil {
		autoLog.Sugar.Error("OBS客户端未连接,准备重连")
		err := EnsureConnected()
		if err != nil {
			return fmt.Errorf("重连还是不行,OBS客户端未连接:%s", err)
		}
	}

	defer func() {
		if r := recover(); r != nil {
			autoLog.Sugar.Errorf("⚠️ 捕获到异常: %v", r)
			err = fmt.Errorf("启动回放缓冲区时发生未知错误: %v", r)
		}
	}()

	buffer, err := client.Outputs.StartReplayBuffer(&outputs.StartReplayBufferParams{})
	if err != nil {
		return fmt.Errorf("开启回放缓冲区失败: %v", err)
	}
	autoLog.Sugar.Infof("📼 已开启回放缓冲区，输出文件: %s", buffer)
	return nil
}

// 停止回放缓存
func StopReplayBuffer() error {
	//捕获错误
	defer func() {
		if r := recover(); r != nil {
			autoLog.Sugar.Errorf("⚠️ 停止回放缓冲区失败: %v", r)

		}
	}()

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
		autoLog.Sugar.Error("OBS客户端未连接,准备重连")
		err := EnsureConnected()
		if err != nil {
			return nil, fmt.Errorf("重连还是不行,OBS客户端未连接:%s", err)
		}
	}

	status, err := client.Outputs.GetReplayBufferStatus(&outputs.GetReplayBufferStatusParams{})
	if err != nil {
		return status, fmt.Errorf("获取回放缓冲区状态失败: %v", err)
	}

	return status, nil
}

// SetOutputSettings 设置回放缓存输出设置
func setOutputSettings(outName, fileName string) (*outputs.SetOutputSettingsResponse, error) {
	if client == nil {
		autoLog.Sugar.Error("OBS客户端未连接,准备重连")
		err := EnsureConnected()
		if err != nil {
			return nil, fmt.Errorf("重连还是不行,OBS客户端未连接:%s", err)
		}
	}

	list, _ := client.Outputs.GetOutputList(&outputs.GetOutputListParams{})
	for _, output := range list.Outputs {
		fmt.Println(output)
	}

	settings, err := client.Outputs.GetOutputSettings(&outputs.GetOutputSettingsParams{
		OutputName: &outName,
	})
	if err != nil {
		return nil, fmt.Errorf("获取输出设置失败: %v", err)
	}
	settings.OutputSettings["format"] = fileName + "%CCYY-%MM-%DD %hh-%mm-%ss"

	outputSettings, err := client.Outputs.SetOutputSettings(&outputs.SetOutputSettingsParams{
		OutputName:     &outName,
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
		autoLog.Sugar.Error("OBS客户端未连接,准备重连")
		err := EnsureConnected()
		if err != nil {
			return fmt.Errorf("重连还是不行,OBS客户端未连接:%s", err)
		}
	}

	_, err := client.Stream.StartStream(&stream.StartStreamParams{})
	if err != nil {
		return fmt.Errorf("启动流失败: %v", err)
	}
	autoLog.Sugar.Infof("🔴 已启动流")
	return nil
}

//-portable --disable-shutdown-check

func Shutdown() {
	if client == nil {
		return
	}
	// 停止录制
	_, err := client.Record.StopRecord(&record.StopRecordParams{})
	if err != nil {
		autoLog.Sugar.Errorf("停止录制失败: %v", err)
	}
	// 停止流
	_, err = client.Stream.StopStream(&stream.StopStreamParams{})
	if err != nil {
		autoLog.Sugar.Errorf("停止流失败: %v", err)
	}

	//关闭回放缓存
	_, err = client.Outputs.StopReplayBuffer(&outputs.StopReplayBufferParams{})
	if err != nil {
		autoLog.Sugar.Errorf("关闭回放缓冲区失败: %v", err)
	}

	time.Sleep(5 * time.Second)

	err = client.Disconnect()
	if err != nil {
		autoLog.Sugar.Errorf("关闭obs失败: %v", err)
	}
	autoLog.Sugar.Infof("obs关闭成功")

}

// days: 0 代表删除所有视频，大于 0 代表删除修改时间在 N 天前的文件
func DeleteVideosByAge(days int) error {
	// 1. 定义视频常见后缀
	videoExtensions := map[string]bool{
		".mp4": true,
		".mkv": true,
		".avi": true,
		".mov": true,
		".flv": true,
		".wmv": true,
	}

	// 2. 读取目录
	files, err := os.ReadDir(config.Cfg.ScreenRecord.ObsSavePath)
	if err != nil {
		return fmt.Errorf("无法读取目录: %v", err)
	}

	now := time.Now()
	deletedCount := 0

	for _, file := range files {
		// 跳过目录
		if file.IsDir() {
			continue
		}

		// 3. 检查文件后缀
		ext := strings.ToLower(filepath.Ext(file.Name()))
		if !videoExtensions[ext] {
			continue
		}

		// 4. 检查修改时间
		info, err := file.Info()
		if err != nil {
			autoLog.Sugar.Errorf("无法获取文件信息 [%s]: %v\n", file.Name(), err)
			continue
		}

		// 计算逻辑：如果 days 为 0，或者修改时间早于截止时间
		modTime := info.ModTime()
		cutoff := now.AddDate(0, 0, -days)

		if days == 0 || modTime.Before(cutoff) {
			fullPath := filepath.Join(config.Cfg.ScreenRecord.ObsSavePath, file.Name())

			// 执行删除
			err := os.Remove(fullPath)
			if err != nil {
				autoLog.Sugar.Errorf("删除失败 [%s]: %v\n", file.Name(), err)
			} else {
				autoLog.Sugar.Infof("已删除: %s (修改日期: %s)\n", file.Name(), modTime.Format("2006-01-02"))
				deletedCount++
			}
		}
	}
	autoLog.Sugar.Infof("处理完成，共删除文件数: %d\n", deletedCount)
	return nil
}
