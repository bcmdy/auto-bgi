package abgiSSE

import (
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"auto-bgi/tools"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

// 联机上线
func Online(c *gin.Context) {
	if config.Cfg.Account.Uid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "账号配置错误"})
		return
	}
	if config.Cfg.Account.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "账号配置错误"})
		return
	}
	if config.Cfg.Account.SecretKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "密钥错误"})
		return
	}
	//解密
	decryptedKey, err3 := tools.Decrypt(config.Cfg.Account.SecretKey)
	if err3 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "密钥错误"})
		return
	}

	//读取联机文件
	num := ReadFile("dogFoodNum.txt")

	runDebug := c.Query("runDebug") == "true"
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
	} else if num >= 100 {
		autoLog.Sugar.Infof("今天狗粮已经跑完，无需再跑")
		runDebug = true
		abgiType = "debug"
	} else if runDebug {
		abgiType = "debug"
	}

	autoLog.Sugar.Infof("当前狗粮批发线路为：%s，是否调试：%t，手动上线类型：%s", dogFoodLine, runDebug, abgiType)

	err := Connect(fmt.Sprintf("ws://%s/api/abgiWs/%s/%s/%s", decryptedKey, abgiType, config.Cfg.Account.Uid, config.Cfg.Account.Name), runDebug, nil)
	if err != nil {
		autoLog.Sugar.Errorf("连接失败")
		c.String(http.StatusBadRequest, err.Error())
		return

	}
	c.String(http.StatusOK, "连接成功")
}
