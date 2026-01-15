package auth

import (
	"auto-bgi/autoLog"
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
		SystemName string `mapstructure:"system_name"`
	} `mapstructure:"sys_config"`
}

var User AbgiUser

func InitAuth() {
	configFile := "abgiUser.yaml"

	// 1. 如果文件不存在，自动创建默认配置文件（你原来的默认内容）
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		defaultConfig := []byte(`auth:
  username: "abgi"
  password: "abgi"
  api_key: "abgi"
sys_config:
  system_name: "隔壁老王"
`)
		if err := os.WriteFile(configFile, defaultConfig, 0644); err != nil {
			panic(err)
		}
		fmt.Println("已生成默认 abgiUser.yaml")
	}

	// 2. 用 viper 加载配置
	viper.SetConfigFile(configFile)
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	// 3. 定义我们期望的默认值（当配置缺失时使用）
	defaults := map[string]interface{}{
		"auth.username":          "abgi",
		"auth.password":          "abgi",
		"auth.api_key":           "abgi",
		"sys_config.system_name": "隔壁老王",
	}

	// 4. 检查缺失并补上默认值
	changed := false
	for key, def := range defaults {
		if !viper.IsSet(key) {
			viper.Set(key, def)
			changed = true
			autoLog.Sugar.Infof("配置缺失，已设置默认: %s = %v", key, def)
		}
	}

	// 5. 如果有改动，写回配置文件（覆盖原文件）
	if changed {
		if err := viper.WriteConfig(); err != nil {
			// 如果 WriteConfig 失败（比如没有写权限或其他），尝试使用 WriteConfigAs 备选方案
			if err2 := viper.WriteConfigAs(configFile); err2 != nil {
				// 两个都失败则记录并继续（或根据需要 panic）
				fmt.Printf("写回配置失败: %v ; 尝试 WriteConfigAs 也失败: %v\n", err, err2)
			} else {
				fmt.Println("已把补全后的配置写回到文件（WriteConfigAs）")
			}
		} else {
			fmt.Println("已把补全后的配置写回到文件（WriteConfig）")
		}
	}

	// 6. 映射到结构体
	if err := viper.Unmarshal(&User); err != nil {
		panic(err)
	}

	// 注册 jwt
	SetSecret(User.Auth.Password)
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

	systemName := viper.GetString("sys_config.system_name")
	context.JSON(200, gin.H{
		"code":       200,
		"systemName": systemName + "のABGI",
	})

}
