package menu

import (
	"auto-bgi/abgiConstant"
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"fmt"
	"github.com/tidwall/gjson"
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

func ABGIHouiPort(newPort string) {
	newPort = strings.ReplaceAll(newPort, ":", "")

	// 1. 获取文件路径
	filePath := abgiConstant.JsPath("ABGIHoui-API", "manifest.json")

	// 2. 读取文件内容
	data, err := os.ReadFile(filePath)
	if err != nil {
		autoLog.Sugar.Warnf("读取 妙妙屋API失败")
		return
	}

	// 3. 获取旧的 URL 列表
	urls := gjson.GetBytes(data, "http_allowed_urls")
	if !urls.Exists() || !urls.IsArray() {
		return
	}

	updatedData := data
	hasChanged := false

	// 4. 遍历并替换端口
	urls.ForEach(func(key, value gjson.Result) bool {
		oldUrl := value.String()

		if strings.Contains(oldUrl, "localhost:") {

			parts := strings.Split(oldUrl, "/")
			if len(parts) > 2 && strings.Contains(parts[2], ":") {
				hostPart := strings.Split(parts[2], ":")
				if len(hostPart) == 2 && hostPart[1] != newPort {
					newUrl := strings.Replace(oldUrl, ":"+hostPart[1], ":"+newPort, 1)

					// 5. 使用 sjson 更新内存中的字节流
					path := "http_allowed_urls." + key.String()
					if temp, err := sjson.SetBytes(updatedData, path, newUrl); err == nil {
						updatedData = temp
						hasChanged = true
					}
				}
			}
		}
		return true
	})
	// 6. 如果有改动，同步回文件
	if hasChanged {
		if err := os.WriteFile(filePath, updatedData, 0644); err != nil {
			fmt.Printf("更新端口失败: %v\n", err)
			autoLog.Sugar.Infof("更新端口失败: %v\n", err)
		}
	}
}
