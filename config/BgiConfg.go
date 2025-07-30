package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type OneLongConfigStruct struct {
	TaskEnabledList          map[string]bool `json:"TaskEnabledList"`
	Name                     string          `json:"Name"`
	CraftingBenchCountry     string          `json:"CraftingBenchCountry"`
	AdventurersGuildCountry  string          `json:"AdventurersGuildCountry"`
	PartyName                string          `json:"PartyName"`
	DomainName               string          `json:"DomainName"`
	WeeklyDomainEnabled      bool            `json:"WeeklyDomainEnabled"`
	DailyRewardPartyName     string          `json:"DailyRewardPartyName"`
	MinResinToKeep           int             `json:"MinResinToKeep"`
	SundayEverySelectedValue string          `json:"SundayEverySelectedValue"`
	SundaySelectedValue      string          `json:"SundaySelectedValue"`
	SecretTreasureObjects    []string        `json:"SecretTreasureObjects"`
	MondayPartyName          string          `json:"MondayPartyName"`
	MondayDomainName         string          `json:"MondayDomainName"`
	TuesdayPartyName         string          `json:"TuesdayPartyName"`
	TuesdayDomainName        string          `json:"TuesdayDomainName"`
	WednesdayPartyName       string          `json:"WednesdayPartyName"`
	WednesdayDomainName      string          `json:"WednesdayDomainName"`
	ThursdayPartyName        string          `json:"ThursdayPartyName"`
	ThursdayDomainName       string          `json:"ThursdayDomainName"`
	FridayPartyName          string          `json:"FridayPartyName"`
	FridayDomainName         string          `json:"FridayDomainName"`
	SaturdayPartyName        string          `json:"SaturdayPartyName"`
	SaturdayDomainName       string          `json:"SaturdayDomainName"`
	SundayPartyName          string          `json:"SundayPartyName"`
	SundayDomainName         string          `json:"SundayDomainName"`

	CompletionAction string `json:"CompletionAction"`
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
	fmt.Println(oneLongConfigStruct)

	return oneLongConfigStruct

}
