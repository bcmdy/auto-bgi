package config

import (
	"auto-bgi/autoLog"
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
	OneLong         oneLong      `json:"OneLong" comment:"一条龙配置"`
	BetterGIAddress string       `json:"BetterGIAddress" comment:"BetterGI地址"`
	Content         string       `json:"content" comment:"通知内容"`
	ConfigNames     []string     `json:"ConfigNames" comment:"一条龙配置名称"`
	BagStatistics   string       `json:"BagStatistics" comment:"需要统计的物品"`
	Post            string       `json:"post" comment:"post地址"`
	MySign          MySign       `json:"MySign" comment:"米游社签到"`
	Backups         []string     `json:"backups" comment:"需要的备份文件"`
	Cookie          string       `json:"cookie"`
	BasePath        string       `json:"basePath"`
	Control         Control      `json:"Control" comment:"控制配置"`
	LogKeywords     []string     `json:"LogKeywords" comment:"日志关键词"`
	OneRemote       OneRemote    `json:"OneRemote" comment:"1Remote配置"`
	ScreenRecord    ScreenRecord `json:"ScreenRecord" comment:"录屏配置"`
	BgiLog          string       `json:"BgiLog" comment:"bgi日志"`
	Notice          Notice       `json:"Notice" comment:"通知配置"`
	AIConfig        AIConfig     `json:"AIConfig" comment:"AI配置"`
}
type AIConfig struct {
	APIKey string `json:"APIKey" comment:"APIKey"`
	Model  string `json:"Model" comment:"模型"`
}

type Notice struct {
	Type     string   `json:"Type" comment:"通知类型"`
	Wechat   string   `json:"Wechat" comment:"企业微信webhook地址"`
	TGNotice TGNotice `json:"TGNotice" comment:"TG通知配置"`
}

type TGNotice struct {
	TGToken string `json:"TGToken" comment:"TG机器人token"`
	ChatID  int64  `json:"ChatID" comment:"TG聊天ID"`
	Proxy   string `json:"Proxy" comment:"TG代理"`
}

type ScreenRecord struct {
	IsRecord        bool   `json:"IsRecord" comment:"是否开启录屏"`
	ScriptGroupName string `json:"ScriptGroupName" comment:"配置组名称"`
}

type OneRemote struct {
	IsMonitor   bool     `json:"IsMonitor" comment:"是否开启1Remote监控"`
	LogFilePath string   `json:"LogFilePath" comment:"1Remote日志文件路径"`
	LogKeywords []string `json:"LogKeywords" comment:"1Remote日志关键词"`
}

type Control struct {
	IsCloseYuanShen bool `json:"IsCloseYuanShen" comment:"bgi关闭需要是否关闭原神"`
	SendWeChatImage bool `json:"SendWeChatImage" comment:"是否开启每隔一小时发送截图"`
}

type oneLong struct {
	IsStartTimeLong bool `json:"isStartTimeLong" comment:"是否开启一条龙"`
	OneLongHour     int  `json:"OneLongHour" comment:"一条龙小时"`
	OneLongMinute   int  `json:"OneLongMinute" comment:"一条龙分钟"`
}

type MySign struct {
	IsMySignIn bool   `json:"isMysSignIn" comment:"是否开启我的签到"`
	Url        string `json:"url" comment:"我的签到url"`
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

// ReloadConfig 重新加载配置文件
func ReloadConfig() error {

	file, err := os.Open("main.json")
	if err != nil {
		fmt.Println("ReloadConfig打开配置文件失败: %v", err)
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
	if len(Cfg.ConfigNames) != 7 {
		Cfg.ConfigNames = []string{"默认配置", "默认配置", "默认配置", "默认配置", "默认配置", "默认配置", "默认配置"}
	}

	if Cfg.OneRemote.LogKeywords == nil {
		Cfg.OneRemote.LogKeywords = []string{"OnRdpClientDisconnected"}
	}
	if Cfg.LogKeywords == nil {
		Cfg.LogKeywords = []string{"未识别到突发任务", "OCR 识别失败", "此路线出现3次卡死", "重试一次路线或放弃此路线！", "检测到复苏界面", "存在角色被击败", "执行路径时出错", "传送点未激活或不存在"}
	}

}

// 获取今天启动的一条龙名字
func GetTodayOneLongName() string {
	var oneLongs = Cfg.ConfigNames
	now := time.Now()
	weekdayNum := int(now.Weekday())
	autoLog.Sugar.Infof("今天是: 星期%d", weekdayNum)
	oneLongName := oneLongs[weekdayNum]
	return oneLongName
}

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
