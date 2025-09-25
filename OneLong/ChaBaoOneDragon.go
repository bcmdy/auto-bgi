package OneLong

import (
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

type ChaBaoBgiConfig struct {
	EverySelectedValueListConverter string              `json:"EverySelectedValueListConverter"`
	TaskEnabledList                 map[string]TaskItem `json:"TaskEnabledList"`
	GenshinUid                      string              `json:"GenshinUid"`
	AccountBinding                  bool                `json:"AccountBinding"`
	Version                         int                 `json:"Version"`
	Name                            string              `json:"Name"`
	IndexId                         int                 `json:"IndexId"`
	NextConfiguration               bool                `json:"NextConfiguration"`
	NextTaskIndex                   int                 `json:"NextTaskIndex"`
	Period                          string              `json:"Period"`
	PeriodList                      map[string]bool     `json:"PeriodList"`
	ScheduleName                    string              `json:"ScheduleName"`
	CustomDomainList                []string            `json:"CustomDomainList"`
	CraftingBenchCountry            string              `json:"CraftingBenchCountry"`
	AdventurersGuildCountry         string              `json:"AdventurersGuildCountry"`
	PartyName                       string              `json:"PartyName"`
	DomainName                      string              `json:"DomainName"`
	WeeklyDomainEnabled             bool                `json:"WeeklyDomainEnabled"`
	DailyRewardPartyName            string              `json:"DailyRewardPartyName"`
	MinResinToKeep                  int                 `json:"MinResinToKeep"`
	SundayEverySelectedValue        string              `json:"SundayEverySelectedValue"`
	SundaySelectedValue             string              `json:"SundaySelectedValue"`
	SereniteaPotTpType              string              `json:"SereniteaPotTpType"`
	SecretTreasureObjects           []string            `json:"SecretTreasureObjects"`
	ResinCount                      ResinCount          `json:"ResinCount"`
	SpecifyResinUse                 bool                `json:"SpecifyResinUse"`
	AccountBindingCode              string              `json:"AccountBindingCode"`
	MondayPartyName                 string              `json:"MondayPartyName"`
	MondayDomainName                string              `json:"MondayDomainName"`
	TuesdayPartyName                string              `json:"TuesdayPartyName"`
	TuesdayDomainName               string              `json:"TuesdayDomainName"`
	WednesdayPartyName              string              `json:"WednesdayPartyName"`
	WednesdayDomainName             string              `json:"WednesdayDomainName"`
	ThursdayPartyName               string              `json:"ThursdayPartyName"`
	ThursdayDomainName              string              `json:"ThursdayDomainName"`
	FridayPartyName                 string              `json:"FridayPartyName"`
	FridayDomainName                string              `json:"FridayDomainName"`
	SaturdayPartyName               string              `json:"SaturdayPartyName"`
	SaturdayDomainName              string              `json:"SaturdayDomainName"`
	SundayPartyName                 string              `json:"SundayPartyName"`
	SundayDomainName                string              `json:"SundayDomainName"`
	CompletionAction                string              `json:"CompletionAction"`
}

type TaskItem struct {
	Item1 bool   `json:"Item1"`
	Item2 string `json:"Item2"`
}

type ResinCount struct {
	Condensed int `json:"浓缩树脂"`
	Original  int `json:"原粹树脂"`
	Transient int `json:"须臾树脂"`
	Fragile   int `json:"脆弱树脂"`
}

// 使用循环遍历检查数字是否包含在数组中
func contains(slice []string, num int) bool {
	for _, v := range slice {
		vInt, _ := strconv.Atoi(v)
		if vInt == num {
			return true
		}
	}
	return false
}

// 读取
func (chaBaoBgiConfig *ChaBaoBgiConfig) ReadChaBaoBgiConfig(longName string) error {

	filename := filepath.Join(config.Cfg.BetterGIAddress, "User", "OneDragon", longName+".json")

	file, err := os.ReadFile(filename)
	if err != nil {

		return fmt.Errorf("读取文件失败:%s ", err)
	}

	if err := json.Unmarshal(file, &chaBaoBgiConfig); err != nil {

		return fmt.Errorf("解析 JSON 失败: %w", err)
	}

	now := time.Now()
	weekdayNum := int(now.Weekday())

	autoLog.Sugar.Infof("今天是: 星期%d", weekdayNum)

	// 示例输出部分内容
	fmt.Println("计划名称:", chaBaoBgiConfig.ScheduleName)
	fmt.Println("原神 UID:", chaBaoBgiConfig.GenshinUid)
	fmt.Println("启用的任务:")

	re := regexp.MustCompile(`\d+`) // 匹配一个或多个连续数字
	for id, task := range chaBaoBgiConfig.TaskEnabledList {

		numbers := re.FindAllString(task.Item2, -1)
		if numbers == nil {
			autoLog.Sugar.Infof("执行===  %s\n", task.Item2)

		} else {
			if contains(numbers, weekdayNum) {
				autoLog.Sugar.Infof("执行===  %s\n", task.Item2)
			} else {

				autoLog.Sugar.Infof("不执行===  %s\n", task.Item2)
				task.Item1 = false
			}
		}
		// 写回去
		chaBaoBgiConfig.TaskEnabledList[id] = task
	}

	return nil
}

func (chaBaoBgiConfig *ChaBaoBgiConfig) SaveChaBaoBgiConfig(longName string) error {

	filename := filepath.Join(config.Cfg.BetterGIAddress, "User", "OneDragon", longName+".json")

	// 格式化输出，方便人工看
	data, err := json.MarshalIndent(chaBaoBgiConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化 JSON 失败: %w", err)
	}

	// 写入文件（0644 表示普通读写权限）
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
}

func (chaBaoBgiConfig *ChaBaoBgiConfig) ChangeTaskEnabledList(longName string) error {

	err := chaBaoBgiConfig.ReadChaBaoBgiConfig(longName)
	if err != nil {
		return err
	}

	err = chaBaoBgiConfig.SaveChaBaoBgiConfig(longName)
	if err != nil {
		return err
	}
	return nil
}
