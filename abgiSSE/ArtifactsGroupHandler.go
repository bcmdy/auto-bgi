package abgiSSE

import (
	"auto-bgi/ArtifactsBulkSupply"
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"fmt"
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
	if dogFoodLine == "" {
		autoLog.Sugar.Errorf("查询今天狗粮批发线路失败")
		runDebug = true
		abgiType = "debug"
	} else if dogFoodLine == "B" {
		runDebug = true
		abgiType = "debug"
	} else if runDebug {
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
}
