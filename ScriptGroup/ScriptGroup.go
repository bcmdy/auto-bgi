package ScriptGroup

import (
	"auto-bgi/config"
	"auto-bgi/tools"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ScriptGroupConfig struct {
	Index    int       `json:"index"`
	Name     string    `json:"name"`
	Config   Config    `json:"config"`
	Projects []Project `json:"projects"`
}

type Config struct {
	PathingConfig     PathingConfig `json:"pathingConfig"`
	ShellConfig       ShellConfig   `json:"shellConfig"`
	EnableShellConfig bool          `json:"enableShellConfig"`
}

type PathingConfig struct {
	Enabled                              bool                   `json:"enabled"`
	AutoPickEnabled                      bool                   `json:"autoPickEnabled"`
	PartyName                            string                 `json:"partyName"`
	IsVisitStatueBeforeSwitchParty       bool                   `json:"isVisitStatueBeforeSwitchParty"`
	MainAvatarIndex                      string                 `json:"mainAvatarIndex"`
	GuardianAvatarIndex                  string                 `json:"guardianAvatarIndex"`
	GuardianElementalSkillSecondInterval string                 `json:"guardianElementalSkillSecondInterval"`
	GuardianElementalSkillLongPress      bool                   `json:"guardianElementalSkillLongPress"`
	OnlyInTeleportRecover                bool                   `json:"onlyInTeleportRecover"`
	JsScriptUseEnabled                   bool                   `json:"jsScriptUseEnabled"`
	SoloTaskUseFightEnabled              bool                   `json:"soloTaskUseFightEnabled"`
	SkipDuring                           string                 `json:"skipDuring"`
	UseGadgetIntervalMs                  int                    `json:"useGadgetIntervalMs"`
	AutoSkipEnabled                      bool                   `json:"autoSkipEnabled"`
	AutoRunEnabled                       bool                   `json:"autoRunEnabled"`
	AutoEatEnabled                       bool                   `json:"autoEatEnabled"`
	AutoEatConfig                        AutoEatConfig          `json:"autoEatConfig"`
	HideOnRepeat                         bool                   `json:"hideOnRepeat"`
	TaskCycleConfig                      TaskCycleConfig        `json:"taskCycleConfig"`
	TaskCompletionSkipRuleConfig         TaskCompletionSkipRule `json:"taskCompletionSkipRuleConfig"`
	PreExecutionPriorityConfig           PreExecutionPriority   `json:"preExecutionPriorityConfig"`
	AutoFightEnabled                     bool                   `json:"autoFightEnabled"`
	AutoFightConfig                      AutoFightConfig        `json:"autoFightConfig"`
}

type AutoEatConfig struct {
	Enabled                    bool    `json:"enabled"`
	ShowNotification           bool    `json:"showNotification"`
	CheckInterval              int     `json:"checkInterval"`
	EatInterval                int     `json:"eatInterval"`
	TestFoodName               *string `json:"testFoodName"`
	DefaultAtkBoostingDishName *string `json:"defaultAtkBoostingDishName"`
	DefaultAdventurersDishName *string `json:"defaultAdventurersDishName"`
	DefaultDefBoostingDishName *string `json:"defaultDefBoostingDishName"`
}

type TaskCycleConfig struct {
	Enable       bool `json:"enable"`
	BoundaryTime int  `json:"boundaryTime"`
	Cycle        int  `json:"cycle"`
	Index        int  `json:"index"`
}

type TaskCompletionSkipRule struct {
	Enable            bool   `json:"enable"`
	SkipPolicy        string `json:"skipPolicy"`
	BoundaryTime      int    `json:"boundaryTime"`
	LastRunGapSeconds int    `json:"lastRunGapSeconds"`
	ReferencePoint    string `json:"referencePoint"`
}

type PreExecutionPriority struct {
	Enabled       bool   `json:"enabled"`
	GroupNames    string `json:"groupNames"`
	MaxRetryCount int    `json:"maxRetryCount"`
}

type AutoFightConfig struct {
	StrategyName               string             `json:"strategyName"`
	TeamNames                  string             `json:"teamNames"`
	FightFinishDetectEnabled   bool               `json:"fightFinishDetectEnabled"`
	ActionSchedulerByCd        string             `json:"actionSchedulerByCd"`
	OnlyPickEliteDropsMode     string             `json:"onlyPickEliteDropsMode"`
	FinishDetectConfig         FinishDetectConfig `json:"finishDetectConfig"`
	PickDropsAfterFightEnabled bool               `json:"pickDropsAfterFightEnabled"`
	PickDropsAfterFightSeconds int                `json:"pickDropsAfterFightSeconds"`
	BattleThresholdForLoot     *string            `json:"battleThresholdForLoot"`
	KazuhaPickupEnabled        bool               `json:"kazuhaPickupEnabled"`
	GuardianAvatar             string             `json:"guardianAvatar"`
	GuardianCombatSkip         bool               `json:"guardianCombatSkip"`
	SkipModel                  bool               `json:"skipModel"`
	GuardianAvatarHold         bool               `json:"guardianAvatarHold"`
	KazuhaPartyName            string             `json:"kazuhaPartyName"`
	Timeout                    int                `json:"timeout"`
}

type FinishDetectConfig struct {
	BattleEndProgressBarColor          string `json:"battleEndProgressBarColor"`
	BattleEndProgressBarColorTolerance string `json:"battleEndProgressBarColorTolerance"`
	FastCheckEnabled                   bool   `json:"fastCheckEnabled"`
	RotateFindEnemyEnabled             bool   `json:"rotateFindEnemyEnabled"`
	FastCheckParams                    string `json:"fastCheckParams"`
	CheckEndDelay                      string `json:"checkEndDelay"`
	BeforeDetectDelay                  string `json:"beforeDetectDelay"`
}

type ShellConfig struct {
	Disable  bool `json:"disable"`
	Timeout  int  `json:"timeout"`
	NoWindow bool `json:"noWindow"`
	Output   bool `json:"output"`
}

type Project struct {
	Name                   string                 `json:"name"`
	FolderName             string                 `json:"folderName"`
	JsScriptSettingsObject map[string]interface{} `json:"jsScriptSettingsObject"`
	Index                  int                    `json:"index"`
	Type                   string                 `json:"type"`
	Status                 string                 `json:"status"`
	Schedule               string                 `json:"schedule"`
	RunNum                 int                    `json:"runNum"`
	AllowJsNotification    bool                   `json:"allowJsNotification"`
}

//type JsScriptSettings struct {
//	Notice       bool   `json:"notice"`
//	SmithyName   string `json:"smithyName"`
//	ForgedOrNot  string `json:"forgedOrNot"`
//	Ore          string `json:"ore"`
//	SecondaryOre string `json:"secondaryOre"`
//	TertiaryOre  string `json:"tertiaryOre"`
//}

func (s *ScriptGroupConfig) ListScriptGroup() []string {
	entries, err := os.ReadDir(config.Cfg.BetterGIAddress + "\\User\\ScriptGroup")
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

type AllPath struct {
	FileName      string    `json:"fileName"`
	FileNameChild []AllPath `json:"fileNameChild"` // 改成嵌套结构
}

func (s *ScriptGroupConfig) ListAllPathing() ([]AllPath, error) {
	repoDir := filepath.Join(config.Cfg.BetterGIAddress, "Repos", "bettergi-scripts-list-git", "repo", "pathing")

	allPaths, err := getAllDirectories(repoDir)
	if err != nil {
		return nil, err
	}
	return allPaths, nil
}

// 递归遍历目录
func getAllDirectories(root string) ([]AllPath, error) {
	directories, err := tools.ListDirectories(root)
	if err != nil {
		return nil, fmt.Errorf("获取目录失败 (%s): %w", root, err)
	}

	var result []AllPath
	for _, dir := range directories {
		fullPath := filepath.Join(root, dir)

		children, err := getAllDirectories(fullPath) // 递归
		if err != nil {
			fmt.Printf("获取子目录失败 (%s): %v\n", fullPath, err)
			continue
		}

		result = append(result, AllPath{
			FileName:      dir,
			FileNameChild: children,
		})
	}

	return result, nil
}

func (s *ScriptGroupConfig) SavePathing(path []config.UpdatePathing) error {

	config.Cfg.UpdatePath = path

	err := config.WriteConfig()
	if err != nil {
		return err
	}
	return nil

}
