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
	"flag"
	"fmt"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/thinkerou/favicon"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

func init() {
	// 初始化日志
	autoLog.Init()
	config.InitDB()
	defer autoLog.Sync()

	////判断目录是否设置正确
	//exists, err := bgiStatus.CheckConfig()
	//if !exists {
	//	fmt.Println(err)
	//	//程序暂停，任意键退出
	//	fmt.Println("=======程序暂停，任意键退出=========")
	//	fmt.Scanln()
	//	os.Exit(1)
	//}
}

func findLastJSONLine(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "未知", err
	}
	defer file.Close()

	var lastJSONLine string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "→ 脚本执行") {
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

func toJson(v interface{}) template.JS {
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

//go:embed html/*
var htmlFS embed.FS

//go:embed static/*
var staticFiles embed.FS

func main() {

	useEmbed := flag.Bool("embed", false, "是否将 static 目录打包进程序")
	flag.Parse()

	gin.SetMode(gin.ReleaseMode)

	//创建一个服务
	ginServer := gin.Default()

	ginServer.SetTrustedProxies(nil)
	ginServer.Use(gzip.Gzip(gzip.DefaultCompression))
	ginServer.Use(favicon.New("./favicon.ico"))

	////加载templates中所有模板文件, 使用不同目录下名称相同的模板,注意:一定要放在配置路由之前才得行
	//ginServer.LoadHTMLGlob("html/*")

	tmpl := template.Must(
		template.New("").Funcs(template.FuncMap{
			"tojson": toJson,
		}).ParseFS(htmlFS, "html/*.html"),
	)

	ginServer.SetHTMLTemplate(tmpl)

	if *useEmbed {
		// 使用 embed 打包的静态文件
		staticFS := http.FS(staticFiles)
		ginServer.StaticFS("/static", staticFS)
		ginServer.GET("/test", func(c *gin.Context) {
			c.String(200, "使用的是 embed 打包的 static")
		})
	} else {
		// 使用本地目录（开发模式）
		ginServer.Static("/static", "static")
		ginServer.GET("/test", func(c *gin.Context) {
			c.String(200, "使用的是本地 static 目录")
		})
	}

	//// 提供静态资源服务，把 html 目录映射为 /static 路径
	//ginServer.Static("/static", "static")

	ginServer.GET("/log", func(context *gin.Context) {

		// 传递给模板
		context.HTML(http.StatusOK, "log.html", nil)
	})

	//查询今日所有日志文件
	ginServer.GET("/logFiles", func(c *gin.Context) {
		filePath := filepath.Clean(fmt.Sprintf("%s\\log", config.Cfg.BetterGIAddress)) // 本地日志路径
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

		filePath := filepath.Join(config.Cfg.BetterGIAddress, "log", logName)
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

		filename := filepath.Clean(fmt.Sprintf("%s\\log\\better-genshin-impact%s.log", config.Cfg.BetterGIAddress, date))

		filePath := filepath.Clean(fmt.Sprintf("%s\\log", config.Cfg.BetterGIAddress)) // 本地日志路径
		files, err := bgiStatus.FindLogFiles(filePath)
		fmt.Println(files)
		if err == nil {
			//获取最后一个文件
			filename = filepath.Clean(fmt.Sprintf("%s\\log\\%s", config.Cfg.BetterGIAddress, files[0]))
		}

		autoLog.Sugar.Infof("日志文件名:%s", filename)

		progress := "0/0"
		group := "未知"
		GetGroup := "未知"
		timestamp := "未知"

		line, err := findLastJSONLine(filename)
		if err != nil {
			autoLog.Sugar.Errorf("findLastJSONLine-Error: %v\n", err)
		} else {
			group, timestamp, err = bgiStatus.FindLastGroup(filename)
			if err != nil {
				autoLog.Sugar.Errorf("配置组查不到: %v\n", err)
			} else {
				calculateTime, err := bgiStatus.CalculateTime(filename, group, timestamp)
				if err != nil {
					timestamp = "未知"
				} else {
					timestamp = calculateTime
				}
				jsonStr := fmt.Sprintf("%s\\User\\ScriptGroup\\%s", config.Cfg.BetterGIAddress, group+".json")
				progress, err = bgiStatus.Progress(jsonStr, line)
				if err != nil {
					autoLog.Sugar.Errorf("%v\n", err)
				}

			}
			GetGroup = bgiStatus.GetGroupP(group)
		}

		running := bgiStatus.IsWechatRunning()

		jsProgress, err := bgiStatus.JsProgress(filename, "当前进度：(.*?)")
		if err != nil {
			jsProgress = "无"
		}

		data := make(map[string]interface{})
		data["group"] = group + "[" + GetGroup + "]"
		data["ExpectedToEnd"] = timestamp
		data["line"] = line
		data["progress"] = progress
		data["running"] = running
		data["jsProgress"] = jsProgress

		c.JSON(http.StatusOK, data)

	})

	ginServer.GET("/archive", func(c *gin.Context) {
		// 传递给模板
		c.HTML(http.StatusOK, "archive.html", nil)
	})

	//查询归档列表查询
	ginServer.GET("/api/archiveList", func(c *gin.Context) {
		// 调用函数获取数据
		archive := bgiStatus.ListArchive()
		c.JSON(http.StatusOK, archive)
	})

	// 删除归档记录
	ginServer.DELETE("/api/archive", func(c *gin.Context) {
		idStr := c.Query("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.String(http.StatusBadRequest, "无效的ID")
			return
		}

		_, err = config.DB.Exec("DELETE FROM archive_records WHERE id = ?", id)
		if err != nil {
			c.String(http.StatusInternalServerError, "删除失败")
			return
		}

		c.String(http.StatusOK, "删除成功")
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

		autoLog.Sugar.Infof("webhook:%s", j["message"])

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

		//获取版本号
		version := bgiStatus.ReadVersion(fmt.Sprintf("%s\\User\\JsScript\\AutoArtifactsPro", config.Cfg.BetterGIAddress))

		//查询更新状态
		jsVersion := bgiStatus.JsVersion("AutoArtifactsPro", version)

		if err != nil {
			// 传递给模板
			context.HTML(http.StatusOK, "AutoArtifactsPro.html", gin.H{
				"title":     "狗粮收益查询" + "【" + version + "】",
				"JsVersion": jsVersion,
				"items":     nil,
			})
			return
		}
		context.HTML(http.StatusOK, "AutoArtifactsPro.html", gin.H{
			"title":     "狗粮收益查询" + "【" + version + "】",
			"JsVersion": jsVersion,
			"items":     pro,
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

		context.HTML(http.StatusOK, "logAnalysis.html", nil)
	})

	ginServer.GET("/api/logAnalysis", func(context *gin.Context) {
		fileName := context.Query("file")

		res := bgiStatus.LogAnalysis(fileName)

		context.JSON(200, res)

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

	ginServer.GET("/other", func(context *gin.Context) {
		context.HTML(http.StatusOK, "other.html", nil)
	})

	//获取仓库提交记录（最新的10条）
	ginServer.GET("/api/gitLog", func(context *gin.Context) {
		gitLog := bgiStatus.GitLog()
		context.JSON(http.StatusOK, gin.H{
			"gitLog": gitLog,
		})
	})

	// 统计配置组执行时间 - 返回JSON
	ginServer.GET("/api/other", func(context *gin.Context) {
		var otherGroup sync.WaitGroup
		otherGroup.Add(3)
		fileName := context.Query("file")

		var (
			GroupTime  []bgiStatus.GroupMap
			signLog    string
			groupPInfo string
			//gitLog     []bgiStatus.GitLogStruct
		)

		//获取配置组执行时长
		go func() {
			defer otherGroup.Done()
			GroupTime, _ = bgiStatus.GroupTime(fileName)
		}()

		//获取米游社签到日志
		go func() {
			defer otherGroup.Done()
			signLog = bgiStatus.GetMysSignLog()
		}()

		//获取今天执行配置组
		go func() {
			defer otherGroup.Done()
			groupPInfo = bgiStatus.GetGroupPInfo()
		}()

		otherGroup.Wait() // 等待所有 goroutine 完成

		context.JSON(http.StatusOK, gin.H{
			"GroupTime":  GroupTime,
			"signLog":    signLog,
			"groupPInfo": groupPInfo,
		})
	})

	//读取statuc文件夹所有的图片
	ginServer.GET("/images", func(context *gin.Context) {

		files, err := os.ReadDir("./static/image")

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

	ginServer.POST("/api/archive", func(c *gin.Context) {
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "参数解析失败: " + err.Error()})
			return
		}
		bgiStatus.Archive(req)

		c.String(200, fmt.Sprintf("成功归档 %d 条记录"))
	})

	//日志分析
	ginServer.GET("/LogAnalysis2Page", func(context *gin.Context) {

		context.HTML(http.StatusOK, "log_analysis.html", nil)
	})

	ginServer.GET("/api/LogAnalysis2Page", func(context *gin.Context) {
		fileName := context.Query("file")
		if fileName == "" {
			context.String(http.StatusBadRequest, "缺少 file 参数")
			return
		}

		results := bgiStatus.LogAnalysis2(fileName)

		context.JSON(http.StatusOK, gin.H{"status": "success", "data": results})
	})

	ginServer.GET("/jsNames", func(context *gin.Context) {
		context.HTML(http.StatusOK, "jsNames.html", nil)
	})

	//查询关注脚本情况
	ginServer.GET("/api/jsNames", func(context *gin.Context) {

		jsNamesInfo := bgiStatus.JsNamesInfo()

		context.JSON(http.StatusOK, gin.H{"status": "success", "data": jsNamesInfo})
	})

	//脚本Js更新
	ginServer.POST("/api/updateJs", func(context *gin.Context) {

		var req struct {
			Name string `json:"name"`
		}
		if err := context.ShouldBindJSON(&req); err != nil || req.Name == "" {
			context.JSON(400, gin.H{"success": false, "message": "无效的请求参数"})
			return
		}

		autoLog.Sugar.Infof("更新插件:%s", req.Name)
		_, err := bgiStatus.UpdateJs(req.Name)
		if err != nil {
			// 成功返回
			context.JSON(400, gin.H{"err": err})
			return
		}

		// 成功返回
		context.JSON(200, gin.H{"success": true})

	})

	ginServer.GET("/Config", func(context *gin.Context) {
		context.HTML(http.StatusOK, "Config.html", nil)
	})

	//查询配置文件
	ginServer.GET("/api/config", func(context *gin.Context) {
		cfg := config.Cfg
		context.JSON(http.StatusOK, gin.H{"status": "success", "data": cfg})
	})

	ginServer.POST("/api/saveConfig", func(c *gin.Context) {
		var newConfig config.Config

		if err := c.ShouldBindJSON(&newConfig); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "参数格式错误", "error": err.Error()})
			return
		}

		// 序列化为JSON字符串，格式化输出
		data, err := json.MarshalIndent(newConfig, "", "  ")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "序列化失败", "error": err.Error()})
			return
		}

		// 写入 main.json，路径可以自定义，这里示例写当前运行目录
		filePath := filepath.Join(".", "main.json")
		err = os.WriteFile(filePath, data, 0644)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "写文件失败", "error": err.Error()})
			return
		}

		fmt.Println("配置保存成功:", newConfig)

		//重新加载配置文件
		_ = config.ReloadConfig()
		time.Sleep(1 * time.Second)
		//c.JSON(http.StatusOK, gin.H{"status": "success", "message": "配置保存成功"})

		// 调用重启脚本
		cmd := exec.Command("cmd", "/c", "restart.bat")
		err2 := cmd.Start()
		if err2 != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err2.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "success", "message": "重启命令已执行"})

	})

	//读取所有一条龙配置
	ginServer.GET("/api/oneLongAllName", func(context *gin.Context) {
		oneLongInfo := config.OneLongAllName()
		context.JSON(http.StatusOK, gin.H{"status": "success", "data": oneLongInfo})
	})

	//查询所有天赋书
	ginServer.GET("/api/talentBooks", func(context *gin.Context) {

		td := &bgiStatus.TalentDomain{}
		talents, _ := td.QueryAllTalents()

		context.JSON(http.StatusOK, gin.H{"status": "success", "data": talents})
	})

	ginServer.GET("/api/talentBooks/search", func(c *gin.Context) {
		name := c.Query("name")
		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "msg": "缺少参数 name"})
			return
		}

		query := `SELECT domain_name, weekday, material_name FROM talent_domains WHERE material_name = ?`
		rows, err := config.DB.Query(query, name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "msg": err.Error()})
			return
		}
		defer rows.Close()

		var results []bgiStatus.TalentDomain
		for rows.Next() {
			var td bgiStatus.TalentDomain
			if err := rows.Scan(&td.DomainName, &td.Weekday, &td.MaterialName); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "msg": err.Error()})
				return
			}
			results = append(results, td)
		}

		if len(results) == 0 {
			c.JSON(http.StatusOK, gin.H{"status": "not_found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"data":   results,
		})
	})

	ginServer.GET("/onelong", func(context *gin.Context) {
		context.HTML(http.StatusOK, "onelong.html", nil)
	})

	// 获取一条龙配置
	ginServer.GET("/api/onelong/config", func(c *gin.Context) {
		name := c.Query("name")
		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "缺少参数 name"})
			return
		}

		cfg := config.OneLongConfig(name)
		c.JSON(http.StatusOK, cfg)
	})

	//保存一条龙配置
	ginServer.POST("/api/onelong/saveConfig", func(c *gin.Context) {
		var newConfig config.OneLongConfigStruct

		if err := c.ShouldBindJSON(&newConfig); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "参数格式错误", "error": err.Error()})
			return
		}

		// 保存配置
		err := config.SaveOneLongConfig(newConfig)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "保存失败", "error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "success", "message": "保存成功"})
	})

	//武器材料

	//查询所有武器升级材料
	ginServer.GET("/api/WeaponDomain", func(context *gin.Context) {

		td := &bgiStatus.WeaponDomain{}
		talents, _ := td.QueryAllWeaponMaterials()

		context.JSON(http.StatusOK, gin.H{"status": "success", "data": talents})
	})

	ginServer.GET("/api/WeaponDomain/search", func(c *gin.Context) {
		name := c.Query("name")
		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "msg": "缺少参数 name"})
			return
		}

		query := `SELECT domain_name, weekday, material_name FROM weapon_domains WHERE material_name = ?`
		rows, err := config.DB.Query(query, name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "msg": err.Error()})
			return
		}
		defer rows.Close()

		var results []bgiStatus.TalentDomain
		for rows.Next() {
			var td bgiStatus.TalentDomain
			if err := rows.Scan(&td.DomainName, &td.Weekday, &td.MaterialName); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "msg": err.Error()})
				return
			}
			results = append(results, td)
		}

		if len(results) == 0 {
			c.JSON(http.StatusOK, gin.H{"status": "not_found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"data":   results,
		})
	})

	//检查BGI状态
	go bgiStatus.CheckBetterGIStatus()
	//更新仓库
	go func() {
		err := bgiStatus.GitPull()
		if err != nil {
			autoLog.Sugar.Errorf("更新仓库失败:%v", err)
		}
	}()

	//开启每隔一小时发送截图
	if config.Cfg.Control.SendWeChatImage {
		autoLog.Sugar.Infof("开启每隔一小时发送截图")
		go task.SendWeChatImageTask()
	} else {
		autoLog.Sugar.Infof("关闭每隔一小时发送截图")
	}

	//实时读取文件
	go bgiStatus.ReadLog()

	go task.UpdateCode()

	//米游社自动签到
	if config.Cfg.MySign.IsMySignIn {

		go task.MysSignIn()
		autoLog.Sugar.Infof("米游社自动签到开启状态")
	} else {
		autoLog.Sugar.Infof("米游社自动签到关闭状态")
	}

	//一条龙
	if config.Cfg.OneLong.IsStartTimeLong {
		go task.OneLong()
		autoLog.Sugar.Infof("一条龙开启状态")

	} else {
		autoLog.Sugar.Infof("一条龙关闭状态")
	}

	//服务器端口
	post := config.Cfg.Post
	if post == "" {
		post = ":8082"
	}
	err := ginServer.Run(post)
	autoLog.Sugar.Infof("启动成功")
	if err != nil {
		return
	}

}

//go build
//go build -embed

//测试
