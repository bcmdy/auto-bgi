package menu

import (
	"auto-bgi/abgiConstant"
	"auto-bgi/autoLog"
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/inconshreveable/go-update"
)

//go:embed web/version.txt
var CurrentVersion string

func GetVersion() (string, error) {

	// 创建请求
	req, err := http.NewRequest("GET", abgiConstant.GetLastVersion, nil)
	if err != nil {
		panic(err)
	}

	// 添加请求头，比如 User-Agent
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json")

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

	autoLog.Sugar.Infof("当前最新版本: %s", release)
	return release, nil
}

// 更新版本
// Update 函数用于执行程序的更新操作
// 它会从指定的URL下载新版本并应用更新
// 如果更新过程中出现任何错误，将返回错误信息
func Update() error {
	//版本判断
	version, _ := GetVersion()
	CurrentVersion = strings.TrimSpace(CurrentVersion)
	if CurrentVersion == version {
		autoLog.Sugar.Infof("当前版本为最新版本")

		return fmt.Errorf("当前版本为最新版本")
	}

	DownloadUrl := abgiConstant.ABgiUpdateUrl

	// 记录开始下载新版本的日志
	autoLog.Sugar.Infof("开始下载新版本: %s", DownloadUrl)
	// 通过HTTP GET请求获取更新文件
	resp, err := http.Get(DownloadUrl)
	if err != nil {
		// 如果请求失败，返回错误
		return fmt.Errorf("下载请求失败: %v", err)
	}
	// 确保响应体被关闭，以释放资源
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败，HTTP状态码: %d", resp.StatusCode)
	}

	// 检查 Content-Type (如果是网页，则报错)
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "text/html") {
		return fmt.Errorf("下载链接返回了网页内容而非二进制文件 (Content-Type: %s)，可能是链接失效或需要验证", contentType)
	}

	// 预读取前 2 个字节检查 MZ 头
	head := make([]byte, 2)
	_, err = io.ReadFull(resp.Body, head)
	if err != nil {
		return fmt.Errorf("读取文件头失败: %v", err)
	}

	if string(head) != "MZ" {
		return fmt.Errorf("下载的文件可能已损坏或不是有效的 PE 文件 (缺少 MZ 头)")
	}

	// 组合回 reader
	reader := io.MultiReader(bytes.NewReader(head), resp.Body)

	// 应用下载的更新
	err = update.Apply(reader, update.Options{})
	if err != nil {
		// 如果应用更新失败，返回错误zoa
		return err
	}

	// 记录更新成功的日志，并提示用户重启程序
	autoLog.Sugar.Infof("更新成功！请重新启动abgi")

	// 更新成功，返回nil表示无错误
	return nil
}
