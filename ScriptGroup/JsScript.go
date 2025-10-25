package ScriptGroup

import (
	"auto-bgi/Notice"
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"fmt"
	"github.com/tidwall/sjson"
	"os"
)

type dd struct {
	ID   int64
	UID  string
	Name string
}

//var dogFood = ArtifactsBulkSupply.DogFood{}

// 启动狗粮联机
func (s *ScriptGroupConfig) StartDogFoodOnline(runDebug bool, data []map[string]interface{}) {

	// 解析 JSON 字符串
	var aa []dd
	yourIndex := 0
	RunningOrder := ""
	for i, item := range data {
		var d dd
		d.ID = item["ID"].(int64)
		d.UID = item["UID"].(string)
		d.Name = item["Name"].(string)
		if item["UID"] == config.Cfg.Account.Uid {
			yourIndex = int(d.ID)
		}
		if item["AbgiType"] == "正常跑" {
			RunningOrder += fmt.Sprintf("%d", i+1)
		}

		aa = append(aa, d)
	}

	filename := config.Cfg.BetterGIAddress + "\\User\\ScriptGroup\\" + config.Cfg.Account.GouLangGroupName + ".json"
	// 读取 狗粮联机JSON
	GouLangGroupData, err := os.ReadFile(filename)
	if err != nil {
		autoLog.Sugar.Errorf("读取 狗粮联机配置组[%s]失败:%d", config.Cfg.Account.GouLangGroupName, err)
	}

	tmp := s.ReadConfig(config.Cfg.Account.GouLangGroupName)

	newData := GouLangGroupData
	for i, proj := range tmp.Projects {
		if proj.FolderName == "ArtifactsGroupPurchasing" {
			// ✅ 修改指定项目下的 jsScriptSettingsObject
			path := fmt.Sprintf("projects.%d.jsScriptSettingsObject", i)

			//修改runningOrder
			newRunningOrder := fmt.Sprintf("%s.runningOrder", path)

			if RunningOrder == "" {
				RunningOrder = "1234"
			}
			newData, err = sjson.SetBytes(newData, newRunningOrder, RunningOrder)
			if err != nil {

				autoLog.Sugar.Errorf("修改runningOrder失败:%d", err)
			}

			//修改groupMode
			newGroupMode := fmt.Sprintf("%s.groupMode", path)
			newData, err = sjson.SetBytes(newData, newGroupMode, "按照下列配置自动进入并运行")
			if err != nil {

				autoLog.Sugar.Errorf("修改groupMode失败:%d", err)

			}

			//修改yourIndex
			newYourIndex := fmt.Sprintf("%s.yourIndex", path)
			newData, err = sjson.SetBytes(newData, newYourIndex, yourIndex)
			if err != nil {
				autoLog.Sugar.Errorf("修改yourIndex失败:%d", err)
			}
			Notice.SentText(fmt.Sprintf("是否是调机模式：%v", runDebug))

			//修改RunDebug
			newRunDebug := fmt.Sprintf("%s.runDebug", path)
			newData, err = sjson.SetBytes(newData, newRunDebug, runDebug)

			if err != nil {

				autoLog.Sugar.Errorf("修改runDebug失败:%d", err)
			}

			for i2, a := range aa {
				//修改uid
				newUid := fmt.Sprintf("%s.p%dUID", path, i2+1)
				newData, err = sjson.SetBytes(newData, newUid, a.UID)
				if err != nil {

					autoLog.Sugar.Errorf("修改修改uid失败:%d", err)
				}
				//修改name
				newName := fmt.Sprintf("%s.p%dName", path, i2+1)
				newData, err = sjson.SetBytes(newData, newName, a.Name)
				if err != nil {

					autoLog.Sugar.Errorf("修改name失败:%d", err)
				}
			}

		}
	}

	// 写回文件
	if err := os.WriteFile(filename, newData, 0644); err != nil {

		autoLog.Sugar.Errorf("写入 狗粮联机配置组[%s]失败:%d", config.Cfg.Account.GouLangGroupName, err)
	}

	//启动配置组
	err2 := startGroups([]string{config.Cfg.Account.GouLangGroupName})
	if err2 != nil {
		autoLog.Sugar.Errorf("启动配置组失败: %v", err)
		return
	}
}
