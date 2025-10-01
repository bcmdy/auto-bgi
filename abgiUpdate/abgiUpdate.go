package abgiUpdate

import (
	"auto-bgi/autoLog"
	"encoding/json"
	"fmt"
	"github.com/inconshreveable/go-update"
	"net/http"
)

const CurrentVersion = "auto-bgi-2.9"

type ABgi struct {
	Version string
}

// 获取当前版本
func (v *ABgi) GetCurrentVersion() string {
	return CurrentVersion
}

func (v *ABgi) GetVersion() (string, error) {

	// 创建请求
	req, err := http.NewRequest("GET", "https://gitee.com/api/v5/repos/wangjian0327/auto-bgi/releases/latest", nil)
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

	var release struct {
		TagName string `json:"tag_name"`
		Assets  []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}
	v.Version = release.TagName

	return release.TagName, nil
}

// 更新版本
func (v *ABgi) Update() error {

	url := "https://gitee.com/api/v5/repos/wangjian0327/auto-bgi/releases/latest"

	// 创建请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}

	// 添加请求头，比如 User-Agent
	req.Header.Set("User-Agent", "MyCustomClient/1.0")
	req.Header.Set("Accept", "application/json")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var release struct {
		TagName string `json:"tag_name"`
		Assets  []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return err
	}

	// 2. 比较版本
	if release.TagName == CurrentVersion {
		autoLog.Sugar.Infof("当前已是最新版本")
		return nil
	}

	// 3. 找到对应平台的文件
	var downloadURL string

	for _, asset := range release.Assets {
		if asset.Name == "auto-bgi.exe" {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("未找到适配的更新文件")
	}

	// 4. 下载并更新

	autoLog.Sugar.Infof("开始下载新版本:%s", release.TagName)
	resp, err = http.Get(downloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = update.Apply(resp.Body, update.Options{})
	if err != nil {
		return err
	}

	autoLog.Sugar.Infof("更新成功！请重新启动abgi")
	return nil
}
