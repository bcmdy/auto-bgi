package mysConfig

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config 配置结构体
type MYSConfig struct {
	Account struct {
		Cookie string `yaml:"cookie"`
		Stuid  string `yaml:"stuid"`
		Stoken string `yaml:"stoken"`
		Mid    string `yaml:"mid"`
	} `yaml:"account"`
	Device struct {
		Name  string `yaml:"name"`
		Model string `yaml:"model"`
		ID    string `yaml:"id"`
		Fp    string `yaml:"fp"`
	} `yaml:"device"`
	Mihoyobbs struct {
		Enable      bool  `yaml:"enable"`
		Checkin     bool  `yaml:"checkin"`
		CheckinList []int `yaml:"checkin_list"`
		Read        bool  `yaml:"read"`
		Like        bool  `yaml:"like"`
		CancelLike  bool  `yaml:"cancel_like"`
		Share       bool  `yaml:"share"`
	} `yaml:"mihoyobbs"`
	Games struct {
		CN struct {
			Enable    bool   `yaml:"enable"`
			UserAgent string `yaml:"useragent"`
			Retries   int    `yaml:"retries"`
			Genshin   struct {
				Checkin   bool     `yaml:"checkin"`
				BlackList []string `yaml:"black_list"`
			} `yaml:"genshin"`
		} `yaml:"cn"`
	} `yaml:"games"`
}

var GlobalConfig MYSConfig

// LoadConfig 加载配置文件
func LoadConfig(configPath string) error {
	// 如果配置文件不存在，创建默认配置
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := createDefaultConfig(configPath); err != nil {
			return fmt.Errorf("创建默认配置文件失败: %v", err)
		}
	}

	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("读取配置文件失败: %v", err)
	}

	if err := viper.Unmarshal(&GlobalConfig); err != nil {
		return fmt.Errorf("解析配置文件失败: %v", err)
	}

	return nil
}

// createDefaultConfig 创建默认配置文件
func createDefaultConfig(configPath string) error {
	// 确保目录存在
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// 默认配置内容
	defaultConfig := `# 这里控制整个config文件是否启用
account:
  # 登入账号只需要修改cookie就行了
  cookie: ""
  stuid: ""
  stoken: ""
  mid: ""

device:
  name: "Xiaomi MI 6"
  model: "Mi 6"
  # 此处留空则脚本随机生成
  id: ""
  # 手动获取
  fp: ""

mihoyobbs:
  # 控制bbs功能是否启用
  enable: true
  # 社区签到
  checkin: true
  # 签到的社区列表
  checkin_list:
    - 5
    - 2
  # 看帖
  read: true
  # 点赞帖子
  like: true
  # 取消点赞
  cancel_like: true
  # 分享帖子
  share: true

# 游戏签到
games:
  # 国服控制区域
  cn:
    # 控制是否启用国内签到
    enable: true
    # 配置签到用的ua
    useragent: "Mozilla/5.0 (Linux; Android 12; Unspecified Device) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/103.0.5060.129 Mobile Safari/537.36"
    # 重试次数
    retries: 3
    # 原神
    genshin:
      # 控制是否启用签到
      checkin: true
      # 这里是不签到的账号
      black_list: []
`

	return os.WriteFile(configPath, []byte(defaultConfig), 0644)
}
