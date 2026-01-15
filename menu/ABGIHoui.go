package menu

import (
	"auto-bgi/abgiConstant"
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"github.com/tidwall/sjson"
	"os"
	"strings"
)

// 修改妙妙屋的abgi版本号
func ABGIHoui() {
	filePath := abgiConstant.JsPath("ABGIHoui", "settings.json")
	// 打开文件

	configData, err := os.ReadFile(filePath)
	if err != nil {
		autoLog.Sugar.Errorf("读取 妙妙屋失败")
		return
	}
	data := configData

	// 修改版本号
	Version := strings.TrimSpace(CurrentVersion)
	data, err2 := sjson.SetBytes(data, "3.default", Version)
	if err2 != nil {
		autoLog.Sugar.Errorf("修改失败:%d", err)
	}
	// 写回文件
	if err := os.WriteFile(filePath, data, 0644); err != nil {

		autoLog.Sugar.Errorf("写入 妙妙屋[%s]失败:%d", config.Cfg.Account.GouLangGroupName, err)

	}
}
