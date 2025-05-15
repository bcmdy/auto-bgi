package task

import (
	"auto-bgi/bgiStatus"
	"auto-bgi/config"
	"auto-bgi/control"
	"encoding/json"
	"fmt"
	"github.com/go-vgo/robotgo"
	"github.com/iancoleman/orderedmap"
	"github.com/robfig/cron/v3"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var Config = config.Cfg

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

// 修改TaskEnabledList
func ChangeTaskEnabledList() error {

	now := time.Now()
	weekdayNum := int(now.Weekday())

	fmt.Printf("今天是: 星期%d", weekdayNum)
	fmt.Println("====")

	//自定义配置路径
	filename := Config.BetterGIAddress + "\\User\\OneDragon\\" + Config.ConfigName + ".json"

	// 1. 读取 JSON 文件
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("读取文件失败: %v", err)
		return err
	}

	//2. 解析为 orderedData

	jsonData := orderedmap.New()
	if err := json.Unmarshal(data, &jsonData); err != nil {
		log.Fatalf("解析 JSON 失败: %v", err)
		return err
	}

	TaskEnabled, b := jsonData.Get("TaskEnabledList")
	if !b {
		return fmt.Errorf("TaskEnabledList 字段不存在")
	}

	aa := TaskEnabled.(orderedmap.OrderedMap)
	re := regexp.MustCompile(`\d+`) // 匹配一个或多个连续数字
	var builder strings.Builder

	builder.WriteString("今日执行配置组：")
	builder.WriteString("\n")

	for _, s := range aa.Keys() {
		//fmt.Println(s)
		//fmt.Println(aa.Get(s))

		numbers := re.FindAllString(s, -1)
		if numbers == nil {
			get, _ := aa.Get(s)

			builder.WriteString(fmt.Sprintf("%s：%v", s, get))
			builder.WriteString("\n")
			continue
		}
		fmt.Println("匹配的数字:", numbers)
		if contains(numbers, weekdayNum) {
			fmt.Println("配置组:[" + s + "]已到执行时间")
			aa.Set(s, true)
			builder.WriteString(fmt.Sprintf("%s：%v", s, true))
			builder.WriteString("\n")
			continue
		} else {
			fmt.Println("配置组:[" + s + "]还未到执行时间")
			aa.Set(s, false)
			builder.WriteString(fmt.Sprintf("%s：%v", s, false))
			builder.WriteString("\n")
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
		//log.Fatalf("写入文件失败: %v", err)
		return fmt.Errorf("自定义配置写入文件失败")

	}

	bgiStatus.SendWeChatNotification(builder.String())

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
		fmt.Println("修改配置失败")
	}

	time.Sleep(4000 * time.Millisecond)
	fmt.Println("修改配置成功")

	control.OpenSoftware(Config.BetterGIAddress + "/BetterGI.exe")
	// 等待一小会儿
	time.Sleep(3000 * time.Millisecond)

	//fmt.Println("启动原神")
	//control.MouseClick(1267, 536, "left", false)
	//
	//fmt.Println("等待一小会儿（等待原神启动）")
	//time.Sleep(60 * time.Second)

	fmt.Println("切换屏幕")
	control.SwitchingScreens("更好的原神")

	time.Sleep(1000 * time.Millisecond)

	windows := control.GetWindows()
	if windows != "更好的原神" {
		control.SwitchingScreens("更好的原神")
	}

	fmt.Println("点击一条龙")
	control.MouseClick(Config.LongX, Config.LongY, "left", false)

	time.Sleep(3000 * time.Millisecond)

	fmt.Println("执行")
	control.MouseClick(Config.ExecuteX, Config.ExecuteY, "left", false)
}

func OneLong() {

	cronTab := cron.New(cron.WithSeconds())

	// 定时任务,cron表达式
	spec := fmt.Sprintf("0 %d %d * * *", Config.OneLongMinute, Config.OneLongHour)

	// 定义定时器调用的任务函数
	task := func() {
		fmt.Print("一条龙服务启动", time.Now().Format("2006-01-02 15:04:05"))

		OneLongTask()

		time.Sleep(1000 * time.Millisecond)

		schedule, err := config.Parser.Parse(spec)
		if err != nil {
			fmt.Println("解析失败:", err)
			return
		}

		fmt.Print("一条龙服务启动完毕", "下次执行时间:", schedule.Next(time.Now()).Format("2006-01-02 15:04:05"))
	}

	// 添加定时任务
	cronTab.AddFunc(spec, task)
	// 启动定时器
	cronTab.Start()
	// 阻塞主线程停止
	select {}

}

func ScriptGroupTask() {
	fmt.Print("调度器服务启动", time.Now().Format("2006-01-02 15:04:05"))

	time.Sleep(2 * time.Second)

	robotgo.KeyDown("alt") // 按下 Alt 键
	robotgo.KeyTap("tab")  // 按一下 Tab 键
	robotgo.KeyUp("alt")   // 释放 Alt 键

	time.Sleep(1000 * time.Millisecond)

	control.SwitchingScreens("更好的原神")
	time.Sleep(1000 * time.Millisecond)

	windows := control.GetWindows()
	if windows != "更好的原神" {
		control.SwitchingScreens("更好的原神")
	}
	time.Sleep(1000 * time.Millisecond)

	//点击全自动
	control.MouseClick(582, 495, "left", false)

	time.Sleep(1000 * time.Millisecond)

	//点击调度器
	control.MouseClick(606, 538, "left", false)

	time.Sleep(1000 * time.Millisecond)

	//点击连续执行
	control.MouseClick(772, 791, "left", false)

	time.Sleep(1000 * time.Millisecond)

	//点击全选
	control.MouseClick(907, 379, "left", false)

	time.Sleep(1000 * time.Millisecond)

	//点击确认执行
	control.MouseClick(919, 704, "left", false)

	time.Sleep(1000 * time.Millisecond)
}

func MysSignIn() {
	cronTab := cron.New(cron.WithSeconds())

	// 定时任务,cron表达式
	spec := fmt.Sprintf("0 %d %d * * *", 10, 2)

	// 定义定时器调用的任务函数
	task := func() {
		fmt.Print("米游社签到服务启动", time.Now().Format("2006-01-02 15:04:05"))

		err := control.HttpGet("http://localhost:8888/qd")
		if err != nil {
			fmt.Println("签到失败", err)
		}

		time.Sleep(1000 * time.Millisecond)

		schedule, err := config.Parser.Parse(spec)
		if err != nil {
			fmt.Println("解析失败:", err)
			return
		}

		fmt.Print("米游社签到服务启动完毕", "下次执行时间:", schedule.Next(time.Now()).Format("2006-01-02 15:04:05"))
	}

	// 添加定时任务
	cronTab.AddFunc(spec, task)
	// 启动定时器
	cronTab.Start()
	// 阻塞主线程停止
	select {}
}
