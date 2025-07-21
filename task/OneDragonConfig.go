package task

import (
	"encoding/json"
	"fmt"
	"os"
)

type Task struct {
	Item1 bool   `json:"Item1"`
	Item2 string `json:"Item2"`
}

type ChaBaoBgiConfig struct {
	SelectedPeriodList       []string        `json:"SelectedPeriodList"`
	TaskEnabledList          map[string]Task `json:"TaskEnabledList"`
	GenshinUid               string          `json:"GenshinUid"`
	AccountBinding           bool            `json:"AccountBinding"`
	Version                  int             `json:"Version"`
	Name                     string          `json:"Name"`
	IndexId                  int             `json:"IndexId"`
	NextConfiguration        bool            `json:"NextConfiguration"`
	NextTaskIndex            int             `json:"NextTaskIndex"`
	Period                   string          `json:"Period"`
	ScheduleName             string          `json:"ScheduleName"`
	CustomDomainList         []string        `json:"CustomDomainList"`
	CraftingBenchCountry     string          `json:"CraftingBenchCountry"`
	AdventurersGuildCountry  string          `json:"AdventurersGuildCountry"`
	PartyName                string          `json:"PartyName"`
	DomainName               string          `json:"DomainName"`
	WeeklyDomainEnabled      bool            `json:"WeeklyDomainEnabled"`
	DailyRewardPartyName     string          `json:"DailyRewardPartyName"`
	MinResinToKeep           int             `json:"MinResinToKeep"`
	SundayEverySelectedValue string          `json:"SundayEverySelectedValue"`
	SundaySelectedValue      string          `json:"SundaySelectedValue"`
	SereniteaPotTpType       string          `json:"SereniteaPotTpType"`
	SecretTreasureObjects    []string        `json:"SecretTreasureObjects"`
	ResinCount               map[string]int  `json:"ResinCount"`
	SpecifyResinUse          bool            `json:"SpecifyResinUse"`
	AccountBindingCode       string          `json:"AccountBindingCode"`
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
	CompletionAction         string          `json:"CompletionAction"`
}

func ReadChaBaoBgiConfig(filename string) {
	file, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("读取文件失败:", err)
		return
	}

	var cfg ChaBaoBgiConfig
	if err := json.Unmarshal(file, &cfg); err != nil {
		fmt.Println("解析 JSON 失败:", err)
		return
	}

	// 示例输出部分内容
	fmt.Println("计划名称:", cfg.ScheduleName)
	fmt.Println("原神 UID:", cfg.GenshinUid)
	fmt.Println("启用的任务:")
	for id, task := range cfg.TaskEnabledList {
		fmt.Printf("  #%s: %v - %s\n", id, task.Item1, task.Item2)
	}
}
