package config

import (
	"encoding/json"
	"fmt"
	"github.com/iancoleman/orderedmap"
	"os"
	"path/filepath"
	"strings"
)

type OneLongConfigStruct struct {
	TaskEnabledList          *orderedmap.OrderedMap `json:"TaskEnabledList"`
	Name                     string                 `json:"Name"`
	CraftingBenchCountry     string                 `json:"CraftingBenchCountry"`
	AdventurersGuildCountry  string                 `json:"AdventurersGuildCountry"`
	PartyName                string                 `json:"PartyName"`
	DomainName               string                 `json:"DomainName"`
	WeeklyDomainEnabled      bool                   `json:"WeeklyDomainEnabled"`
	DailyRewardPartyName     string                 `json:"DailyRewardPartyName"`
	MinResinToKeep           int                    `json:"MinResinToKeep"`
	SundayEverySelectedValue string                 `json:"SundayEverySelectedValue"`
	SundaySelectedValue      string                 `json:"SundaySelectedValue"`
	SereniteaPotTpType       string                 `json:"SereniteaPotTpType"`
	SecretTreasureObjects    []string               `json:"SecretTreasureObjects"`
	MondayPartyName          string                 `json:"MondayPartyName"`
	MondayDomainName         string                 `json:"MondayDomainName"`
	TuesdayPartyName         string                 `json:"TuesdayPartyName"`
	TuesdayDomainName        string                 `json:"TuesdayDomainName"`

	WednesdayPartyName  string `json:"WednesdayPartyName"`
	WednesdayDomainName string `json:"WednesdayDomainName"`
	ThursdayPartyName   string `json:"ThursdayPartyName"`
	ThursdayDomainName  string `json:"ThursdayDomainName"`
	FridayPartyName     string `json:"FridayPartyName"`
	FridayDomainName    string `json:"FridayDomainName"`
	SaturdayPartyName   string `json:"SaturdayPartyName"`
	SaturdayDomainName  string `json:"SaturdayDomainName"`
	SundayPartyName     string `json:"SundayPartyName"`
	SundayDomainName    string `json:"SundayDomainName"`

	CompletionAction string `json:"CompletionAction"`
}

type TaskItem struct {
	Name    string `json:"Name"`
	Enabled bool   `json:"Enabled"`
}

// 读取一条龙配置
func OneLongConfig(name string) OneLongConfigStruct {
	filename := Cfg.BetterGIAddress + "\\User\\OneDragon\\" + name + ".json"
	file, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("读取文件失败:", err)
		return OneLongConfigStruct{}
	}

	var oneLongConfigStruct OneLongConfigStruct
	if err := json.Unmarshal(file, &oneLongConfigStruct); err != nil {
		fmt.Println("解析 JSON 失败:", err)
		return OneLongConfigStruct{}
	}

	return oneLongConfigStruct
}

// 读取所有一条龙配置
func OneLongAllName() []string {
	entries, err := os.ReadDir(Cfg.BetterGIAddress + "\\User\\OneDragon")
	if err != nil {
		return []string{}
	}
	var oneLongInfo []string
	for _, entry := range entries {

		//去除后缀：.json
		name := strings.ReplaceAll(entry.Name(), ".json", "")

		oneLongInfo = append(oneLongInfo, name)

	}
	return oneLongInfo
}

// 保存一条龙配置（保持 TaskEnabledList 顺序）
func SaveOneLongConfig(cfg OneLongConfigStruct) error {
	// 目标路径
	filename := filepath.Join(Cfg.BetterGIAddress, "User", "OneDragon", cfg.Name+".json")

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// ✅ 使用 json.MarshalIndent 保证 JSON 格式美观
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化 JSON 失败: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	fmt.Println("配置已保存到:", filename)
	return nil
}

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
