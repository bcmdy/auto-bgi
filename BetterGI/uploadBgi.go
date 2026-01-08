package BetterGI

import (
	"auto-bgi/abgiConstant"
	"auto-bgi/autoLog"
	"auto-bgi/bgiStatus"
	"auto-bgi/config"
	"auto-bgi/control"
	"auto-bgi/tools"
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	abgiCopy "github.com/otiai10/copy"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
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
		bgiStatus.GitPull()
	}()

	c.JSON(http.StatusOK, gin.H{
		"message":  "更新成功",
		"filename": file.Filename,
		"size":     file.Size,
	})
}

var (
	downloadLock   sync.Mutex
	lastInvokeTime time.Time
	// 2分钟的时间间隔
	invokeInterval = 5 * time.Minute
)

// DownloadBgi 是 Gin handler：从请求中读取 url（form/json/query），下载压缩包并替换到 ./uploads/BetterGI.zip
// DownloadBgi 处理下载BGI的请求函数
func DownloadBgi(c *gin.Context) {

	//if !strings.Contains(config.BgiCfg.RunForVersion, "lcb") {
	//	autoLog.Sugar.Error("当前版本不是lcb版本，无法更新")
	//	c.JSON(http.StatusBadRequest, gin.H{
	//		"message": "当前版本不是lcb版本，无法更新",
	//	})
	//	return
	//}
	//
	//// 加锁检查调用频率
	//downloadLock.Lock()
	//// 检查是否在2分钟内重复调用
	//if time.Since(lastInvokeTime) < invokeInterval && !lastInvokeTime.IsZero() {
	//	downloadLock.Unlock()
	//	autoLog.Sugar.Warn("操作过于频繁，请等待2分钟后再试")
	//	c.JSON(http.StatusTooManyRequests, gin.H{
	//		"message": "操作过于频繁，请等待2分钟后再试",
	//	})
	//	return
	//}
	//// 更新最后调用时间
	//lastInvokeTime = time.Now()
	//downloadLock.Unlock()
	//
	////判断是否需要更新
	//// 获取最新版本信息
	//version, err := GetVersion()
	//if err != nil {
	//	// 记录错误日志并返回错误响应
	//	autoLog.Sugar.Error("获取版本失败:", err)
	//	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	//	return
	//}
	//// 记录最新版本信息
	//autoLog.Sugar.Infof("当前BGI版本: %s", config.BgiCfg.RunForVersion)
	//autoLog.Sugar.Infof("最新BGI版本: %s", version)
	//
	//// 检查当前版本是否为最新版本
	//if version == config.BgiCfg.RunForVersion {
	//	c.JSON(http.StatusOK, gin.H{
	//		"message": "当前版本已经是最新版本",
	//	})
	//	return
	//}
	//
	//// 关闭软件（和上传逻辑一致）
	//control.CloseSoftware()

	// 下载并保存
	dst := filepath.Join("./uploads", "BetterGI.zip")
	// ctx with timeout for the whole download operation
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	err2 := downloadFileWithProgress(ctx, GetTeaBaoUpdateUrl(), dst, func(current, total int64) {
		percent := float64(current) / float64(total) * 100
		fmt.Printf("\r下载进度: %.2f%% (%d/%d bytes)", percent, current, total)
	})
	if err2 != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err2.Error()})
		return
	}

	//if err := downloadFileFromURL(ctx, GetTeaBaoUpdateUrl(), dst); err != nil {
	//	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	//	return
	//}
	//// 小延时确保文件系统稳定（保持与 UploadBgi 一致）
	//time.Sleep(1 * time.Second)
	//
	//// 更新 bgi（和 upload 的行为一致）
	//err = UpdateBgi()
	//if err != nil {
	//	autoLog.Sugar.Error("更新失败:", err)
	//	c.JSON(http.StatusBadRequest, gin.H{
	//		"message": err,
	//	})
	//	return
	//}
	//
	//// 更新仓库
	//bgiStatus.GitPull()
	//
	//config.BgiCfg.RunForVersion = version
	//
	//control.OpenSoftware(config.Cfg.BetterGIAddress + "\\BetterGI.exe")

	c.JSON(http.StatusOK, gin.H{
		"message": "下载并更新成功,请自启bgi,更新版本",
		"path":    dst,
	})
}

//func DownloadBgiProgress(c *gin.Context) {
//
//	// 下载并保存
//	dst := filepath.Join("./uploads", "BetterGI.zip")
//	// ctx with timeout for the whole download operation
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
//	defer cancel()
//
//	err2 := downloadFileWithProgress(ctx, GetTeaBaoUpdateUrl(), dst, func(current, total int64) {
//		percent := float64(current) / float64(total) * 100
//		fmt.Printf("\r下载进度: %.2f%% (%d/%d bytes)", percent, current, total)
//	})
//	if err2 != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": err2.Error()})
//		return
//	}
//}

func DownloadBgiProgress(c *gin.Context) {

	if !strings.Contains(config.BgiCfg.RunForVersion, "lcb") {
		autoLog.Sugar.Error("当前版本不是lcb版本，无法更新")
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "当前版本不是lcb版本，无法更新",
		})
		return
	}

	// 加锁检查调用频率
	downloadLock.Lock()
	// 检查是否在2分钟内重复调用
	if time.Since(lastInvokeTime) < invokeInterval && !lastInvokeTime.IsZero() {
		downloadLock.Unlock()
		autoLog.Sugar.Warn("操作过于频繁，请等待2分钟后再试")
		c.JSON(http.StatusTooManyRequests, gin.H{
			"message": "操作过于频繁，请等待2分钟后再试",
		})
		return
	}
	// 更新最后调用时间
	lastInvokeTime = time.Now()
	downloadLock.Unlock()

	//判断是否需要更新
	// 获取最新版本信息
	version, err := GetVersion()
	if err != nil {
		// 记录错误日志并返回错误响应
		autoLog.Sugar.Error("获取版本失败:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// 记录最新版本信息
	autoLog.Sugar.Infof("当前BGI版本: %s", config.BgiCfg.RunForVersion)
	autoLog.Sugar.Infof("最新BGI版本: %s", version)

	// 检查当前版本是否为最新版本
	if version == config.BgiCfg.RunForVersion {
		c.JSON(http.StatusOK, gin.H{
			"message": "当前版本已经是最新版本",
		})
		return
	}

	// 关闭软件（和上传逻辑一致）
	control.CloseSoftware()

	// 设置响应头，声明这是一个 SSE 流
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	// 设置内容类型为服务器发送事件
	dst := filepath.Join("./uploads", "BetterGI.zip")                       // 禁用缓存
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Minute) // 保持长连接
	defer cancel()
	// 构建目标文件路径

	// 创建一个10分钟超时的上下文
	err = downloadFileWithProgress(ctx, GetTeaBaoUpdateUrl(), dst, func(current, total int64) {
		percent := float64(current) / float64(total) * 100 // 确保在函数返回前取消上下文
		// 发送 SSE 格式的数据：data: {"percent": 50.5}\n\n
		// 调用下载函数，并传入进度回调函数
		c.SSEvent("progress", gin.H{
			"percent": fmt.Sprintf("%.2f", percent),
			"current": current,
			"total":   total, // 使用SSE发送进度事件
		}) // 保留两位小数的百分比
		c.Writer.Flush() // 立即推送到前端                      // 当前已下载字节数
	}) // 文件总字节数

	if err != nil {
		c.SSEvent("error", err.Error())
		return
		// 如果下载过程中发生错误，发送错误事件
	}

	c.SSEvent("done", "success")

	// 小延时确保文件系统稳定（保持与 UploadBgi 一致）
	time.Sleep(1 * time.Second)

	// 更新 bgi（和 upload 的行为一致）
	err = UpdateBgi()
	if err != nil {
		autoLog.Sugar.Error("更新失败:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err,
		})
		return
	}

	//// 更新仓库
	//bgiStatus.GitPull()
	//更新仓库
	go func() {
		bgiStatus.GitPull()
	}()

	config.BgiCfg.RunForVersion = version

	control.OpenSoftware(config.Cfg.BetterGIAddress + "\\BetterGI.exe")

	autoLog.Sugar.Infof("更新成功,请自启bgi,更新版本")

	//重定向
	c.Redirect(http.StatusFound, "/")

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

	contentLength, _ := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
	if contentLength <= 0 {
		return errors.New("cannot determine file size")
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

		// 创建带 Range 头的请求
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
		if err != nil {
			return err
		}
		rangeHeader := fmt.Sprintf("bytes=%d-%d", downloaded, end)
		req.Header.Set("Range", rangeHeader)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}

		// 写入文件
		n, err := io.Copy(tmpFile, resp.Body)
		resp.Body.Close()
		if err != nil {
			return fmt.Errorf("write chunk failed: %w", err)
		}

		downloaded += n

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

// downloadFileFromURL 从给定 URL 下载文件并原子写入到 dst（覆盖）。
// ctx: 下载超时/取消上下文。
// reqURL: 要下载的文件 URL。
// dst: 最终目标文件路径（会先写入临时文件再重命名）。
func downloadFileFromURL(ctx context.Context, reqURL string, dst string) error {
	// 验证 URL
	parsed, err := url.ParseRequestURI(reqURL)
	if err != nil {
		return fmt.Errorf("invalid url: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return errors.New("url scheme must be http or https")
	}

	// 确保目标目录存在
	dir := filepath.Dir(dst)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create dir %s: %w", dir, err)
	}

	// HTTP 请求（使用自定义 client 以便设置超时）
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	client := &http.Client{
		// 不在这里设置 Timeout，因为我们使用 ctx 带超时；但为了保险可以设置一个大超时
		Timeout: 0,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("download request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("download failed: server returned status %s", resp.Status)
	}

	// 写入临时文件
	tmpFile, err := os.CreateTemp(dir, "bettergi-*.zip.tmp")
	if err != nil {
		return fmt.Errorf("create temp file failed: %w", err)
	}
	tmpPath := tmpFile.Name()
	// 确保临时文件在出错时被移除
	cleanup := func() {
		tmpFile.Close()
		os.Remove(tmpPath)
	}
	defer func() {
		// 如果已经重命名成功则 tmpPath 不存在，不过 Remove 不会报错
		_ = tmpFile.Close()
	}()

	// 流式拷贝（不把文件全部读入内存）
	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		cleanup()
		return fmt.Errorf("writing to temp file failed: %w", err)
	}

	// 确保写入磁盘
	if err := tmpFile.Sync(); err != nil {
		cleanup()
		return fmt.Errorf("sync temp file failed: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		cleanup()
		return fmt.Errorf("close temp file failed: %w", err)
	}

	// 原子替换目标文件（在同一文件系统上，os.Rename 是原子的）
	if err := os.Rename(tmpPath, dst); err != nil {
		// 若 Rename 失败，尝试删除临时文件
		_ = os.Remove(tmpPath)
		return fmt.Errorf("rename temp to dst failed: %w", err)
	}

	return nil
}

// 更新bgi
func UpdateBgi() error {

	now := time.Now().Format("2006-01-02-15-04-05")

	//1、备份user文件
	err4 := bgiStatus.ZipDir(config.Cfg.BetterGIAddress+"\\User\\", "Users\\User"+now+".zip", true)
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

// 获取茶包版更新url
func GetTeaBaoUpdateUrl() string {
	version, err := GetVersion()
	if err != nil {
		autoLog.Sugar.Errorf("获取远程版本失败: %v", err)
		return ""
	}
	BgiUpdateUrl := fmt.Sprintf("%s_v%s.7z", abgiConstant.BgiUpdateUrl, version)

	return BgiUpdateUrl
}
