package config

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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
	GenShinStartConfigInstallPath string       `json:"genshinStartConfig.installPath"` // 原神安装目录
	RunForVersion                 string       `json:"runForVersion"`                  // bgi版本号
	SelectedChannelName           string       `json:"selectedChannelName"`            // selectedChannelName：仓库
	RepoUrl                       string       `json:"repoUrl"`                        // 仓库地址
	MiYouSheConfigCookie          string       `json:"miyousheConfig"`                 // miyousheConfig
	HotKeyConfig                  HotKeyConfig `json:"shortcutKey"`                    //bgi快捷键
}

type HotKeyConfig struct {
	BgiEnabledHotkey     string `json:"BgiEnabledHotkey"`      //启动停止BetterGI
	CancelTaskHotkey     string `json:"CancelTaskHotkey" `     //停止当前脚本/独立任务
	SuspendHotkey        string `json:"SuspendHotkey" `        //暂停当前脚本/独立任务
	TakeScreenshotHotkey string `json:"TakeScreenshotHotkey" ` // 游戏截图
	LogBoxDisplayHotkey  string `json:"LogBoxDisplayHotkey"`   //日志与状态窗口展示开关
	OnedragonHotkey      string `json:"OnedragonHotkey"`       //启动停止一条龙

}

var BgiCfg BgiConfig

func ReadBgiConfig() {

	configPath := Cfg.BetterGIAddress + "\\User\\Config.json"
	// 读取配置文件内容，忽略可能出现的错误
	configData, _ := os.ReadFile(configPath)

	data := string(configData)

	/**
	*获取快捷键
	 */
	BgiCfg.HotKeyConfig.BgiEnabledHotkey = gjson.Get(data, "hotKeyConfig.bgiEnabledHotkey").String()
	BgiCfg.HotKeyConfig.CancelTaskHotkey = gjson.Get(data, "hotKeyConfig.cancelTaskHotkey").String()
	BgiCfg.HotKeyConfig.SuspendHotkey = gjson.Get(data, "hotKeyConfig.suspendHotkey").String()
	BgiCfg.HotKeyConfig.TakeScreenshotHotkey = gjson.Get(data, "hotKeyConfig.takeScreenshotHotkey").String()
	BgiCfg.HotKeyConfig.LogBoxDisplayHotkey = gjson.Get(data, "hotKeyConfig.logBoxDisplayHotkey").String()
	BgiCfg.HotKeyConfig.OnedragonHotkey = gjson.Get(data, "hotKeyConfig.onedragonHotkey").String()

	//end

	//原神安装目录
	BgiCfg.GenShinStartConfigInstallPath = gjson.Get(data, "genshinStartConfig.installPath").String()

	//bgi版本号
	BgiCfg.RunForVersion = gjson.Get(data, "commonConfig.runForVersion").String()

	BgiCfg.MiYouSheConfigCookie = strings.ReplaceAll(gjson.Get(data, "otherConfig.miyousheConfig.cookie").String(), " ", "")
	BgiCfg.MiYouSheConfigCookie = strings.ReplaceAll(BgiCfg.MiYouSheConfigCookie, "\n", "")
	fmt.Println(BgiCfg.MiYouSheConfigCookie)

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

	fmt.Println(BgiCfg.HotKeyConfig)

}

func HotKeyQuery(context *gin.Context) {

	configPath := Cfg.BetterGIAddress + "\\User\\Config.json"
	// 读取配置文件内容，忽略可能出现的错误
	configData, _ := os.ReadFile(configPath)

	data := string(configData)

	BgiCfg.HotKeyConfig.BgiEnabledHotkey = gjson.Get(data, "hotKeyConfig.bgiEnabledHotkey").String()
	BgiCfg.HotKeyConfig.CancelTaskHotkey = gjson.Get(data, "hotKeyConfig.cancelTaskHotkey").String()
	BgiCfg.HotKeyConfig.SuspendHotkey = gjson.Get(data, "hotKeyConfig.suspendHotkey").String()
	BgiCfg.HotKeyConfig.TakeScreenshotHotkey = gjson.Get(data, "hotKeyConfig.takeScreenshotHotkey").String()
	BgiCfg.HotKeyConfig.LogBoxDisplayHotkey = gjson.Get(data, "hotKeyConfig.logBoxDisplayHotkey").String()
	BgiCfg.HotKeyConfig.OnedragonHotkey = gjson.Get(data, "hotKeyConfig.onedragonHotkey").String()

	mapData := make(map[string]string)
	mapData["启动停止BetterGI"] = BgiCfg.HotKeyConfig.BgiEnabledHotkey
	mapData["停止当前脚本/ 独立任务"] = BgiCfg.HotKeyConfig.CancelTaskHotkey
	mapData["暂停当前脚本/ 独立任务"] = BgiCfg.HotKeyConfig.SuspendHotkey
	mapData["游戏截图"] = BgiCfg.HotKeyConfig.TakeScreenshotHotkey
	mapData["日志与状态窗口展示开关"] = BgiCfg.HotKeyConfig.LogBoxDisplayHotkey
	mapData["启动停止一条龙"] = BgiCfg.HotKeyConfig.OnedragonHotkey

	context.JSON(http.StatusOK, gin.H{"status": "success", "data": mapData})

}
