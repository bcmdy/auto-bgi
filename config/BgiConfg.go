package config

import (
	"encoding/json"
	"github.com/tidwall/gjson"
	"os"
	"path/filepath"
)

type ManifestStruct struct {
	ManifestVersion int      `json:"manifest_version" comment:"版本"`
	Name            string   `json:"Name" comment:"名称"`
	Version         string   `json:"Version" comment:"版本"`
	Description     string   `json:"Description" comment:"描述"`
	Authors         []Author `json:"authors" comment:"作者"`
	SettingsUi      string   `json:"settings_ui" comment:"设置界面"`
	Main            string   `json:"main" comment:"主文件"`
	SavedFiles      []string `json:"saved_files" comment:"需要备份的文件"`
}

type Author struct {
	Name  string `json:"name"`
	Links string `json:"links"`
}

// 读取manifest.json
func ReadManifest(jsName string) (ManifestStruct, error) {
	manifestPath := filepath.Join(jsName, "manifest.json")
	file, err := os.ReadFile(manifestPath)
	if err != nil {
		return ManifestStruct{}, err
	}
	var manifest ManifestStruct
	if err := json.Unmarshal(file, &manifest); err != nil {
		return ManifestStruct{}, err
	}
	return manifest, nil
}

type BgiConfig struct {
	CancelTaskHotkey              string `json:"cancelTaskHotkey"`               // 取消任务的快捷键值
	GenShinStartConfigInstallPath string `json:"genshinStartConfig.installPath"` // 原神安装目录
	RunForVersion                 string `json:"runForVersion"`                  // bgi版本号
	SelectedChannelName           string `json:"selectedChannelName"`            // selectedChannelName：仓库
	RepoUrl                       string `json:"repoUrl"`                        // 仓库地址
	MiYouSheConfigCookie          string `json:"miyousheConfig"`                 // miyousheConfig
}

var BgiCfg BgiConfig

func ReadBgiConfig() {

	configPath := Cfg.BetterGIAddress + "\\User\\Config.json"
	// 读取配置文件内容，忽略可能出现的错误
	configData, _ := os.ReadFile(configPath)

	data := string(configData)
	// 从配置文件中获取取消任务的快捷键值
	BgiCfg.CancelTaskHotkey = gjson.Get(data, "hotKeyConfig.cancelTaskHotkey").String()

	//原神安装目录
	BgiCfg.GenShinStartConfigInstallPath = gjson.Get(data, "genshinStartConfig.installPath").String()

	//bgi版本号
	BgiCfg.RunForVersion = gjson.Get(data, "commonConfig.runForVersion").String()

	BgiCfg.MiYouSheConfigCookie = gjson.Get(data, "otherConfig.miyousheConfig.cookie").String()

	//selectedChannelName：仓库
	BgiCfg.SelectedChannelName = gjson.Get(data, "scriptConfig.selectedChannelName").String()

	if BgiCfg.SelectedChannelName == "GitHub" {
		BgiCfg.RepoUrl = "https://github.com/babalae/bettergi-scripts-list.git"
	} else if BgiCfg.SelectedChannelName == "GitCode" {
		BgiCfg.RepoUrl = "https://gitcode.com/huiyadanli/bettergi-scripts-list.git"
	} else if BgiCfg.SelectedChannelName == "" {
		BgiCfg.RepoUrl = "https://cnb.cool/bettergi/bettergi-scripts-list.git"
	} else {
		BgiCfg.RepoUrl = "https://cnb.cool/bettergi/bettergi-scripts-list.git"
	}

}
