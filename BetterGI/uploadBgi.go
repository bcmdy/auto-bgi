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

// DownloadBgi 是 Gin handler：从请求中读取 url（form/json/query），下载压缩包并替换到 ./uploads/BetterGI.zip
// DownloadBgi 处理下载BGI的请求函数
func DownloadBgi(c *gin.Context) {

	if !strings.Contains(config.BgiCfg.RunForVersion, "lcb") {
		autoLog.Sugar.Error("当前版本不是lcb版本，无法更新")
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "当前版本不是lcb版本，无法更新",
		})
		return
	}

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

	// 下载并保存
	dst := filepath.Join("./uploads", "BetterGI.zip")
	// ctx with timeout for the whole download operation
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := downloadFileFromURL(ctx, abgiConstant.BgiUpdateUrl, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

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
	// 更新仓库（异步）
	go func() {
		bgiStatus.GitPull()
	}()

	config.BgiCfg.RunForVersion = version

	c.JSON(http.StatusOK, gin.H{
		"message": "下载并更新成功,请自启bgi,更新版本",
		"path":    dst,
	})
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

	//删除log
	//err5 = os.RemoveAll("backups\\log\\")
	//if err5 != nil {
	//	autoLog.Sugar.Errorf("删除log失败: %v", err5)
	//}
	//autoLog.Sugar.Infof("删除log")

}
