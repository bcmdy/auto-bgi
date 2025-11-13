package config

import (
	"encoding/json"
	"fmt"
	"github.com/robfig/cron/v3"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type Config struct {
	OneLong         oneLong         `json:"OneLong" comment:"一条龙配置"`
	BetterGIAddress string          `json:"BetterGIAddress" comment:"BetterGI地址"`
	RepoUrl         string          `json:"RepoUrl" comment:"仓库地址"`
	Content         string          `json:"content" comment:"通知内容"`
	BagStatistics   string          `json:"BagStatistics" comment:"需要统计的物品"`
	Post            string          `json:"post" comment:"post地址"`
	BasePath        string          `json:"basePath"`
	Control         Control         `json:"Control" comment:"控制配置"`
	LogKeywords     []string        `json:"LogKeywords" comment:"日志关键词"`
	OneRemote       OneRemote       `json:"OneRemote" comment:"1Remote配置"`
	ScreenRecord    ScreenRecord    `json:"ScreenRecord" comment:"录屏配置"`
	BgiLog          string          `json:"BgiLog" comment:"bgi日志"`
	Notice          Notice          `json:"Notice" comment:"通知配置"`
	CommandBot      CommandBot      `json:"CommandBot" comment:"命令机器人配置"`
	UpdatePath      []UpdatePathing `json:"UpdatePath" comment:"地图追踪更新配置"`
	Account         Account         `json:"Account" comment:"账号配置"`
	AbgiAiConfig    AbgiAiConfig    `json:"AbgiAiConfig" comment:"abgiAi配置"`
}
type Account struct {
	Uid                string `json:"Uid" comment:"账号UID"`
	Name               string `json:"Name" comment:"账号名称"`
	IsMultiUser        bool   `json:"IsMultiUser" comment:"是否是多用户"`
	GouLangGroupName   string `json:"GouLangGroupName" comment:"狗粮联机配置组名称"`
	OnlineKeyword      string `json:"OnlineKeyword" comment:"联机上线关键词"`
	OnlineAfterKeyword string `json:"OnlineAfterKeyword" comment:"联机之后关键词"`
	OnlineAfterOneLong string `json:"OnlineAfterOneLong" comment:"联机之后一条龙"`
	SecretKey          string `json:"SecretKey" comment:"加密密钥"`
	AccountKey         string `json:"AccountKey" comment:"密钥"`
}
type UpdatePathing struct {
	Name       string `json:"name" comment:"配置组"`
	FolderName string `json:"folderName" comment:"地图追踪文件夹名称"`
}
type Notice struct {
	Type     string   `json:"Type" comment:"通知类型"`
	Wechat   string   `json:"Wechat" comment:"企业微信webhook地址"`
	TGNotice TGNotice `json:"TGNotice" comment:"TG通知配置"`
	OneBot   OneBot   `json:"OneBot" comment:"OneBot配置"`
	FeiShu   FeiShu   `json:"FeiShu" comment:"飞书配置"`
}

type FeiShu struct {
	FeiShuWebhookURL string `json:"FeiShuWebhookURL" comment:"飞书webhook地址"`
	AppID            string `json:"AppID" comment:"飞书AppID"`
	AppSecret        string `json:"AppSecret" comment:"飞书AppSecret"`
}

// OneBot 封装配置
type OneBot struct {
	APIBase  string // OneBot API 地址，例如 http://127.0.0.1:5700
	Token    string // 可选 Token，用于鉴权
	QQNum    int    `json:"QQNum"`    // QQ 号
	GroupNum int    `json:"groupNum"` // 群号
}

type TGNotice struct {
	TGToken string `json:"TGToken" comment:"TG机器人token"`
	ChatID  int64  `json:"ChatID" comment:"TG聊天ID"`
	Proxy   string `json:"Proxy" comment:"TG代理"`
}

type ScreenRecord struct {
	IsRecord    bool   `json:"IsRecord" comment:"是否开启录屏"`
	StartScreen string `json:"StartScreen" comment:"开始录屏关键字"`
	EndScreen   string `json:"EndScreen" comment:"结束录屏关键字"`
	ObsSavePath string `json:"ObsSavePath" comment:"OBS录屏保存路径"`
}

type OneRemote struct {
	IsMonitor   bool     `json:"IsMonitor" comment:"是否开启1Remote监控"`
	LogFilePath string   `json:"LogFilePath" comment:"1Remote日志文件路径"`
	LogKeywords []string `json:"LogKeywords" comment:"1Remote日志关键词"`
}

type Control struct {
	IsCloseYuanShen  bool `json:"IsCloseYuanShen" comment:"bgi关闭需要是否关闭原神"`
	BackupUsersHour  int  `json:"BackupUsersHour" comment:"每隔几个小时备份users文件夹"`
	SendWeChatImage  bool `json:"SendWeChatImage" comment:"是否开启每隔一小时发送截图"`
	StartOpenBrowser bool `json:"StartOpenBrowser" comment:"是否开启启动时打开浏览器"`
	AbgiScreen       bool `json:"AbgiScreen" comment:"是否开启bgi实时屏幕"`
	OBSReplayBuffer  bool `json:"OBSReplayBuffer" comment:"一条龙启动是否开启OBS重放缓冲"`
}

type oneLong struct {
	AutoUpdateJs bool `json:"AutoUpdateJs" comment:"是否开启自动更新js"`
}

type CommandBot struct {
	TgBOT     bool `json:"TgBOT" comment:"是否开启TG机器人"`
	FeiShuBot bool `json:"FeiShuBot" comment:"是否开启飞书机器人"`
}

type AbgiAiConfig struct {
	IsAbgiAi bool   `json:"IsAbgiAi" comment:"是否开启bgiai"`
	ApiKey   string `json:"ApiKey" comment:"密钥"`
	BaseURL  string `json:"BaseUrl" comment:"地址"`
	Model    string `json:"Model" comment:"模型"`
}

var Cfg Config
var Parser = cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)

func init() {
	err := ReloadConfig()
	if err != nil {
		//autoLog.Sugar.Fatalf("首次加载配置失败: %v", err)
		fmt.Println("首次加载配置失败: %v", err)
	}
}

// WriteConfig 重新写入main.json
func WriteConfig() error {
	// 序列化为JSON字符串，格式化输出
	data, err := json.MarshalIndent(Cfg, "", "  ")
	if err != nil {
		fmt.Println("writeConfig序列化失败: %v", err)
		return err
	}

	// 写入 main.json，路径可以自定义，这里示例写当前运行目录
	filePath := filepath.Join(".", "main.json")
	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		fmt.Println("writeConfig写文件失败: %v", err)
		return err
	}
	return nil
}

// ReloadConfig 重新加载配置文件
func ReloadConfig() error {

	file, err := os.Open("main.json")
	if err != nil {
		fmt.Println("ReloadConfig打开配置文件失败: %v", err)
		//创建配置文件
		Cfg = Config{}
		WriteConfig()
		return err
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {

		fmt.Println("ReloadConfig读取配置文件失败: %v", err)
		return err
	}

	if err := json.Unmarshal(bytes, &Cfg); err != nil {

		fmt.Println("ReloadConfig解析配置文件失败: %v", err)
		return err
	}

	// 更新 BasePath 为当前可执行文件目录
	ex, err := os.Executable()
	if err != nil {
		log.Printf("无法获取可执行文件路径: %v", err)
		// 不返回错误，继续执行
	} else {
		Cfg.BasePath = filepath.Dir(ex)
	}

	//读取bgi日志
	logDir := filepath.Clean(fmt.Sprintf("%s\\log", Cfg.BetterGIAddress))
	files, err := FindLogFiles(logDir)
	if len(files) == 0 {
		Cfg.BgiLog = "无"
	} else {
		Cfg.BgiLog = files[0]
	}
	DefaultConfig()

	//重新写入
	// 写入 main.json，路径可以自定义，这里示例写当前运行目录
	filePath := filepath.Join(".", "main.json")
	data, err := json.MarshalIndent(Cfg, "", "  ")
	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		fmt.Println("ReloadConfig写文件失败: %v", err)
		return err
	}

	return nil
}

// 配置验证补全
func DefaultConfig() {

	if Cfg.OneRemote.LogKeywords == nil {
		Cfg.OneRemote.LogKeywords = []string{"OnRdpClientDisconnected"}
	}
	if Cfg.LogKeywords == nil {
		Cfg.LogKeywords = []string{"未识别到突发任务", "OCR 识别失败", "此路线出现3次卡死", "重试一次路线或放弃此路线！", "检测到复苏界面", "存在角色被击败", "执行路径时出错", "传送点未激活或不存在"}
	}
	if Cfg.Notice.Type == "" {
		Cfg.Notice.Type = "Wechat"
	}

	if Cfg.Control.BackupUsersHour == 0 {
		Cfg.Control.BackupUsersHour = 72
	}

	if Cfg.UpdatePath == nil {
		Cfg.UpdatePath = []UpdatePathing{}
	}

}

// 获取今天启动的一条龙名字

func FindLogFiles(dirPath string) ([]string, error) {
	pattern := filepath.Join(dirPath, "*.log")

	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	// 保存文件名和时间
	type fileInfo struct {
		name string
		time time.Time
	}

	var fileInfos []fileInfo
	for _, f := range files {
		info, err := os.Stat(f)
		if err != nil {
			continue // 读取失败跳过
		}
		fileInfos = append(fileInfos, fileInfo{
			name: filepath.Base(f),
			time: info.ModTime(),
		})
	}

	// 按时间倒序排序
	sort.Slice(fileInfos, func(i, j int) bool {
		return fileInfos[i].time.After(fileInfos[j].time)
	})

	// 只返回文件名
	var filenames []string
	for _, fi := range fileInfos {
		filenames = append(filenames, fi.name)
	}

	return filenames, nil
}
