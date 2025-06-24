package main

import (
	"auto-bgi/autoLog"
	"auto-bgi/bgiStatus"
	"auto-bgi/config"
	"auto-bgi/control"
	"auto-bgi/task"
	"bufio"
	"embed"
	"encoding/json"
	"fmt"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"html/template"
	"io"
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

func init() {
	//判断目录是否设置正确
	exists, err := bgiStatus.CheckConfig()
	if !exists {
		fmt.Println(err)
		//程序暂停，任意键退出
		fmt.Println("=======程序暂停，任意键退出=========")
		fmt.Scanln()
		os.Exit(1)
	}
}

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
		//autoLog.Sugar.Infof("最后匹配的行:", lastMatch)
		//autoLog.Sugar.Infof("配置组名称:", lastGroup)
	} else {
		errs := fmt.Errorf("没有找到匹配的行", 500)
		return "", errs
	}
	return lastGroup, nil
}

var Config = config.Cfg

//go:embed html/*
var htmlFS embed.FS

func tojson(v interface{}) template.JS {
	a, _ := json.Marshal(v)
	return template.JS(a)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 如果跨域就写逻辑
	},
}

func main() {

	// 初始化日志
	autoLog.Init()
	defer autoLog.Sync()

	gin.SetMode(gin.ReleaseMode)

	//创建一个服务
	ginServer := gin.Default()

	ginServer.SetTrustedProxies(nil)
	ginServer.Use(gzip.Gzip(gzip.DefaultCompression))

	////加载templates中所有模板文件, 使用不同目录下名称相同的模板,注意:一定要放在配置路由之前才得行
	//ginServer.LoadHTMLGlob("html/*")

	tmpl := template.Must(
		template.New("").Funcs(template.FuncMap{
			"tojson": tojson,
		}).ParseFS(htmlFS, "html/*.html"),
	)

	ginServer.SetHTMLTemplate(tmpl)

	// 引入html
	//ginServer.SetHTMLTemplate(template.Must(template.New("").ParseFS(htmlFS, "html/*.html")))

	// 提供静态资源服务，把 html 目录映射为 /static 路径
	ginServer.Static("/static", "static")

	ginServer.GET("/log", func(context *gin.Context) {

		// 传递给模板
		context.HTML(http.StatusOK, "log.html", nil)
	})

	//查询今日所有日志文件
	ginServer.GET("/logFiles", func(c *gin.Context) {
		filePath := filepath.Clean(fmt.Sprintf("%s\\log", Config.BetterGIAddress)) // 本地日志路径
		files, err := bgiStatus.FindLogFiles(filePath)
		if err != nil {
			return
		}
		c.JSON(http.StatusOK, gin.H{"files": files})
	})

	// WebSocket 处理器
	ginServer.GET("/ws/:name", func(c *gin.Context) {
		logName := c.Param("name")
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		if logName == "" {
			date := time.Now().Format("20060102")
			logName = fmt.Sprintf("better-genshin-impact%s.log", date)
		}

		filePath := filepath.Join(Config.BetterGIAddress, "log", logName)
		file, err := os.Open(filePath)
		if err != nil {
			conn.WriteMessage(websocket.TextMessage, []byte("无法打开日志文件"))
			return
		}
		defer file.Close()

		// 定位到文件末尾
		file.Seek(0, io.SeekEnd)

		reader := bufio.NewReader(file)

		for {
			// 尝试读取一行
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					// 没新数据就稍等
					time.Sleep(500 * time.Millisecond)
					continue
				}
				autoLog.Sugar.Errorf("读取日志出错: %v\n", err)
				break
			}

			// 检查连接是否还活着
			err = conn.WriteMessage(websocket.TextMessage, []byte(line))
			if err != nil {
				log.Println("WebSocket 写入失败:", err)
				break
			}
		}
	})

	ginServer.GET("/", func(c *gin.Context) {
		// 传递给模板
		c.HTML(http.StatusOK, "index.html", nil)
	})

	//日志查询
	ginServer.GET("/index", func(c *gin.Context) {
		// 生成日志文件名
		date := time.Now().Format("20060102")

		filename := filepath.Clean(fmt.Sprintf("%s\\log\\better-genshin-impact%s.Log", Config.BetterGIAddress, date))

		line, err := findLastJSONLine(filename)
		if err != nil {

			autoLog.Sugar.Errorf("Error: %v\n", err)
		}

		group, err := findLastGroup(filename)
		if err != nil {

			autoLog.Sugar.Errorf("Error: %v\n", err)
		}
		jsonStr := fmt.Sprintf("%s\\User\\ScriptGroup\\%s", Config.BetterGIAddress, group+".json")
		progress, err := bgiStatus.Progress(jsonStr, line)
		if err != nil {

			autoLog.Sugar.Errorf("%v\n", err)
			progress = "0/0"
		}

		running := bgiStatus.IsWechatRunning()

		jsProgress, err := bgiStatus.JsProgress(filename, "当前进度：(.*?)")
		if err != nil {
			jsProgress = "无"
		}

		data := make(map[string]interface{})
		data["group"] = group
		data["line"] = line
		data["progress"] = progress
		data["running"] = running
		data["jsProgress"] = jsProgress

		c.JSON(http.StatusOK, data)

	})

	//日志查询
	ginServer.GET("/mark", func(c *gin.Context) {
		// 生成日志文件名
		date := time.Now().Format("20060102")

		filename := filepath.Clean(fmt.Sprintf("%s\\autoLog\\better-genshin-impact%s.autoLog", Config.BetterGIAddress, date))

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

	//一条龙
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

		autoLog.Sugar.Infof("查询所有配置组:%s", groups)

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
		task.StartGroups(data["name"])
		if err != nil {
			return
		}
		context.JSON(http.StatusOK, gin.H{"message": "Success"})
	})

	//查询狗粮日志
	ginServer.GET("/getAutoArtifactsPro", func(context *gin.Context) {
		pro, err := bgiStatus.GetAutoArtifactsPro()
		autoLog.Sugar.Infof("狗粮记录:%s", pro)

		if err != nil {
			// 传递给模板
			context.HTML(http.StatusOK, "AutoArtifactsPro.html", gin.H{
				"title": "狗粮收益查询",
				"items": nil,
			})
			return
		}
		context.HTML(http.StatusOK, "AutoArtifactsPro.html", gin.H{
			"title": "狗粮收益查询",
			"items": pro,
		})

	})

	//查询狗粮日志
	ginServer.GET("/getAutoArtifactsPro2", func(context *gin.Context) {

		fileName := context.Query("fileName")
		if fileName == "" {
			context.HTML(http.StatusInternalServerError, "error.html", gin.H{
				"error": fmt.Errorf("文件名不能为空"),
			})
			return
		}
		data, err := bgiStatus.GetAutoArtifactsPro2(fileName)

		// 判断是否请求 JSON 数据
		fmt.Println("=============", context.Query("json"))
		if context.Query("json") == "1" {
			if err != nil {
				context.JSON(http.StatusInternalServerError, gin.H{"error": "读取失败"})
				return
			}
			context.JSON(http.StatusOK, data)
			return
		}

		// 正常页面渲染
		context.HTML(http.StatusOK, "AutoArtifactsPro2.html", gin.H{
			"title": "狗粮日志查询",
			"items": data,
		})
	})

	//日志分析
	ginServer.GET("/logAnalysis", func(context *gin.Context) {
		res := bgiStatus.LogAnalysis()

		context.HTML(http.StatusOK, "logAnalysis.html", gin.H{
			"title": "日志分析",
			"items": res,
		})
	})

	//自动更新仓库脚本仓库和地图追踪
	ginServer.POST("/autoUpdateJsAndPathing", func(context *gin.Context) {
		err := bgiStatus.UpdateJsAndPathing()
		if err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"status": "received", "data": err})
			return
		}
		context.JSON(http.StatusOK, gin.H{"status": "received", "data": "更新成功"})
	})

	//备份文件
	ginServer.POST("/backup", func(context *gin.Context) {
		err := bgiStatus.Backup()
		if err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"status": "received", "data": err})
			return
		}
		context.JSON(http.StatusOK, gin.H{"status": "received", "data": "备份成功"})
	})

	ginServer.GET("/CalculateTaskEnabledList", func(context *gin.Context) {
		list, err := task.CalculateTaskEnabledList()
		if err != nil {
			context.String(http.StatusInternalServerError, "任务状态读取失败: %v", err)
			return
		}

		// 渲染 HTML 模板
		context.HTML(http.StatusOK, "CalculateTaskEnabledList.html", gin.H{
			"title": "配置组执行",
			"tasks": list,
		})
	})

	//统计配置组执行时间
	ginServer.GET("/other", func(context *gin.Context) {
		GroupTime, err := bgiStatus.GroupTime()
		if err != nil {
			context.HTML(http.StatusInternalServerError, "error.html", gin.H{
				"error": err.Error(),
			})
		}
		context.HTML(http.StatusOK, "other.html", gin.H{
			"title":     "其他",
			"GroupTime": GroupTime,
		})
	})

	//测试
	ginServer.GET("/test", func(context *gin.Context) {
		control.GetMysQDXy()
	})

	//读取statuc文件夹所有的图片
	ginServer.GET("/images", func(context *gin.Context) {
		currentDir, err := os.Getwd()
		autoLog.Sugar.Infof("当前目录:%s", currentDir)
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
			return
		}

		//imageDir := filepath.Join(currentDir, "static/image")

		files, err := os.ReadDir("static/image")
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
			return
		}

		var imageNames []string
		for _, file := range files {
			if !file.IsDir() {
				imageNames = append(imageNames, file.Name())
			}
		}

		context.JSON(http.StatusOK, gin.H{"status": "success", "data": imageNames})
	})

	//一条龙
	if Config.IsStartTimeLong {
		go task.OneLong()

		autoLog.Sugar.Infof("一条龙开启状态")

	} else {
		autoLog.Sugar.Infof("一条龙关闭状态")
	}

	//检查BGI状态
	go bgiStatus.CheckBetterGIStatus()

	if Config.IsMysSignIn {
		//米游社自动签到
		go task.MysSignIn()
		autoLog.Sugar.Infof("米游社自动签到开启状态")
	} else {
		autoLog.Sugar.Infof("米游社自动签到关闭状态")
	}

	//服务器端口
	post := Config.Post
	if post == "" {
		post = ":8082"
	}
	err := ginServer.Run(post)
	autoLog.Sugar.Infof("启动成功")
	if err != nil {
		return
	}

}

//go build -o auto-bgi.exe main.go
//go build -o auto-bgi.exe -ldflags="-H windowsgui" main.go
