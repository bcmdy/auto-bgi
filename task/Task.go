package task

import (
	"auto-bgi/autoLog"
	"auto-bgi/bgiStatus"
	"auto-bgi/config"
	"auto-bgi/control"
	"auto-bgi/internal/gamecheckin"
	"auto-bgi/internal/mihoyobbs"
	"auto-bgi/internal/mysConfig"
	"auto-bgi/internal/utils"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/iancoleman/orderedmap"
	"github.com/robfig/cron/v3"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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

// 修改TaskEnabledList
func ChangeTaskEnabledList() error {

	now := time.Now()
	weekdayNum := int(now.Weekday())

	autoLog.Sugar.Infof("今天是: 星期%d", weekdayNum)

	OneLongName := config.GetTodayOneLongName()

	//自定义配置路径
	filename := config.Cfg.BetterGIAddress + "\\User\\OneDragon\\" + OneLongName + ".json"

	// 1. 读取 JSON 文件
	data, err := os.ReadFile(filename)
	if err != nil {
		autoLog.Sugar.Errorf("一条龙读取文件失败%s: %v", OneLongName, err)
		return err
	}

	//2. 解析为 orderedData
	jsonData := orderedmap.New()
	if err := json.Unmarshal(data, &jsonData); err != nil {

		autoLog.Sugar.Errorf("解析 JSON 失败: %v", err)
		return err
	}
	_, b2 := jsonData.Get("SelectedPeriodList")
	if !b2 {
		autoLog.Sugar.Errorf("SelectedPeriodList 字段不存在")
	} else {
		autoLog.Sugar.Infof("SelectedPeriodList 字段存在")
		ReadChaBaoBgiConfig(filename)
		return nil
	}

	TaskEnabled, b := jsonData.Get("TaskEnabledList")
	if !b {
		autoLog.Sugar.Errorf("TaskEnabledList 字段不存在")
		return fmt.Errorf("TaskEnabledList 字段不存在")
	}

	aa := TaskEnabled.(orderedmap.OrderedMap)
	re := regexp.MustCompile(`\d+`) // 匹配一个或多个连续数字
	var builder strings.Builder

	builder.WriteString("今日执行一条龙：" + OneLongName + "\n")
	builder.WriteString("今日执行配置组：")
	builder.WriteString("\n")

	var oneLongLog strings.Builder

	for _, s := range aa.Keys() {

		autoLog.Sugar.Infof("配置组:%s", s)
		numbers := re.FindAllString(s, -1)
		if numbers == nil {
			get, _ := aa.Get(s)

			if get == true {
				builder.WriteString(fmt.Sprintf("%s：%s", s, "执行"))
				builder.WriteString("\n")

				oneLongLog.WriteString(fmt.Sprintf("%s：%s", s, "执行"))
				oneLongLog.WriteString("\n")

				continue
			}
			continue
		}
		autoLog.Sugar.Infof("匹配的数字:%v", numbers)
		if contains(numbers, weekdayNum) {
			autoLog.Sugar.Infof("配置组:[" + s + "]已到执行时间")
			aa.Set(s, true)
			//builder.WriteString(fmt.Sprintf("%s：%v", s, true))
			builder.WriteString(fmt.Sprintf("%s：%s", s, "执行"))
			builder.WriteString("\n")

			oneLongLog.WriteString(fmt.Sprintf("%s：%s", s, "执行"))
			oneLongLog.WriteString("\n")
			continue
		} else {
			autoLog.Sugar.Infof("配置组:[" + s + "]还未到执行时间")
			aa.Set(s, false)
			//builder.WriteString(fmt.Sprintf("%s：%v", s, false))
			//builder.WriteString("\n")
			continue
		}
	}

	//fmt.Println("修改后的 jsonData:", jsonData)
	//// 5. 重新编码 JSON（保持缩进）
	updatedData, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON 编码失败")
	}

	// 6. 写回文件
	if err := os.WriteFile(filename, updatedData, 0644); err != nil {

		autoLog.Sugar.Errorf("写入文件失败: %v", err)
		return fmt.Errorf("自定义配置写入文件失败")

	}

	//将执行配置写入文件，直接覆盖
	// 定义要写入的内容
	content := []byte(oneLongLog.String())
	// 打开文件，如果文件不存在则创建
	file, err := os.OpenFile("OneLongTask.txt", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()
	file.Write(content)

	//发送通知
	bgiStatus.SentText(builder.String())

	return nil

}

func OneLongTask() {
	autoLog.Sugar.Info("开始执行一条龙任务")

	//// 1. 并行执行日志监控
	//go func() {
	//	autoLog.Sugar.Info("启动日志监控")
	//	bgiStatus.LogM()
	//}()

	// 2. 并行执行用户目录备份
	go func() {
		autoLog.Sugar.Info("开始备份 User 目录")
		BackupUsers()
	}()

	// 3. 关闭软件（同步，后续任务依赖此步骤）
	control.CloseSoftware()
	autoLog.Sugar.Info("软件已关闭")

	// 4. 批量更新脚本
	autoLog.Sugar.Info("开始批量更新脚本")
	if err := bgiStatus.BatchUpdateScript(); err != "" {
		autoLog.Sugar.Errorf("批量更新脚本失败: %v", err)
		return
	}

	// 5. 修改配置
	if err := ChangeTaskEnabledList(); err != nil {
		autoLog.Sugar.Errorf("修改配置失败: %v", err)
		return
	}
	autoLog.Sugar.Info("修改配置成功")

	// 6. 启动今日一条龙
	longName := config.GetTodayOneLongName()
	autoLog.Sugar.Infof("今日启动一条龙: %s", longName)

	StartOneDragon(longName)

	autoLog.Sugar.Info("一条龙任务执行完成")
}

func OneLong() {

	cronTab := cron.New(cron.WithSeconds())

	// 定时任务,cron表达式
	spec := fmt.Sprintf("0 %d %d * * *", config.Cfg.OneLong.OneLongMinute, config.Cfg.OneLong.OneLongHour)

	// 定义定时器调用的任务函数
	task := func() {

		autoLog.Sugar.Infof("一条龙服务启动 %v", time.Now().Format("2006-01-02 15:04:05"))

		OneLongTask()

		time.Sleep(1000 * time.Millisecond)

		schedule, err := config.Parser.Parse(spec)
		if err != nil {
			autoLog.Sugar.Error("解析失败:", err)
			return
		}

		autoLog.Sugar.Infof("一条龙服务启动完毕 %v", schedule.Next(time.Now()).Format("2006-01-02 15:04:05"))
	}

	// 添加定时任务
	cronTab.AddFunc(spec, task)
	// 启动定时器
	cronTab.Start()
	// 阻塞主线程停止
	select {}

}

func MysSignIn() {
	cronTab := cron.New(cron.WithSeconds())

	// 定时任务,cron表达式
	spec := fmt.Sprintf("0 %d %d * * *", 20, 0)

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

// 启动一条龙
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

// 定时更新代码
func UpdateCode() {
	cronTab := cron.New(cron.WithSeconds())

	// 定时任务,cron表达式
	//每1个小时执行一次
	spec := fmt.Sprintf("0 0 */2 * * *")
	//spec := fmt.Sprintf("0 %d %d * * *", Config.OneLongMinute, Config.OneLongHour)

	// 定义定时器调用的任务函数
	task := func() {
		autoLog.Sugar.Infof("仓库更新 %v", time.Now().Format("2006-01-02 15:04:05"))

		err := bgiStatus.GitPull()
		if err != nil {
			autoLog.Sugar.Error("更新失败:", err)
		}

		autoLog.Sugar.Infof("仓库更新启动完毕")
	}

	// 添加定时任务
	cronTab.AddFunc(spec, task)
	// 启动定时器
	cronTab.Start()
	// 阻塞主线程停止
	select {}
}

const interval = 72 * time.Hour

// 每周一备份users文件夹
func BackupUsers() {

	var lastBackupStr string
	err := config.DB.QueryRow(`SELECT autobgi_value FROM autoBgi_config WHERE autobgi_key = 'BackupUserTime'`).Scan(&lastBackupStr)
	if err != nil && err != sql.ErrNoRows {
		autoLog.Sugar.Errorf("查询 BackupUserTime 失败: %v", err)
		return
	}
	// 解析上次时间
	var lastBackup time.Time
	if lastBackupStr != "" {
		parsed, per := time.ParseInLocation("2006-01-02 15:04:05", lastBackupStr, time.Local)
		if per == nil {
			lastBackup = parsed
		} else {
			autoLog.Sugar.Warnf("时间解析失败(%v)，使用默认时间", per)
			lastBackup = time.Now().Add(-interval)
		}
	}

	now := time.Now()

	if now.Sub(lastBackup) >= interval {
		autoLog.Sugar.Info("🟢 满足条件，开始备份 users 文件夹...")
		autoLog.Sugar.Infof("开始备份user文件夹")
		err4 := bgiStatus.ZipDir(config.Cfg.BetterGIAddress+"\\User\\", "Users\\User"+time.Now().Format("2006100215020405")+".zip", true)
		if err4 != nil {
			autoLog.Sugar.Errorf("备份失败: %v")
			return
		}

		autoLog.Sugar.Info("备份成功")

		// 更新数据库记录
		_, err = config.DB.Exec(`UPDATE autoBgi_config SET autobgi_value = ? WHERE autobgi_key = 'BackupUserTime'`, time.Now().Format("2006-01-02 15:04:05"))
		if err != nil {
			autoLog.Sugar.Errorf("更新 BackupUserTime 失败: %v", err)
		} else {
			autoLog.Sugar.Info("✅ 备份完成，时间已更新")
		}
	} else {
		autoLog.Sugar.Infof("⏳ 未满足条件（上次：%v，下次至少需等待：%.0f小时）", lastBackup, (interval - now.Sub(lastBackup)).Hours())
	}
}

// 每隔1个小时发送截图
func SendWeChatImageTask() {

	cronTab := cron.New(cron.WithSeconds())

	// 定时任务,cron表达式
	//每1个小时执行一次
	spec := fmt.Sprintf("0 */59 * * * *")

	// 定义定时器调用的任务函数
	task := func() {

		autoLog.Sugar.Infof("图片发送 %v", time.Now().Format("2006-01-02 15:04:05"))

		err := control.ScreenShot()
		if err != nil {
			autoLog.Sugar.Error("图片发送失败:", err)
			return
		}

		bgiStatus.SentImage("jt.png")

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
	// 解析命令行参数
	var configPath string
	flag.StringVar(&configPath, "mysConfig", "mysConfig.yaml", "配置文件路径")
	flag.Parse()

	// 初始化随机数种子
	utils.InitRandom()

	// 加载配置文件
	autoLog.Sugar.Infof("米游社-正在加载配置文件: %s", configPath)
	if err := mysConfig.LoadConfig(configPath); err != nil {

		autoLog.Sugar.Errorf("米游社-加载配置文件失败: %v", err)
		os.Exit(1)
	}

	// 检查Cookie是否配置
	if mysConfig.GlobalConfig.Account.Cookie == "" {
		autoLog.Sugar.Errorf("米游社-Cookie未配置，请先在配置文件中设置Cookie")
		os.Exit(1)
	}

	// 生成设备ID（如果未配置）
	if mysConfig.GlobalConfig.Device.ID == "" {
		deviceID := utils.GetDeviceID(mysConfig.GlobalConfig.Account.Cookie)
		mysConfig.GlobalConfig.Device.ID = deviceID
		autoLog.Sugar.Infof("米游社-自动生成设备ID: %s", deviceID)
	}

	autoLog.Sugar.Infof("米游社-签到工具启动")

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
