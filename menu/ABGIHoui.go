package menu

import (
	"auto-bgi/abgiConstant"
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"fmt"
	"github.com/tidwall/sjson"
	"log"
	"os"
	"regexp"
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
	// 1. 获取文件路径
	filePath := abgiConstant.JsPath("ABGIHoui", "main.js")

	// 2. 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("读取文件失败: %v", err)
		return
	}

	// 3. 正则表达式逻辑
	// 匹配 localhost: 后面跟着的 1 到 5 位数字
	// `localhost:\d+` 也可以，但 `\d{1,5}` 更严谨（对应合法端口位）
	re := regexp.MustCompile(`localhost:\d{1,5}`)

	// 构造新的目标字符串，例如 "localhost:9000"
	target := fmt.Sprintf("localhost:%s", newPort)

	// 执行正则替换
	newContent := re.ReplaceAllString(string(content), target)

	// 4. 写回文件
	err = os.WriteFile(filePath, []byte(newContent), 0644)
	if err != nil {
		log.Printf("写入文件失败: %v", err)
		return
	}

	fmt.Printf("✅ 已成功将端口号动态修改为: %s\n", newPort)

}
