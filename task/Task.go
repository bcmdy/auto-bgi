package task

import (
	"auto-bgi/Notice"
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"auto-bgi/control"
	"auto-bgi/internal/gamecheckin"
	"auto-bgi/internal/mihoyobbs"
	"auto-bgi/internal/mysConfig"
	"auto-bgi/internal/utils"
	"encoding/json"
	"fmt"
	"github.com/robfig/cron/v3"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// 使用循环遍历检查数字是否包含在数组中
func contains(slice []string, num int) bool {
	for _, v := range slice {
		vInt, _ := strconv.Atoi(v)
		if vInt == num {
			return true
		}
	}
	return false
}

func calculateExecutionDay(boundaryTime int, cycle int) int {
	// 获取当前日期和时间
	now := time.Now()
	// 获取当前日期的年、月、日
	year, month, day := now.Date()
	// 计算从分界时间开始的当天时间
	boundaryDateTime := time.Date(year, month, day, boundaryTime, 0, 0, 0, time.Local)
	// 如果当前时间小于分界时间，则算前一天的
	if now.Before(boundaryDateTime) {
		// 计算前一天的日期
		previousDay := now.AddDate(0, 0, -1)
		year, month, day = previousDay.Date()
	}
	// 获取分界日期对象（当天或前一天）
	boundaryDate := time.Date(year, month, day, 0, 0, 0, 0, time.Local)
	// 计算从分界时间开始的天数（基于某个起始日期，这里假设起始日期为2025-01-01）
	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.Local)
	deltaDays := int(boundaryDate.Sub(startDate).Hours() / 24)
	// 计算执行序号
	executionSequence := (deltaDays % cycle) + 1
	return executionSequence
}

type TaskCycleConfig struct {
	Name         string
	Cycle        float64
	BoundaryTime float64
	Enable       bool
	Index        float64
	Mark         string
}

// 计算配置组今日是否执行
func CalculateTaskEnabledList() ([]TaskCycleConfig, error) {
	//读取目录下所有的json文件
	dir := config.Cfg.BetterGIAddress + "\\User\\ScriptGroup"
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return []TaskCycleConfig{}, err
	}
	var TaskCycleConfigs []TaskCycleConfig

	// 遍历目录中的所有文件
	for _, file := range files {
		// 检查文件是否为 JSON 文件
		if filepath.Ext(file.Name()) == ".json" {
			// 构建完整的文件路径
			filePath := filepath.Join(dir, file.Name())
			//fmt.Println("正在读取文件:", filePath)
			// 打开 JSON 文件
			configFile, err2 := os.Open(filePath) // 假设 JSON 文件名为 mysConfig.json
			if err2 != nil {
				return []TaskCycleConfig{}, err2
			}
			defer configFile.Close()
			// 读取文件内容
			byteValue, err3 := ioutil.ReadAll(configFile)
			if err3 != nil {
				fmt.Println("Failed to read JSON file:", err)
				return []TaskCycleConfig{}, err3
			}
			// 定义一个 map 来解析 JSON 数据
			var result map[string]interface{}

			// 解析 JSON 数据到 map
			err = json.Unmarshal(byteValue, &result)
			if err != nil {
				fmt.Println("Failed to unmarshal JSON data:", err)
				return []TaskCycleConfig{}, err
			}
			// 获取 taskCycleConfig 内容
			// 需要逐步深入嵌套的 map
			pathingConfig, ok := result["mysConfig"].(map[string]interface{})["pathingConfig"].(map[string]interface{})
			if !ok {
				fmt.Println("Failed to get pathingConfig")
				return []TaskCycleConfig{}, fmt.Errorf("Failed to get pathingConfig")
			}
			taskCycleConfig, ok := pathingConfig["taskCycleConfig"].(map[string]interface{})
			if !ok {
				fmt.Println("Failed to get taskCycleConfig")
				return []TaskCycleConfig{}, fmt.Errorf("Failed to get taskCycleConfig")
			}

			var data = TaskCycleConfig{}
			data.Name = file.Name()
			data.Enable = taskCycleConfig["enable"].(bool)
			data.BoundaryTime = taskCycleConfig["boundaryTime"].(float64)
			data.Cycle = taskCycleConfig["cycle"].(float64)
			data.Index = taskCycleConfig["index"].(float64)

			if data.Enable == true {
				data.Mark = "今日执行"
			} else {
				data.Mark = "今日不执行"
				day := calculateExecutionDay(int(data.BoundaryTime), int(data.Cycle))
				if day == int(data.Index) {
					data.Mark = "今日执行"
				} else {
					data.Mark = "今日不执行"
				}
			}
			TaskCycleConfigs = append(TaskCycleConfigs, data)
		}
	}

	return TaskCycleConfigs, nil
}

func MysSignIn() {

	MySignTime := config.Cfg.MySign.Time
	MySignTime = strings.TrimSpace(MySignTime)
	timeList := strings.Split(MySignTime, ",")

	cronTab := cron.New(cron.WithSeconds())

	one, _ := strconv.Atoi(timeList[1])
	ling, _ := strconv.Atoi(timeList[0])

	// 定时任务,cron表达式
	spec := fmt.Sprintf("0 %d %d * * *", one, ling)

	// 定义定时器调用的任务函数
	task := func() {
		fmt.Print("米游社签到服务启动", time.Now().Format("2006-01-02 15:04:05"))

		//mysConfig.GenShinSign()

		MiYouSheSign()

		//err := control.HttpGet(config.Cfg.MySign.Url + "/qd")
		//if err != nil {
		//
		//	autoLog.Sugar.Error("签到失败:", err)
		//	return
		//}

		time.Sleep(1000 * time.Millisecond)

		schedule, err := config.Parser.Parse(spec)
		if err != nil {

			autoLog.Sugar.Error("解析失败:", err)
			return
		}

		autoLog.Sugar.Infof("米游社签到服务启动完毕,下次执行时间: %s", schedule.Next(time.Now()).Format("2006-01-02 15:04:05"))
	}

	// 添加定时任务
	cronTab.AddFunc(spec, task)
	// 启动定时器
	cronTab.Start()
	// 阻塞主线程停止
	select {}
}

func ListGroups() ([]string, error) {
	// 指定要读取的文件夹路径
	//自定义配置路径
	folderPath := config.Cfg.BetterGIAddress + "\\User\\ScriptGroup"

	var groupNames []string

	// 遍历文件夹
	err := filepath.WalkDir(folderPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// 检查是否是 JSON 文件
		if filepath.Ext(d.Name()) == ".json" {
			// 打印 JSON 文件名（相对于文件夹的路径）
			relativePath, err := filepath.Rel(folderPath, path)
			if err != nil {
				return err
			}

			name := strings.Replace(relativePath, ".json", "", -1)

			groupNames = append(groupNames, name)

		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return groupNames, nil
}

// StartGroups 启动配置组
func StartGroups(names []string) error {
	control.CloseSoftware()
	time.Sleep(5 * time.Second)

	betterGIPath := filepath.Join(config.Cfg.BetterGIAddress, "BetterGI.exe")

	// 检查文件是否存在
	if _, err := os.Stat(betterGIPath); err != nil {
		autoLog.Sugar.Errorf("BetterGI.exe 不存在: %v", err)
		return err
	}

	args := append([]string{"--startGroups"}, names...) // 每个组名单独参数

	cmd := exec.Command(betterGIPath, args...)

	cmd.Stdout = nil
	cmd.Stderr = nil
	err := cmd.Start()
	if err != nil {
		autoLog.Sugar.Errorf("启动配置组失败: %v", err)
		return err
	}
	autoLog.Sugar.Infof("启动配置组成功: %v", names)
	return nil
}

// StartOneDragon 启动一条龙
// StartOneDragon 启动一条龙任务（异步）
func StartOneDragon(name string) {
	autoLog.Sugar.Infof("准备启动一条龙：%s", name)

	// 关闭软件
	control.CloseSoftware()

	// 延迟，确保软件关闭完成
	delay := 3 * time.Second
	autoLog.Sugar.Infof("等待 %v 后启动...", delay)
	time.Sleep(delay)

	// 构建执行路径
	betterGIPath := filepath.Join(config.Cfg.BetterGIAddress, "BetterGI.exe")

	// 检查文件是否存在
	if _, err := os.Stat(betterGIPath); err != nil {
		autoLog.Sugar.Errorf("BetterGI.exe 不存在: %v", err)
		return
	}

	// 构建命令
	cmd := exec.Command(betterGIPath, "--startOneDragon", name)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} // 可选：隐藏窗口
	cmd.Stdout = nil
	cmd.Stderr = nil
	err := cmd.Start()
	if err != nil {
		autoLog.Sugar.Errorf("启动一条龙失败: %v", err)
		return
	}
	autoLog.Sugar.Infof("启动一条龙成功: %s", name)
}

// 每隔1个小时发送截图
func SendWeChatImageTask() {

	cronTab := cron.New(cron.WithSeconds())

	// 定时任务,cron表达式
	//每1个小时执行一次
	spec := fmt.Sprintf("0 0 * * * *")

	// 定义定时器调用的任务函数
	task := func() {

		autoLog.Sugar.Infof("图片发送 %v", time.Now().Format("2006-01-02 15:04:05"))

		err := control.ScreenShot("jt.jpg")
		if err != nil {
			autoLog.Sugar.Error("图片发送失败:", err)
			return
		}

		Notice.SentImage("jt.jpg")

	}

	// 添加定时任务
	cronTab.AddFunc(spec, task)
	// 启动定时器
	cronTab.Start()
	// 阻塞主线程停止
	select {}

}

// 米游社签到
func MiYouSheSign() {
	//// 解析命令行参数
	//var configPath string
	//flag.StringVar(&configPath, "mysCfg", "mysConfig.yaml", "配置文件路径")
	//flag.Parse()
	//
	//// 初始化随机数种子
	utils.InitRandom()
	//
	//// 加载配置文件
	//autoLog.Sugar.Infof("米游社-正在加载配置文件: %s", configPath)
	//if err := mysConfig.LoadConfig(configPath); err != nil {
	//
	//	autoLog.Sugar.Errorf("米游社-加载配置文件失败: %v", err)
	//	return
	//}

	// 检查Cookie是否配置
	if mysConfig.GlobalConfig.Account.Cookie == "" {
		autoLog.Sugar.Errorf("米游社-Cookie未配置，请先在配置文件中设置Cookie")
		return
	}

	// 生成设备ID（如果未配置）
	if mysConfig.GlobalConfig.Device.ID == "" {
		deviceID := utils.GetDeviceID(mysConfig.GlobalConfig.Account.Cookie)
		mysConfig.GlobalConfig.Device.ID = deviceID
		//autoLog.Sugar.Infof("米游社-自动生成设备ID: %s", deviceID)
	}

	//autoLog.Sugar.Infof("米游社-签到工具启动")

	// 运行米游社签到
	if mysConfig.GlobalConfig.Mihoyobbs.Enable {
		autoLog.Sugar.Infof("米游社-开始签到任务")
		mihoyobbsClient := mihoyobbs.NewMihoyobbs()
		if err := mihoyobbsClient.Run(); err != nil {

			autoLog.Sugar.Errorf("米游社-签到失败: %v", err)
		}
	}

	// 运行游戏签到
	if mysConfig.GlobalConfig.Games.CN.Enable {

		autoLog.Sugar.Infof("米游社-开始游戏签到任务")
		if err := gamecheckin.RunAllGames(); err != nil {

			autoLog.Sugar.Errorf("米游社-游戏签到失败: %v", err)
		}
	}

	autoLog.Sugar.Infof("米游社-所有任务完成")
}
