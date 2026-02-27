package bgiStatus

import (
	"auto-bgi/abgiConstant"
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"auto-bgi/control"
	"auto-bgi/tools"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	abgiCopy "github.com/otiai10/copy"
)

func GetVersion() (string, error) {

	// 创建请求
	unix := time.Now().Unix()
	req, err := http.NewRequest("GET", abgiConstant.BgiRunForVersion+strconv.FormatInt(unix, 10), nil)
	if err != nil {
		panic(err)
	}

	// 添加请求头，比如 User-Agent
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Cache-Control", "no-cache")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		// 读取简短的 body 用于诊断（限长）
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return "", fmt.Errorf("http status %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}

	// 读取为纯文本
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	release := strings.TrimSpace(string(b))

	return release, nil
}

// 获取公版最新版本号
type VersionInfo struct {
	Version            string
	CreatedAt          string
	BrowserDownloadUrl string
	Size               int
}

func GetYDVersionInfo() VersionInfo {
	versionInfo, status, err := abgiConstant.GetHttp[VersionInfo]("/api/bgi/GetVersionInfo")
	if err != nil {
		fmt.Printf("获取版本信息失败: %v\n", err)
	} else {
		fmt.Printf("状态码: %d\n", status)
		fmt.Printf("版本信息: %+v\n", versionInfo)
	}

	return versionInfo
}

// 上传bgi压缩包
func UploadBgi(c *gin.Context) {

	control.CloseSoftware()

	// 单文件上传
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 保存文件到指定目录
	dst := filepath.Join("./uploads", "BetterGI.zip")
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	time.Sleep(1 * time.Second)

	err = UpdateBgi()
	if err != nil {
		autoLog.Sugar.Error("更新失败:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message":  err,
			"filename": file.Filename,
			"size":     file.Size,
		})
		return
	}

	//更新仓库
	go func() {
		GitPull()
	}()

	c.JSON(http.StatusOK, gin.H{
		"message":  "更新成功",
		"filename": file.Filename,
		"size":     file.Size,
	})
}

type BgiDownloadStatus struct {
	Status  string
	Percent float64
	Current int64
	Total   int64
	Error   string
}

var (
	downloadLock   sync.Mutex
	lastInvokeTime time.Time
	invokeInterval = 5 * time.Minute

	bgiDownloadStatusMu sync.RWMutex
	bgiDownloadStatus   = BgiDownloadStatus{Status: "idle"}
)

func DownloadBgiProgress(c *gin.Context) {
	if !strings.Contains(config.BgiCfg.RunForVersion, "lcb") {
		autoLog.Sugar.Error("使用的是公版")

	} else {
		autoLog.Sugar.Error("使用的是茶包版")
	}

	bgiDownloadStatusMu.RLock()
	currentStatus := bgiDownloadStatus
	bgiDownloadStatusMu.RUnlock()

	if currentStatus.Status != "downloading" {
		downloadLock.Lock()
		if time.Since(lastInvokeTime) < invokeInterval && !lastInvokeTime.IsZero() {
			downloadLock.Unlock()
			autoLog.Sugar.Warn("操作过于频繁，请等待2分钟后再试")
			c.JSON(http.StatusTooManyRequests, gin.H{
				"message": "操作过于频繁，请等待2分钟后再试",
			})
			return
		}
		lastInvokeTime = time.Now()
		downloadLock.Unlock()

		version, err := GetVersion()
		if err != nil {
			autoLog.Sugar.Error("获取版本失败:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		autoLog.Sugar.Infof("当前BGI版本: %s", config.BgiCfg.RunForVersion)
		autoLog.Sugar.Infof("最新BGI版本: %s", version)

		if version == config.BgiCfg.RunForVersion {
			c.JSON(http.StatusOK, gin.H{
				"message": "当前版本已经是最新版本",
			})
			return
		}

		bgiDownloadStatusMu.Lock()
		bgiDownloadStatus.Status = "downloading"
		bgiDownloadStatus.Percent = 0
		bgiDownloadStatus.Current = 0
		bgiDownloadStatus.Total = 0
		bgiDownloadStatus.Error = ""
		bgiDownloadStatusMu.Unlock()

		go runBgiDownload(version)

	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming unsupported"})
		return
	}

	ctx := c.Request.Context()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	sentDone := false

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			bgiDownloadStatusMu.RLock()
			status := bgiDownloadStatus
			bgiDownloadStatusMu.RUnlock()

			if status.Status == "idle" {
				return
			}

			c.SSEvent("progress", gin.H{
				"percent": fmt.Sprintf("%.2f", status.Percent),
				"current": status.Current,
				"total":   status.Total,
				"status":  status.Status,
				"error":   status.Error,
			})
			flusher.Flush()

			if !sentDone && (status.Status == "done" || status.Status == "error") {
				sentDone = true
				if status.Status == "done" {
					c.SSEvent("done", "success")
				} else {
					c.SSEvent("error", status.Error)
				}
				flusher.Flush()
				return
			}
		}
	}

}

func GetBgiDownloadStatus(c *gin.Context) {
	bgiDownloadStatusMu.RLock()
	status := bgiDownloadStatus
	bgiDownloadStatusMu.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"status":  status.Status,
		"percent": fmt.Sprintf("%.2f", status.Percent),
		"current": status.Current,
		"total":   status.Total,
		"error":   status.Error,
	})
}

func StartDownloadBgi(c *gin.Context) {
	if !strings.Contains(config.BgiCfg.RunForVersion, "lcb") {
		autoLog.Sugar.Error("使用的是公版")

	} else {
		autoLog.Sugar.Error("使用的是茶包版")
	}

	bgiDownloadStatusMu.RLock()
	currentStatus := bgiDownloadStatus
	bgiDownloadStatusMu.RUnlock()

	if currentStatus.Status == "downloading" {
		c.JSON(http.StatusOK, gin.H{
			"message": "下载任务已在进行中",
		})
		return
	}

	downloadLock.Lock()
	if time.Since(lastInvokeTime) < invokeInterval && !lastInvokeTime.IsZero() {
		downloadLock.Unlock()
		autoLog.Sugar.Warn("操作过于频繁，请等待2分钟后再试")
		c.JSON(http.StatusTooManyRequests, gin.H{
			"message": "操作过于频繁，请等待2分钟后再试",
		})
		return
	}
	lastInvokeTime = time.Now()
	downloadLock.Unlock()

	version := ""
	err := error(nil)

	if !strings.Contains(config.BgiCfg.RunForVersion, "lcb") {
		autoLog.Sugar.Error("使用的是公版")
		ydVersionInfo := GetYDVersionInfo()
		version = ydVersionInfo.Version
	} else {
		version, err = GetVersion()
		autoLog.Sugar.Error("使用的是茶包版")
	}

	if err != nil {
		autoLog.Sugar.Error("获取版本失败:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	autoLog.Sugar.Infof("当前BGI版本: %s", config.BgiCfg.RunForVersion)
	autoLog.Sugar.Infof("最新BGI版本: %s", version)

	if version == config.BgiCfg.RunForVersion {
		c.JSON(http.StatusOK, gin.H{
			"message": "当前版本已经是最新版本",
		})
		return
	}

	bgiDownloadStatusMu.Lock()
	bgiDownloadStatus.Status = "downloading"
	bgiDownloadStatus.Percent = 0
	bgiDownloadStatus.Current = 0
	bgiDownloadStatus.Total = 0
	bgiDownloadStatus.Error = ""
	bgiDownloadStatusMu.Unlock()

	go runBgiDownload(version)

	c.JSON(http.StatusOK, gin.H{
		"message": "下载任务已启动",
	})
}

func runBgiDownload(version string) {
	defer func() {
		if r := recover(); r != nil {
			autoLog.Sugar.Errorf("runBgiDownload panic: %v", r)
			bgiDownloadStatusMu.Lock()
			bgiDownloadStatus.Status = "error"
			bgiDownloadStatus.Error = fmt.Sprintf("%v", r)
			bgiDownloadStatusMu.Unlock()
		}
	}()

	control.CloseSoftware()

	dst := filepath.Join("./uploads", "BetterGI.zip")
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Minute)
	defer cancel()

	err := downloadFileWithProgress(ctx, GetTeaBaoUpdateUrl(), dst, func(current, total int64) {

		var percent float64
		if total > 0 {
			percent = float64(current) / float64(total) * 100
		}

		bgiDownloadStatusMu.Lock()
		bgiDownloadStatus.Current = current
		bgiDownloadStatus.Total = total
		bgiDownloadStatus.Percent = percent
		bgiDownloadStatusMu.Unlock()
	})

	if err != nil {
		autoLog.Sugar.Error("下载BGI失败:", err)
		bgiDownloadStatusMu.Lock()
		bgiDownloadStatus.Status = "error"
		bgiDownloadStatus.Error = err.Error()
		bgiDownloadStatusMu.Unlock()
		return
	}

	time.Sleep(1 * time.Second)

	err = UpdateBgi()
	if err != nil {
		autoLog.Sugar.Error("更新失败:", err)
		bgiDownloadStatusMu.Lock()
		bgiDownloadStatus.Status = "error"
		bgiDownloadStatus.Error = err.Error()
		bgiDownloadStatusMu.Unlock()
		return
	}

	go func() {
		GitPull()
	}()

	config.BgiCfg.RunForVersion = version
	control.OpenSoftware(config.Cfg.BetterGIAddress + "\\BetterGI.exe")

	autoLog.Sugar.Infof("更新成功,请自启bgi,更新版本")
	control.GetMachineFingerprint("茶包bgi", version)

	bgiDownloadStatusMu.Lock()
	bgiDownloadStatus.Status = "done"
	bgiDownloadStatus.Percent = 100
	bgiDownloadStatusMu.Unlock()

	//重启aBgi
	time.Sleep(10 * time.Second)
	autoLog.Sugar.Infof("准备重启abgi")
	if err := tools.RestartProgram(); err != nil {
		autoLog.Sugar.Errorf("重启程序失败: %v", err)
		autoLog.Sugar.Error(err.Error())
	}

}

// 下载完成后发送成功事件
// ProgressFunc 用于反馈下载进度
// current: 已下载字节数, total: 总字节数
type ProgressFunc func(current, total int64)

// downloadFileWithProgress 支持分片下载并实时反馈进度
func downloadFileWithProgress(ctx context.Context, reqURL string, dst string, onProgress ProgressFunc) error {
	// 1. 验证 URL
	parsed, err := url.ParseRequestURI(reqURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return errors.New("invalid http/https url")
	}

	var contentLength int64

	// 如果版本号包含 "lcb"，直接使用大小
	if !strings.Contains(strings.ToLower(config.BgiCfg.RunForVersion), "lcb") {
		contentLength = int64(GetYDVersionInfo().Size)
	} else {
		// 2. 获取文件总大小
		headReq, err := http.NewRequestWithContext(ctx, http.MethodHead, reqURL, nil)
		if err != nil {
			return err
		}
		resp, err := http.DefaultClient.Do(headReq)
		if err != nil {
			return fmt.Errorf("head request failed: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("server returned status: %s", resp.Status)
		}

		contentLength, _ = strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
		if contentLength <= 0 {
			return errors.New("cannot determine file size")
		}
	}

	// 3. 准备临时文件
	dir := filepath.Dir(dst)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmpFile, err := os.CreateTemp(dir, "download-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()
	defer func() {
		tmpFile.Close()
		_ = os.Remove(tmpPath) // 如果成功 Rename，此处删除会失效，符合预期
	}()

	// 4. 分片下载逻辑
	const chunkSize = 1024 * 1024 // 每次下载 1MB
	var downloaded int64 = 0

	for downloaded < contentLength {
		// 检查上下文是否取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		end := downloaded + chunkSize - 1
		if end >= contentLength {
			end = contentLength - 1
		}
		expectedSize := end - downloaded + 1

		var lastErr error
		const maxRetries = 5
		success := false

		for i := 0; i < maxRetries; i++ {
			if i > 0 {
				autoLog.Sugar.Warnf("下载分片失败，正在重试 (%d/%d): %v", i, maxRetries, lastErr)
				time.Sleep(time.Second * time.Duration(i))
			}

			// 重置文件指针到当前分片起始位置
			if _, err := tmpFile.Seek(downloaded, 0); err != nil {
				return fmt.Errorf("seek failed: %w", err)
			}

			// 创建带 Range 头的请求
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
			if err != nil {
				return err
			}
			rangeHeader := fmt.Sprintf("bytes=%d-%d", downloaded, end)
			req.Header.Set("Range", rangeHeader)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				lastErr = err
				continue
			}

			// 检查状态码
			if resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK {
				resp.Body.Close()
				lastErr = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
				continue
			}

			// 写入文件
			n, err := io.Copy(tmpFile, resp.Body)
			resp.Body.Close()

			if err != nil {
				lastErr = err
				continue
			}

			if n != expectedSize {
				lastErr = fmt.Errorf("short write: expected %d bytes, got %d", expectedSize, n)
				continue
			}

			downloaded += n
			success = true
			break
		}

		if !success {
			return fmt.Errorf("download chunk failed after %d retries: %w", maxRetries, lastErr)
		}

		// 5. 触发进度回调
		if onProgress != nil {
			onProgress(downloaded, contentLength)
		}
	}

	// 6. 原子重命名
	if err := tmpFile.Sync(); err != nil {
		return err
	}
	tmpFile.Close()

	if err := os.Rename(tmpPath, dst); err != nil {
		return err
	}

	return nil
}

// 更新bgi
func UpdateBgi() error {

	now := time.Now().Format("2006-01-02-15-04-05")

	//1、备份user文件
	err4 := ZipDir(config.Cfg.BetterGIAddress+"\\User\\", "Users\\User"+now+".zip", true)
	if err4 != nil {
		autoLog.Sugar.Errorf("备份失败: %v", err4)
		return fmt.Errorf("备份失败: %v", err4)
	}
	autoLog.Sugar.Infof("备份user文件")

	//备份log
	err2 := abgiCopy.Copy(config.Cfg.BetterGIAddress+"\\log\\", "backups\\log\\")
	if err2 != nil {
		autoLog.Sugar.Errorf("备份log失败: %v", err2)
		return fmt.Errorf("备份log失败: %v", err2)
	}
	autoLog.Sugar.Infof("备份log")

	//2、删除bgi文件夹
	err := os.RemoveAll(config.Cfg.BetterGIAddress)
	if err != nil {
		autoLog.Sugar.Errorf("删除bgi文件夹失败: %v", err)
		return fmt.Errorf("删除bgi文件夹失败: %v", err)
	}
	autoLog.Sugar.Infof("删除bgi文件夹")

	//3、解压bgi压缩包
	base := filepath.Base(config.Cfg.BetterGIAddress)
	path := strings.ReplaceAll(config.Cfg.BetterGIAddress, base, "")
	err4 = tools.Un7z("./uploads/BetterGI.zip", path)
	if err4 != nil {
		autoLog.Sugar.Errorf("解压bgi压缩包失败: %v", err)
		return fmt.Errorf("解压bgi压缩包失败: %v", err)
	}

	autoLog.Sugar.Infof("解压bgi压缩包")

	//4、删除user文件夹
	err2 = os.RemoveAll(config.Cfg.BetterGIAddress + "\\User\\")
	if err2 != nil {
		autoLog.Sugar.Errorf("删除user文件夹失败: %v", err2)
		return fmt.Errorf("删除user文件夹失败: %v", err2)
	}
	autoLog.Sugar.Infof("删除user文件夹")

	//5、复制压缩包到User
	err3 := abgiCopy.Copy("./Users/User"+now+".zip", config.Cfg.BetterGIAddress+"\\User.zip")
	if err3 != nil {
		autoLog.Sugar.Errorf("复制压缩包到User失败: %v", err3)
		return fmt.Errorf("复制压缩包到User失败: %v", err3)
	}
	autoLog.Sugar.Infof("复制压缩包到User")

	// 6、解压user压缩包
	err4 = tools.Unzip(config.Cfg.BetterGIAddress+"\\User.zip", config.Cfg.BetterGIAddress)
	if err4 != nil {
		autoLog.Sugar.Errorf("解压user压缩包失败: %v", err4)
		return fmt.Errorf("解压user压缩包失败: %v", err4)
	}
	autoLog.Sugar.Infof("解压user压缩包")

	//7、删除user压缩包
	err5 := os.Remove(config.Cfg.BetterGIAddress + "\\User.zip")
	if err5 != nil {
		autoLog.Sugar.Errorf("删除user压缩包失败: %v", err5)
	}

	//8、还原log
	err5 = abgiCopy.Copy("backups\\log\\", config.Cfg.BetterGIAddress+"\\log\\")
	if err5 != nil {
		autoLog.Sugar.Errorf("还原log失败: %v", err5)
	}
	autoLog.Sugar.Infof("还原log")

	err5 = os.Remove("./uploads/BetterGI.zip")
	if err5 != nil {
		autoLog.Sugar.Errorf("删除BetterGI压缩包失败: %v", err5)
	}
	autoLog.Sugar.Infof("删除BetterGI压缩包")
	return nil

}

// 获取更新url
func GetTeaBaoUpdateUrl() string {

	DownloadUrl := ""
	if strings.Contains(config.BgiCfg.RunForVersion, "lcb") {
		version, err := GetVersion()
		if err != nil {
			autoLog.Sugar.Errorf("获取远程版本失败: %v", err)
			return ""
		}
		DownloadUrl = fmt.Sprintf("%s_v%s.7z", abgiConstant.BgiUpdateUrl, version)
	} else {
		ydVersionInfo := GetYDVersionInfo()
		DownloadUrl = ydVersionInfo.BrowserDownloadUrl
	}

	return DownloadUrl
}
