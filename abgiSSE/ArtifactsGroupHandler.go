package abgiSSE

import (
	"auto-bgi/ArtifactsBulkSupply"
	"auto-bgi/Notice"
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"auto-bgi/tools"
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var dogFood = ArtifactsBulkSupply.DogFood{}

func OnStart() {

	decrypt, err := tools.Decrypt(config.Cfg.Account.SecretKey)
	if err != nil {
		autoLog.Sugar.Infof("密钥错误")
		return
	}

	runDebug := false
	abgiType := "noDebug"
	//查询今天狗粮批发是什么线路
	dogFoodLine := dogFood.DogFoodIsAOrB()
	//读取联机文件
	num := ReadFile("dogFoodNum.txt")

	if dogFoodLine == "" {
		autoLog.Sugar.Errorf("查询今天狗粮批发线路失败")
		runDebug = true
		abgiType = "debug"
	} else if dogFoodLine == "B" {
		runDebug = true
		abgiType = "debug"
	} else if num >= 100 {
		autoLog.Sugar.Infof("今天狗粮已经跑完，无需再跑")
		runDebug = true
		abgiType = "debug"
	}

	autoLog.Sugar.Infof("当前狗粮批发线路为：%s，是否调试：%t，上线类型：%s", dogFoodLine, runDebug, abgiType)

	ConnectErr := Connect(fmt.Sprintf("ws://%s/api/abgiWs/%s/%s/%s", decrypt, abgiType, config.Cfg.Account.Uid, config.Cfg.Account.Name), runDebug, nil)
	if ConnectErr != nil {
		autoLog.Sugar.Infof("上线失败")
		return

	}
	autoLog.Sugar.Infof("上线成功")
}

// 调试上线
func OnStartDebug() {
	decrypt, err := tools.Decrypt(config.Cfg.Account.SecretKey)
	if err != nil {
		autoLog.Sugar.Infof("密钥错误")
		return
	}
	ConnectErr := Connect(fmt.Sprintf("ws://%s/api/abgiWs/%s/%s/%s", decrypt, "debug", config.Cfg.Account.Uid, config.Cfg.Account.Name), true, nil)
	if ConnectErr != nil {
		autoLog.Sugar.Infof("上线失败")
		return

	}
	autoLog.Sugar.Infof("上线成功")
}

// WriteDogFoodNum
func WriteDogFoodNum(num string) {
	dogFood.WriteDogFoodNum(num)

	//标记是否已经正常跑完联机
	b, _ := strconv.Atoi(num)
	if b >= 10 {
		//写入文件
		format := time.Now().Format("2006-01-02")
		WriteFile("dogFoodNum.txt", format+"--"+num)

	}
}

func WriteFile(fileName string, content string) {
	err := os.WriteFile(fileName, []byte(content), 0644)
	if err != nil {
		autoLog.Sugar.Errorf("写入文件失败：%s", err)
	} else {
		autoLog.Sugar.Infof("写入文件成功：%s", content)
	}
}

// 读取
func ReadFile(fileName string) int {
	data, err := os.ReadFile(fileName)
	if err != nil {
		autoLog.Sugar.Errorf("联机狗粮读取文件失败：%s", err)
		return 0
	} else {
		autoLog.Sugar.Infof("联机狗粮Num：%s", string(data))

		s := string(data)

		split := strings.Split(s, "--")
		if len(split) != 2 {
			return 0
		}
		s2 := split[0]
		format := time.Now().Format("2006-01-02")
		if s2 != format {
			return 0
		}
		b, _ := strconv.Atoi(split[1])
		return b
	}
}

// 判断联机狗粮版本是否是最新版
func IsNewestVersion() (is bool, newV, nowV string) {
	//读取公版版本号:https://cnb.cool/bettergi/bettergi-scripts-list/-/git/raw/main/repo/js/ArtifactsGroupPurchasing/manifest.json
	resp, err := http.Get("https://cnb.cool/bettergi/bettergi-scripts-list/-/git/raw/main/repo/js/ArtifactsGroupPurchasing/manifest.json")
	if err != nil {
		autoLog.Sugar.Errorf("读取版本号失败：%s", err)
		return false, "未知", "未知"
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		autoLog.Sugar.Error("读取响应失败:", err)

		return false, "未知", "未知"
	}

	newVersion := gjson.GetBytes(body, "version")

	//读取本地版本号
	scriptDir := filepath.Join(config.Cfg.BetterGIAddress, "User", "JsScript")
	//nowVersion, _ := bgiStatus.GetJsNowVersion(scriptDir, "ArtifactsGroupPurchasing")
	nowVersion, _ := readVersion(filepath.Join(scriptDir, "ArtifactsGroupPurchasing", "manifest.json"))

	//比较版本号
	if newVersion.String() == nowVersion {
		autoLog.Sugar.Infof("联机狗粮版本:当前版本号是最新版,可以上线")
	}
	if newVersion.String() != nowVersion {
		autoLog.Sugar.Errorf(fmt.Sprintf("联机狗粮版本错误,隔壁老王提醒你,你的版本是:%s,:最新版本是:%s, 当前版本号不是最新版,需要更新哦,亲", nowVersion, newVersion.String()))
		autoLog.Sugar.Errorf("联机狗粮版本错误:你的版本是:%s", nowVersion)
		autoLog.Sugar.Errorf("联机狗粮版本错误:最新版本是:%s", newVersion.String())
		Notice.SentText(fmt.Sprintf("联机狗粮版本错误,隔壁老王提醒你,你的版本是:%s,:最新版本是:%s, 当前版本号不是最新版,需要更新哦,亲", nowVersion, newVersion.String()))
		return false, newVersion.String(), nowVersion
	}

	return true, newVersion.String(), nowVersion
}

func readVersion(manifestPath string) (string, string) {
	//捕获异常
	defer func() {
		if r := recover(); r != nil {
			autoLog.Sugar.Warnf("捕获异常: %v", r)
			return
		}
	}()

	file, err := os.Open(manifestPath)
	if err != nil {
		autoLog.Sugar.Warnf("readVersion打开文件失败: %v", err)
		return "未知版本", "未知"
	}
	defer file.Close()

	var data map[string]interface{}
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		autoLog.Sugar.Warnf("解析JSON失败: %s %v", manifestPath, err)
		return "未知版本", data["name"].(string)
	}

	if version, ok := data["version"].(string); ok {

		return version, data["name"].(string)
	}
	return "未知版本", data["name"].(string)
}
