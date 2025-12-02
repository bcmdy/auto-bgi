package auth

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"os"
)

type AbgiUser struct {
	Auth struct {
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
		APIKey   string `mapstructure:"api_key"`
	} `mapstructure:"auth"`

	SysConfig struct {
		SystemName string `mapstructure:"systemName"`
	} `mapstructure:"sysConfig"`
}

// 3. 映射到结构体
var User AbgiUser

func InitAuth() {
	configFile := "abgiUser.yaml"

	// 1. 如果文件不存在，自动创建默认配置
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		defaultConfig := []byte(`auth:
  username: "abgi"
  password: "abgi"
  api_key: "abgi"
sysConfig:
  systemName: "隔壁老王"
`)
		err := os.WriteFile(configFile, defaultConfig, 0644)
		if err != nil {
			panic(err)
		}
		fmt.Println("已生成默认 abgiUser.yaml")
	}

	// 2. 用 viper 加载配置
	viper.SetConfigFile(configFile)

	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	if err := viper.Unmarshal(&User); err != nil {
		panic(err)
	}

}

// 登录获取token
func login(username string, password string) (string, error) {

	if username == viper.GetString("auth.username") && password == viper.GetString("auth.password") {
		//生成token
		tokenString, err := MakeToken(viper.GetString("auth.username") + viper.GetString("auth.password"))
		if err != nil {
			return "", err
		}
		return tokenString, nil
	}
	// 登录失败
	return "登录失败", fmt.Errorf("登录失败")
}

func GetSystemConfig(context *gin.Context) {

	systemName := viper.GetString("sysConfig.systemName")
	context.JSON(200, gin.H{
		"code":       200,
		"systemName": systemName + "-aBgi",
	})

}
