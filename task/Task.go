package task

import (
	"auto-bgi/autoLog"
	"auto-bgi/bgiStatus"
	"auto-bgi/config"
	"auto-bgi/control"
	"encoding/json"
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
			configFile, err2 := os.Open(filePath) // 假设 JSON 文件名为 config.json
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
			pathingConfig, ok := result["config"].(map[string]interface{})["pathingConfig"].(map[string]interface{})
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
	get, b2 := jsonData.Get("SelectedPeriodList")
	fmt.Println(get)
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

	bgiStatus.SendWeChatNotification(builder.String())

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

	return nil

}

func OneLongTask() {

	//关闭软件
	control.CloseSoftware()

	// 等待一小会儿
	time.Sleep(3000 * time.Millisecond)

	//修改我的配置
	err := ChangeTaskEnabledList()
	if err != nil {
		autoLog.Sugar.Errorf("修改配置失败: %v", err)
	}

	time.Sleep(4000 * time.Millisecond)
	autoLog.Sugar.Info("修改配置成功")

	longName := config.GetTodayOneLongName()

	autoLog.Sugar.Infof("今日启动一条龙: %s", longName)

	////开启录屏视频
	go control.StartRecord()

	StartOneDragon(longName)

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

		//config.GenShinSign()

		err := control.HttpGet(config.Cfg.MySign.Url + "/qd")
		if err != nil {

			autoLog.Sugar.Error("签到失败:", err)
			return
		}

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

// 启动配置组
func StartGroups(name string) {

	control.CloseSoftware()
	time.Sleep(5 * time.Second)

	betterGIPath := filepath.Join(config.Cfg.BetterGIAddress, "BetterGI.exe")
	cmd := exec.Command(betterGIPath, "--startGroups", name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		autoLog.Sugar.Errorf("启动配置组失败: %v", err)
		return
	}

	autoLog.Sugar.Infof("%s 启动配置组成功", name)

}

// 启动一条龙
func StartOneDragon(name string) {
	control.CloseSoftware()
	time.Sleep(5 * time.Second)

	betterGIPath := filepath.Join(config.Cfg.BetterGIAddress, "BetterGI.exe")
	cmd := exec.Command(betterGIPath, "--startOneDragon", name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		autoLog.Sugar.Errorf("启动一条龙失败: %v", err)
		return
	}
	autoLog.Sugar.Infof("%s 启动一条龙成功", name)
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

// 每周一备份users文件夹
func BackupUsers() {

}

// 每隔半个小时发送截图
func SendWeChatImageTask() {

	cronTab := cron.New(cron.WithSeconds())

	// 定时任务,cron表达式
	//每1个小时执行一次
	spec := fmt.Sprintf("0 0 */1 * * *")

	// 定义定时器调用的任务函数
	task := func() {

		running := bgiStatus.IsWechatRunning()
		if !running {
			return
		}

		autoLog.Sugar.Infof("图片发送 %v", time.Now().Format("2006-01-02 15:04:05"))

		err := control.ScreenShot()
		if err != nil {
			autoLog.Sugar.Error("图片发送失败:", err)
			return
		}
		time.Sleep(2000 * time.Millisecond)
		err2 := bgiStatus.SendWeChatImage("jt.png")
		if err2 != nil {
			autoLog.Sugar.Error("图片发送失败:", err2)
			return
		}
		autoLog.Sugar.Infof("图片发送成功")

	}

	// 添加定时任务
	cronTab.AddFunc(spec, task)
	// 启动定时器
	cronTab.Start()
	// 阻塞主线程停止
	select {}

}
