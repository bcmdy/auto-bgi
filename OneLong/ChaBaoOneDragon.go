package OneLong

import (
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"encoding/json"
	"fmt"
	"github.com/iancoleman/orderedmap"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"os"
)

type ChaBaoOneDragon struct {
	TaskEnabledList *orderedmap.OrderedMap `json:"TaskEnabledList"`
}

type Item struct {
	Item1 bool   `json:"Item1"`
	Item2 string `json:"Item2"`
}

func TestOneDragon(OneDragonName string) string {
	OneDragonFile := config.Cfg.BetterGIAddress + "\\User\\OneDragon\\" + OneDragonName + ".json"
	// 读取
	OneDragonFileData, err := os.ReadFile(OneDragonFile)
	if err != nil {
		autoLog.Sugar.Errorf("读取 狗粮联机配置组[%s]失败:%d", config.Cfg.Account.GouLangGroupName, err)
	}
	newData := OneDragonFileData
	TaskEnabledList := gjson.Get(string(newData), "TaskEnabledList")
	//
	orderedMap := orderedmap.OrderedMap{}
	err = json.Unmarshal([]byte(TaskEnabledList.Raw), &orderedMap)
	if err != nil {
		autoLog.Sugar.Errorf("读取 狗粮联机配置组[%s]失败:%d", config.Cfg.Account.GouLangGroupName, err)
	}
	for _, s := range orderedMap.Keys() {
		if s == "1" {
			get, _ := orderedMap.Get(s)
			//var Item Item
			//json.Unmarshal([]byte(get.(string)), &Item)
			o := get.(orderedmap.OrderedMap)
			Item2, _ := o.Get("Item2")
			if Item2 == "领取邮件" {
				newData, err = sjson.SetBytes(newData, "TaskEnabledList."+s+".Item1", false)
				if err != nil {
					autoLog.Sugar.Errorf("修改修改uid失败:%d", err)
				}

			}

		}

	}
	// 写回文件
	if err := os.WriteFile(OneDragonFile, newData, 0644); err != nil {

		autoLog.Sugar.Errorf("写入 狗粮联机配置组[%s]失败:%d", config.Cfg.Account.GouLangGroupName, err)
	}
	return ""

}

func WriteOneDragonTaskEnabledList(OneDragonName string, chaBaoOneDragon ChaBaoOneDragon) ChaBaoOneDragon {

	OneDragonFile := config.Cfg.BetterGIAddress + "\\User\\OneDragon\\" + OneDragonName + ".json"
	// 读取
	OneDragonFileData, err := os.ReadFile(OneDragonFile)
	if err != nil {
		autoLog.Sugar.Errorf("读取 狗粮联机配置组[%s]失败:%d", config.Cfg.Account.GouLangGroupName, err)
	}
	TaskEnabledList := gjson.Get(string(OneDragonFileData), "TaskEnabledList").Raw
	chaBaoOneDragon.TaskEnabledList = orderedmap.New()
	err = json.Unmarshal([]byte(TaskEnabledList), chaBaoOneDragon.TaskEnabledList)
	if err != nil {
		autoLog.Sugar.Errorf("读取 狗粮联机配置组[%s]失败:%d", config.Cfg.Account.GouLangGroupName, err)
	}

	var item Item
	item.Item1 = true
	item.Item2 = "领取邮件"

	chaBaoOneDragon.TaskEnabledList.Set("1", item)

	newData := OneDragonFileData

	newData, err = sjson.SetBytes(newData, "TaskEnabledList", chaBaoOneDragon.TaskEnabledList)
	if err != nil {
		fmt.Println(err)
		return ChaBaoOneDragon{}
	}

	// 关键：写回文件
	fmt.Println(string(newData))
	if err := os.WriteFile(OneDragonFile, newData, 0644); err != nil {
		autoLog.Sugar.Errorf("写回文件失败: %v", err)
		return ChaBaoOneDragon{}
	}

	return chaBaoOneDragon

}
