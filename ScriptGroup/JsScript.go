package ScriptGroup

import (
	"auto-bgi/autoLog"
	"auto-bgi/config"
)

type dd struct {
	ID   int64
	UID  string
	Name string
}

// 启动狗粮联机
func (s *ScriptGroupConfig) StartDogFoodOnline(data []map[string]interface{}) {
	// 解析 JSON 字符串
	var aa []dd
	yourIndex := 0
	for _, item := range data {
		var d dd
		d.ID = item["ID"].(int64)
		d.UID = item["UID"].(string)
		d.Name = item["Name"].(string)
		if item["UID"] == config.Cfg.Account.UID {
			yourIndex = int(d.ID)
		}
		aa = append(aa, d)
	}

	//修改狗粮配置
	readConfig := s.ReadConfig(config.Cfg.Account.GouLangGroupName)
	projects := readConfig.Projects
	for _, project := range projects {
		if project.Type == "Javascript" && project.FolderName == "ArtifactsGroupPurchasing" {
			object := project.JsScriptSettingsObject
			object["yourIndex"] = yourIndex
			object["p1UID"] = aa[0].UID
			object["p1Name"] = aa[0].Name
			object["p2UID"] = aa[1].UID
			object["p2Name"] = aa[1].Name
			object["p3UID"] = aa[2].UID
			object["p3Name"] = aa[2].Name
			object["p4UID"] = aa[3].UID
			object["p4Name"] = aa[3].Name
			project.JsScriptSettingsObject = object
		}
	}

	//保存配置
	err := s.SaveConfig(config.Cfg.Account.GouLangGroupName, readConfig)
	if err != nil {
		autoLog.Sugar.Errorf("保存配置失败: %v", err)
		return
	}

	//启动配置组
	err = startGroups([]string{config.Cfg.Account.GouLangGroupName})
	if err != nil {
		autoLog.Sugar.Errorf("启动配置组失败: %v", err)
		return
	}
}
