package menu

import (
	"auto-bgi/abgiConstant"
	"auto-bgi/autoLog"
	_ "embed"
	"fmt"
	"github.com/inconshreveable/go-update"
	"io"
	"net/http"
	"strings"
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

	// 记录开始下载新版本的日志
	autoLog.Sugar.Infof("开始下载新版本")
	// 通过HTTP GET请求获取更新文件
	resp, err := http.Get(abgiConstant.ABgiUpdateUrl)
	if err != nil {
		// 如果请求失败，返回错误
		return err
	}
	// 确保响应体被关闭，以释放资源
	defer resp.Body.Close()

	// 应用下载的更新
	err = update.Apply(resp.Body, update.Options{})
	if err != nil {
		// 如果应用更新失败，返回错误zoa
		return err
	}

	// 记录更新成功的日志，并提示用户重启程序
	autoLog.Sugar.Infof("更新成功！请重新启动abgi")

	//// 调用 run_auto_bgi.vbs 脚本来启动新的 auto-bgi.exe 程序
	//if err := tools.RestartProgram(); err != nil {
	//	return fmt.Errorf("重启程序失败: %v", err)
	//}

	// 更新成功，返回nil表示无错误
	return nil
}
