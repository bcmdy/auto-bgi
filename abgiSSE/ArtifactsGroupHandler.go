package abgiSSE

import (
	"auto-bgi/ArtifactsBulkSupply"
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

var dogFood = ArtifactsBulkSupply.DogFood{}

func OnStart() {

	decrypt, err := Decrypt(config.Cfg.Account.SecretKey, config.Cfg.Account.AccountKey)
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
	decrypt, err := Decrypt(config.Cfg.Account.SecretKey, config.Cfg.Account.AccountKey)
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
