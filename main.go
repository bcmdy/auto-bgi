package main

import (
	"auto-bgi/bgiStatus"
	"auto-bgi/config"
	_ "auto-bgi/config"
	"auto-bgi/control"
	"auto-bgi/task"
	"bufio"
	"embed"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"syscall"
	"time"
)

var (
	user32               = syscall.NewLazyDLL("user32.dll")
	procFindWindow       = user32.NewProc("FindWindowW")
	procSetForegroundWnd = user32.NewProc("SetForegroundWindow")
)

func findLastJSONLine(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var lastJSONLine string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, ".json") {
			lastJSONLine = line
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	if lastJSONLine == "" {
		return "", fmt.Errorf("no line containing '.json' found")
	}

	return lastJSONLine, nil
}

func findLastGroup(filename string) (string, error) {

	pattern := `配置组 "(.*?)" 加载完成，共\d+个脚本，开始执行`

	re := regexp.MustCompile(pattern)

	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// 用于存储最后匹配的行和配置组名称
	var lastMatch string
	var lastGroup string

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		matches := re.FindStringSubmatch(line)
		if matches != nil {
			lastMatch = line
			lastGroup = matches[1] // 第一个捕获组是配置组名称
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}
	// 输出结果
	if lastMatch != "" {
		fmt.Println("最后匹配的行:", lastMatch)
		fmt.Println("配置组名称:", lastGroup)
	} else {
		errs := fmt.Errorf("没有找到匹配的行", 500)
		return "", errs
	}
	return lastGroup, nil
}

// 修改json文件
func modifyJSONFile(filename string, targetProjectName string, newNextFlag bool) error {
	// 1. 读取 JSON 文件
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("读取文件失败: %v", err)
		return err
	}
	// 2. 解析为 map[string]interface{}（保持原始结构）
	var jsonData map[string]interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		log.Fatalf("解析 JSON 失败: %v", err)
		return err
	}
	// 3. 获取 projects 数组
	projects, ok := jsonData["projects"].([]interface{})
	if !ok {
		log.Fatal("projects 字段不是数组或不存在")
		return err
	}
	// 4. 遍历查找目标 Project
	found := false
	for _, p := range projects {
		project, ok := p.(map[string]interface{})
		if !ok {
			continue
		}
		if name, ok := project["name"].(string); ok && name == targetProjectName {
			project["nextFlag"] = newNextFlag // 只修改 NextFlag
			found = true
			break
		}
	}
	if !found {
		log.Fatalf("未找到项目: %s", targetProjectName)
		return err
	}

	// 5. 重新编码 JSON（保持缩进）
	updatedData, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		log.Fatalf("JSON 编码失败: %v", err)
		return err
	}

	// 6. 写回文件
	if err := os.WriteFile(filename, updatedData, 0644); err != nil {
		log.Fatalf("写入文件失败: %v", err)
		return err
	}

	fmt.Printf("已更新 %s 的 NextFlag → %v\n", targetProjectName, newNextFlag)
	return err

}

var Config = config.Cfg

func start() error {

	// 生成日志文件名
	date := time.Now().Format("20060102")

	filename := filepath.Clean(fmt.Sprintf("%s\\log\\better-genshin-impact%s.log", Config.BetterGIAddress, date))

	line, err := findLastJSONLine(filename)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return err
	}
	fmt.Println("Last line containing '.json':")
	fmt.Println(line)

	start := strings.Index(line, `"`)
	end := strings.LastIndex(line, `"`)
	//
	group, err := findLastGroup(filename)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return err
	}

	if start != -1 && end != -1 && start < end {
		content := line[start+1 : end]
		fmt.Println(content)
		err := modifyJSONFile(fmt.Sprintf("%s\\User\\ScriptGroup\\%s", Config.BetterGIAddress, group+".json"), content, true)
		if err != nil {
			return err
		}
	}

	control.OpenSoftware(fmt.Sprintf("%s\\BetterGI.exe", Config.BetterGIAddress))

	// 等待一小会儿
	time.Sleep(1000 * time.Millisecond)

	//fmt.Println("切换屏幕")
	control.SwitchingScreens("更好的原神")

	// 等待一小会儿
	time.Sleep(1000 * time.Millisecond)

	//点击全自动
	control.MouseClick(582, 495, "left", false)

	// 等待一小会儿
	time.Sleep(1000 * time.Millisecond)

	//点击调度器
	control.MouseClick(606, 538, "left", false)

	time.Sleep(1000 * time.Millisecond)

	//获取当前配置组index
	num, err := bgiStatus.GetGroupNum(fmt.Sprintf("%s\\User\\ScriptGroup\\%s", Config.BetterGIAddress, group+".json"))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return err
	}
	fmt.Println("index", num)
	i := (num - 1) * 38

	fmt.Println("坐标", 325+i)
	//点击锄地
	control.MouseClick(722, 325+i, "left", false)

	time.Sleep(1000 * time.Millisecond)

	//点击运行
	control.MouseClick(908, 365, "left", false)

	time.Sleep(1000 * time.Millisecond)

	bgiStatus.SendWeChatNotification(fmt.Sprintf("BIG启动成功,当前配置组配置组%s,脚本:%s", group, line))

	return nil
}

// 获取材料名称的前缀
func getPrefix(name string) string {
	parts := strings.Split(name, "")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

//go:embed html/*
var htmlFS embed.FS

func main() {

	//创建一个服务
	ginServer := gin.Default()

	////加载templates中所有模板文件, 使用不同目录下名称相同的模板,注意:一定要放在配置路由之前才得行
	//ginServer.LoadHTMLGlob("html/*")

	// 引入html
	ginServer.SetHTMLTemplate(template.Must(template.New("").ParseFS(htmlFS, "html/*.html")))

	// 提供静态资源服务，把 html 目录映射为 /static 路径
	ginServer.Static("/static", ".")

	//重启接口
	ginServer.GET("/test", func(c *gin.Context) {

		err := start()
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"message": "错误"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Success"})

	})

	//日志查询
	ginServer.GET("/", func(c *gin.Context) {
		// 生成日志文件名
		date := time.Now().Format("20060102")

		filename := filepath.Clean(fmt.Sprintf("%s\\log\\better-genshin-impact%s.log", Config.BetterGIAddress, date))

		line, err := findLastJSONLine(filename)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
		fmt.Println("Last line containing '.json':")
		fmt.Println(line)

		group, err := findLastGroup(filename)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
		jsonStr := fmt.Sprintf("%s\\User\\ScriptGroup\\%s", Config.BetterGIAddress, group+".json")
		progress, err := bgiStatus.Progress(jsonStr, line)
		if err != nil {
			fmt.Printf("%v\n", err)
			progress = "0/0"
		}

		running := bgiStatus.IsWechatRunning()

		jsProgress, err := bgiStatus.JsProgress(filename, "当前进度：(.*?)")
		if err != nil {
			jsProgress = "无"
		}

		c.HTML(http.StatusOK, "index.html", map[string]interface{}{
			"group":      group,
			"line":       line,
			"progress":   progress,
			"running":    running,
			"jsProgress": jsProgress,
		})

	})

	//日志查询
	ginServer.GET("/mark", func(c *gin.Context) {
		// 生成日志文件名
		date := time.Now().Format("20060102")

		filename := filepath.Clean(fmt.Sprintf("%s\\log\\better-genshin-impact%s.log", Config.BetterGIAddress, date))

		line, err := findLastJSONLine(filename)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
		fmt.Println("Last line containing '.json':")
		fmt.Println(line)

		group, err := findLastGroup(filename)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
		jsonStr := fmt.Sprintf("%s\\User\\ScriptGroup\\%s", Config.BetterGIAddress, group+".json")
		progress, err := bgiStatus.Progress(jsonStr, line)
		if err != nil {
			fmt.Printf("%v\n", err)
			progress = "0/0"
		}

		running := bgiStatus.IsWechatRunning()

		jsProgress, err := bgiStatus.JsProgress(filename, "当前进度：(.*?)")
		if err != nil {
			jsProgress = "无"
		}

		c.JSON(http.StatusOK, map[string]interface{}{
			"group":      group,
			"line":       line,
			"progress":   progress,
			"running":    running,
			"jsProgress": jsProgress,
		})

	})

	ginServer.POST("/oneLong", func(context *gin.Context) {

		task.OneLongTask()

		context.JSON(http.StatusOK, gin.H{"status": "received", "data": "一条龙启动成功"})
	})

	ginServer.POST("/closeBgi", func(context *gin.Context) {

		control.CloseSoftware()

		context.JSON(http.StatusOK, gin.H{"status": "received", "data": "BGI关闭成功"})
	})

	ginServer.POST("/closeYuanShen", func(context *gin.Context) {

		control.CloseYuanShen()

		context.JSON(http.StatusOK, gin.H{"status": "received", "data": "原神关闭成功"})
	})

	ginServer.GET("/TodayHarvest", func(context *gin.Context) {

		// 获取统计结果
		stats, err := bgiStatus.TodayHarvest()
		if err != nil {
			context.HTML(http.StatusInternalServerError, "error.html", gin.H{
				"error": err.Error(),
			})
			return
		}

		// 转换为前端更易处理的格式
		var items []struct {
			Name  string `json:"name"`
			Count int    `json:"count"`
		}

		for name, count := range stats {
			items = append(items, struct {
				Name  string `json:"name"`
				Count int    `json:"count"`
			}{
				Name:  name,
				Count: count,
			})
		}

		// 按数量排序
		sort.Slice(items, func(i, j int) bool {
			return items[i].Count > items[j].Count
		})

		// 传递给模板
		context.HTML(http.StatusOK, "harvest.html", gin.H{
			"title": "今日收获统计",
			"items": items,
		})
	})

	//发送截图
	ginServer.POST("/getImage", func(c *gin.Context) {

		err := control.ScreenShot()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "received", "data": "截图失败"})
			return
		} else {
			err := bgiStatus.SendWeChatImage("jt.png")
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"status": "received", "data": "截图失败"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"status": "received", "data": "发送成功"})
			return
		}

	})

	//webhook接口
	ginServer.POST("/webhook", func(c *gin.Context) {
		var j map[string]interface{}

		// 绑定JSON数据到map
		if err := c.ShouldBindJSON(&j); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		fmt.Println(j["message"])

		c.JSON(http.StatusOK, gin.H{"status": "received", "data": j})
	})

	//米游社签到
	ginServer.POST("/MysSignIn", func(context *gin.Context) {

		err := control.HttpGet("http://localhost:8888/qd")
		if err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"status": "received", "data": err})
			return
		}
		context.JSON(http.StatusOK, gin.H{"status": "received", "data": "米游社签到成功"})
		return
	})

	//背包统计
	ginServer.GET("/BagStatistics", func(context *gin.Context) {
		statistics, err := bgiStatus.BagStatistics()

		//// 按 Cl 字段排序
		//sort.Slice(statistics, func(i, j int) bool {
		//	return statistics[i].Cl < statistics[j].Cl
		//})

		// 按材料名称排序，再按日期排序
		sort.Slice(statistics, func(i, j int) bool {
			// 首先按材料名称排序
			if statistics[i].Cl != statistics[j].Cl {
				return statistics[i].Cl < statistics[j].Cl
			}
			// 如果材料名称相同，则按日期排序
			layout := "2006/1/2 15:04:05"
			ti, _ := time.Parse(layout, statistics[i].Data)
			tj, _ := time.Parse(layout, statistics[j].Data)
			return ti.Before(tj)
		})

		if err != nil {
			// 传递给模板
			context.HTML(http.StatusOK, "bg.html", gin.H{
				"title": "背包统计",
				"items": nil,
			})
			return
		}

		// 传递给模板
		context.HTML(http.StatusOK, "bg.html", gin.H{
			"title": "背包统计",
			"items": statistics,
		})
	})

	//删除背包统计记录
	ginServer.POST("/deleteBag", func(context *gin.Context) {
		isOk := bgiStatus.DeleteBagStatistics()

		data := gin.H{
			"message": isOk,
		}

		context.JSON(http.StatusOK, data)
	})

	//删除背包统计记录
	ginServer.GET("/abc", func(context *gin.Context) {
		statistics, _ := bgiStatus.MorasStatistics()

		data := gin.H{
			"message": statistics,
		}

		context.JSON(http.StatusOK, data)
	})

	//查询所有配置组
	ginServer.GET("/listGroups", func(context *gin.Context) {
		groups, err := task.ListGroups()
		if err != nil {
			return
		}
		fmt.Println(groups)

		// 传递给模板
		context.HTML(http.StatusOK, "listGroups.html", gin.H{
			"title": "调度器",
			"items": groups,
		})
	})

	//启动配置组
	ginServer.POST("/startGroups", func(context *gin.Context) {

		var data map[string]string
		err := context.BindJSON(&data)
		if err != nil {
			fmt.Println("err:", err)
			return
		}
		fmt.Println(data)
		task.StartGroups(data["name"])
		if err != nil {
			return
		}
		context.JSON(http.StatusOK, gin.H{"message": "Success"})
	})

	//一条龙
	if Config.IsStartTimeLong {
		go task.OneLong()
		fmt.Println("一条龙开启状态")
	} else {
		fmt.Println("一条龙关闭状态")
	}

	//检查BGI状态
	go bgiStatus.CheckBetterGIStatus()

	if Config.IsMysSignIn {
		//米游社自动签到
		go task.MysSignIn()
		fmt.Println("米游社自动签到开启状态")
	} else {
		fmt.Println("米游社自动签到关闭状态")
	}

	//服务器端口
	post := Config.Post
	if post == "" {
		post = ":8082"
	}
	err := ginServer.Run(post)
	if err != nil {
		return
	}

}

//go build -o auto-bgi.exe main.go
