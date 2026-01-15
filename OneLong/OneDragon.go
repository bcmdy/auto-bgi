package OneLong

import (
	"auto-bgi/config"
	"encoding/json"
	"fmt"
	"github.com/iancoleman/orderedmap"
	"os"
	"path/filepath"
)

type OneDragon struct {
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
	WednesdayPartyName       string                 `json:"WednesdayPartyName"`
	WednesdayDomainName      string                 `json:"WednesdayDomainName"`
	ThursdayPartyName        string                 `json:"ThursdayPartyName"`
	ThursdayDomainName       string                 `json:"ThursdayDomainName"`
	FridayPartyName          string                 `json:"FridayPartyName"`
	FridayDomainName         string                 `json:"FridayDomainName"`
	SaturdayPartyName        string                 `json:"SaturdayPartyName"`
	SaturdayDomainName       string                 `json:"SaturdayDomainName"`
	SundayPartyName          string                 `json:"SundayPartyName"`
	SundayDomainName         string                 `json:"SundayDomainName"`
	CompletionAction         string                 `json:"CompletionAction"`
}

func (oneDragon *OneDragon) ReadOneDragon(longName string) error {
	filename := filepath.Join(config.Cfg.BetterGIAddress, "User", "OneDragon", longName+".json")

	file, err := os.ReadFile(filename)
	if err != nil {

		return fmt.Errorf("ReadOneDragon读取文件失败:%s ", err)
	}

	if err := json.Unmarshal(file, &oneDragon); err != nil {

		return fmt.Errorf("解析 JSON 失败: %w", err)
	}

	return nil
}

func (oneDragon *OneDragon) SaveOneDragon(longName string) error {

	filename := filepath.Join(config.Cfg.BetterGIAddress, "User", "OneDragon", longName+".json")

	// 格式化输出，方便人工看是的
	data, err := json.MarshalIndent(oneDragon, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化 JSON 失败: %w", err)
	}

	// 写入文件（0644 表示普通读写权限）
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
}

func (oneDragon *OneDragon) ChangeTaskEnabledList(longName string) error {

	err := oneDragon.ReadOneDragon(longName)
	if err != nil {
		return err
	}

	err = oneDragon.SaveOneDragon(longName)
	if err != nil {
		return err
	}
	return nil
}
