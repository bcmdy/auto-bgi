package main

import (
	"auto-bgi/Notice"
	"auto-bgi/ScriptGroup"
	"auto-bgi/ScriptRepo"
	"auto-bgi/abgiSSE"
	"auto-bgi/autoLog"
	"auto-bgi/bgiStatus"
	"auto-bgi/config"
	"auto-bgi/control"
	"auto-bgi/internal/mysConfig"
	"auto-bgi/task"
	"auto-bgi/tools"
	"bufio"
	"embed"
	"encoding/json"
	"fmt"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"io"
	"io/fs"
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

//go:embed web/dist
var embeddedFiles embed.FS

func init() {
	// 初始化日志
	autoLog.Init()
	config.InitDB()
	defer autoLog.Sync()
	ips, err := tools.GetLocalIPs()
	if err != nil {
		autoLog.Sugar.Infof("获取本机IP失败: %v", err)
	} else {
		autoLog.Sugar.Infof("浏览器使用本机局域网IP")
		for _, ip := range ips {
			if strings.Contains(ip, "192.168") {
				autoLog.Sugar.Infof("本机局域网IP: %s%s", ip, config.Cfg.Post)
			} else {
				autoLog.Sugar.Infof("本机其他IP: %s%s", ip, config.Cfg.Post)
			}

		}
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 如果跨域就写逻辑
	},
}

var imageList []string
var imageListOnce sync.Once

func loadImages() {
	imageDir := "./img"
	filepath.WalkDir(imageDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			ext := filepath.Ext(d.Name())
			switch ext {
			case ".jpg", ".jpeg", ".png", ".gif", ".webp":
				imageList = append(imageList, "/img/"+d.Name())
			}
		}

		return nil
	})
	autoLog.Sugar.Infof("加载图片: %d", len(imageList))
}

func main() {

	gin.SetMode(gin.ReleaseMode)

	//创建一个服务
	ginServer := gin.Default()

	// 创建嵌入的文件系统
	distFS, err := fs.Sub(embeddedFiles, "web/dist")
	if err != nil {
		autoLog.Sugar.Fatalf("无法创建嵌入文件系统: %v", err)
	}

	ginServer.SetTrustedProxies(nil)
	ginServer.Use(gzip.Gzip(gzip.DefaultCompression))

	//查询今日所有日志文件
	ginServer.GET("/api/logFiles", func(c *gin.Context) {
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

	abgiWs := ginServer.Group("/api/abgiSSE")
	{
		//上线
		abgiWs.POST("/connect", func(c *gin.Context) {
			if config.Cfg.Account.Uid == "" {
				c.JSON(http.StatusBadRequest, gin.H{"message": "账号配置错误"})
				return
			}
			if config.Cfg.Account.Name == "" {
				c.JSON(http.StatusBadRequest, gin.H{"message": "账号配置错误"})
				return
			}
			if config.Cfg.Account.SecretKey == "" {
				c.JSON(http.StatusBadRequest, gin.H{"message": "密钥错误"})
				return
			}
			//解密
			decryptedKey, err3 := abgiSSE.Decrypt(config.Cfg.Account.SecretKey, config.Cfg.Account.AccountKey)
			if err3 != nil {
				c.JSON(http.StatusBadRequest, gin.H{"message": "密钥错误"})
				return
			}

			err := abgiSSE.Connect(fmt.Sprintf("ws://%s/api/abgiWs/%s/%s", decryptedKey, config.Cfg.Account.Uid, config.Cfg.Account.Name), nil)
			if err != nil {
				autoLog.Sugar.Errorf("连接失败: %v", err)
			}
			c.JSON(http.StatusOK, gin.H{"message": "连接成功"})
		})

		//下线
		abgiWs.POST("/disconnect", func(c *gin.Context) {
			abgiSSE.Close()
		})

	}

	//日志查询
	ginServer.GET("/api/index", func(c *gin.Context) {
		// 生成日志文件名
		date := time.Now().Format("20060102")

		filename := filepath.Clean(fmt.Sprintf("%s\\log\\better-genshin-impact%s.log", config.Cfg.BetterGIAddress, date))

		filePath := filepath.Clean(fmt.Sprintf("%s\\log", config.Cfg.BetterGIAddress)) // 本地日志路径
		files, err := bgiStatus.FindLogFiles(filePath)
		if len(files) == 0 {
			autoLog.Sugar.Errorf("日志文件不存在")
			c.JSON(http.StatusBadRequest, gin.H{"error": "日志文件不存在"})
			return
		}
		if err == nil {
			//获取最后一个文件
			filename = filepath.Clean(fmt.Sprintf("%s\\log\\%s", config.Cfg.BetterGIAddress, files[0]))
		}

		autoLog.Sugar.Infof("日志文件名:%s", filename)

		progress := "0/0"
		group := "未知"
		GetGroup := "未知"
		timestamp := "未知"
		onlineUser := abgiSSE.GetAllOnlineUser()

		line, err := bgiStatus.FindLastExecLine(filename)
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

		jsProgress, err := bgiStatus.JsProgress(filename, "当前进度：(.*?)", "当前次数：(.*?)")
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
		data["onlineUser"] = onlineUser

		c.JSON(http.StatusOK, data)

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

	// 删除全部归档记录
	ginServer.DELETE("/api/allArchives", func(c *gin.Context) {
		_, err := config.DB.Exec("DELETE FROM archive_records")
		if err != nil {
			c.String(http.StatusInternalServerError, "删除失败")
			return
		}

		c.String(http.StatusOK, "删除成功")
	})

	//一条龙
	ginServer.POST("/api/oneLong", func(context *gin.Context) {

		task.OneLongTask()

		context.JSON(http.StatusOK, gin.H{"status": "received", "data": "一条龙启动成功"})
	})

	ginServer.POST("/api/closeBgi", func(context *gin.Context) {

		control.CloseSoftware()

		context.JSON(http.StatusOK, gin.H{"status": "received", "data": "BGI关闭成功"})
	})

	ginServer.POST("/api/closeYuanShen", func(context *gin.Context) {

		control.CloseYuanShen()

		context.JSON(http.StatusOK, gin.H{"status": "received", "data": "原神关闭成功"})
	})

	//发送截图
	ginServer.POST("/api/sendImage", func(c *gin.Context) {

		err := control.ScreenShot()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "received", "data": "截图失败"})
			return
		} else {
			err2 := Notice.SentImage("jt.png")
			if err2 != nil {
				c.JSON(http.StatusBadRequest, gin.H{"status": "received", "data": err2})
				return
			}
			c.JSON(http.StatusOK, gin.H{"status": "received", "data": "发送成功"})
			return
		}

	})

	//背包统计
	ginServer.GET("/api/BagStatistics", func(context *gin.Context) {
		statistics, _ := bgiStatus.BagStatistics()

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

		context.JSON(http.StatusOK, statistics)

	})

	//检查背包材料是否超过8000
	ginServer.GET("/api/checkBag", func(context *gin.Context) {
		checkBag := bgiStatus.CheckBag()
		context.JSON(http.StatusOK, checkBag)
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

	//启动配置组
	ginServer.POST("/api/startGroups", func(context *gin.Context) {

		var data []string
		err := context.BindJSON(&data)
		if err != nil {
			fmt.Println("err:", err)
			return
		}

		err = task.StartGroups(data)
		if err != nil {
			return
		}
		context.JSON(http.StatusOK, gin.H{"message": "Success"})
	})

	//查询狗粮日志
	ginServer.GET("/api/getAutoArtifactsPro", func(context *gin.Context) {

		pro, err := bgiStatus.GetAutoArtifactsPro()
		autoLog.Sugar.Infof("狗粮记录:%s", pro)

		//获取版本号
		version := bgiStatus.ReadVersion(fmt.Sprintf("%s\\User\\JsScript\\AAA-Artifacts-Bulk-Supply", config.Cfg.BetterGIAddress))

		//查询更新状态
		jsVersion := bgiStatus.JsVersion("AAA-Artifacts-Bulk-Supply", version)

		if err != nil {
			// 传递给模板

			context.JSON(http.StatusInternalServerError, gin.H{
				"title":     "狗粮批发查询" + "【" + version + "】",
				"JsVersion": jsVersion,
				"items":     nil,
			})

			return
		}

		context.JSON(http.StatusOK, gin.H{
			"title":     "狗粮批发查询" + "【" + version + "】",
			"JsVersion": jsVersion,
			"items":     pro,
		})

	})

	//查询狗粮日志
	ginServer.GET("/api/getAutoArtifactsPro2", func(context *gin.Context) {

		fileName := context.Query("fileName")
		if fileName == "" {
			context.JSON(http.StatusBadRequest, gin.H{"error": "fileName不能为空"})
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

		context.JSON(http.StatusOK, gin.H{
			"items": data,
		})

	})

	//查询收获前10的材料
	ginServer.GET("/api/logAnalysis", func(context *gin.Context) {
		fileName := context.Query("file")

		res := bgiStatus.LogAnalysis(fileName)

		context.JSON(200, res)

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

	//获取仓库提交记录（最新的10条）
	ginServer.GET("/api/gitLog", func(context *gin.Context) {
		//gitLog, err := bgiStatus.GitLog(10)
		//fmt.Println(err)
		//context.JSON(http.StatusOK, gin.H{
		//	"gitLog": gitLog,
		//})
		gitLog := ScriptRepo.Read()
		context.JSON(http.StatusOK, gin.H{
			"gitLog": gitLog,
		})
	})

	// 统计配置组执行时间 - 返回JSON
	ginServer.GET("/api/other", func(context *gin.Context) {
		var otherGroup sync.WaitGroup
		otherGroup.Add(2)
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
		//
		////获取米游社签到日志
		//go func() {
		//	defer otherGroup.Done()
		//	signLog = bgiStatus.GetMysSignLog()
		//}()

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
	ginServer.GET("/api/LogAnalysis2Page", func(context *gin.Context) {
		fileName := context.Query("file")
		if fileName == "" {
			context.String(http.StatusBadRequest, "缺少 file 参数")
			return
		}

		results := bgiStatus.LogAnalysis2(fileName)

		context.JSON(http.StatusOK, gin.H{"status": "success", "data": results})
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

	//读取js的md文件
	ginServer.GET("/api/md", func(c *gin.Context) {
		filePath := c.Query("filePath")

		jsMd := bgiStatus.ReadMd(filePath)
		c.JSON(http.StatusOK, gin.H{"status": "success", "data": jsMd})

	})

	//批量更新仓库
	ginServer.GET("/api/batchUpdate", func(c *gin.Context) {
		script := bgiStatus.BatchUpdateScript()
		if script != "" {
			c.JSON(http.StatusOK, gin.H{"status": "success", "message": script})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "success", "message": "更新成功"})
	})

	//米游社手动签到
	ginServer.POST("/api/mysSignIn", func(c *gin.Context) {

		task.MiYouSheSign()

		c.JSON(http.StatusOK, gin.H{"status": "success", "message": "签到成功"})

	})

	var scriptGroupConfig ScriptGroup.ScriptGroupConfig

	//配置组api
	scriptGroup := ginServer.Group("/api/scriptGroup")
	{
		//读取配置组配置
		scriptGroup.POST("/UpdatePathing", func(c *gin.Context) {
			var updatePath config.UpdatePathing
			if err := c.ShouldBindJSON(&updatePath); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"message": "参数格式错误", "error": err.Error()})
				return
			}

			res, err := scriptGroupConfig.UpdatePathing(updatePath)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "msg": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"status": "success", "data": res})
		})

		//查询地图追踪配置
		scriptGroup.GET("/ConfigPathing", func(c *gin.Context) {

			scriptGroupConfig.ListPathingUpdatePaths()

			UpdatePathData := config.Cfg.UpdatePath

			c.JSON(http.StatusOK, gin.H{"status": "success", "data": UpdatePathData})
		})

		//保存配置
		scriptGroup.POST("/savePathing", func(c *gin.Context) {
			var updatePath []config.UpdatePathing
			if err := c.ShouldBindJSON(&updatePath); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"message": "参数格式错误", "error": err.Error()})
				return
			}
			err := scriptGroupConfig.SavePathing(updatePath)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "msg": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"status": "success", "message": "保存成功"})
		})

		//查询所有配置组
		scriptGroup.GET("/listGroups", func(context *gin.Context) {
			groups, err := task.ListGroups()
			if err != nil {
				return
			}

			context.JSON(http.StatusOK, groups)
		})

		//查询所有地图追踪文件
		scriptGroup.GET("/listAllGroups", func(context *gin.Context) {
			listAllPathing, err := scriptGroupConfig.ListAllPathing()
			if err != nil {
				return
			}
			context.JSON(http.StatusOK, gin.H{"status": "success", "data": listAllPathing})
		})

		//清理地图追踪文件
		scriptGroup.POST("/cleanAllPathing", scriptGroupConfig.CleanAllPathing)

		//读取配置组所有的地图追踪
		scriptGroup.GET("/listPathingUpdatePaths", scriptGroupConfig.UpdatePaths)

	}

	// 定义 GitHub Push Webhook 的结构体
	type GitHubWebhookPayload struct {
		Ref        string `json:"ref"`
		Repository struct {
			FullName string `json:"full_name"`
		} `json:"repository"`
		Commits []struct {
			ID        string `json:"id"`
			Message   string `json:"message"`
			Timestamp string `json:"timestamp"`
			URL       string `json:"url"`
			Author    struct {
				Name  string `json:"name"`
				Email string `json:"email"`
			} `json:"author"`
		} `json:"commits"`
	}

	//webhook
	ginServer.POST("/webhook", func(c *gin.Context) {
		var payload GitHubWebhookPayload
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
			return
		}

		branch := strings.TrimPrefix(payload.Ref, "refs/heads/")
		fmt.Println("分支:", branch)
		fmt.Println("仓库:", payload.Repository.FullName)

		for _, commit := range payload.Commits {
			GITLOG := fmt.Sprintf("Git通知=====提交ID: %s\n消息: %s\n作者: %s\n时间: %s\nURL: %s\n",
				commit.ID, commit.Message, commit.Author.Name, commit.Timestamp, commit.URL)
			autoLog.Sugar.Infof(GITLOG)
			// 发送通知
			Notice.SentText(GITLOG)
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	//检查BGI状态
	go bgiStatus.CheckBetterGIStatus()

	//开启每隔一小时发送截图
	if config.Cfg.Control.SendWeChatImage {
		autoLog.Sugar.Infof("开启每隔一小时发送截图")
		go task.SendWeChatImageTask()
	} else {
		autoLog.Sugar.Infof("关闭每隔一小时发送截图")
	}

	//实时读取文件
	go bgiStatus.LogM()

	if config.Cfg.OneRemote.IsMonitor {
		go bgiStatus.Log1Remote()
		autoLog.Sugar.Infof("1Remote监控开启状态")
	}

	//米游社自动签到
	mysConfig.LoadConfig("mysConfig.yaml")
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

	//获取机器码
	machineCode, err := tools.GetMachineCode()
	if err != nil {
		autoLog.Sugar.Error("获取机器码失败:", err)
	} else {
		autoLog.Sugar.Infof("机器码: %s", machineCode)
	}

	// 1. 静态资源挂载（直接让前端可以访问图片）
	ginServer.Static("/img", "./img")

	imageListOnce.Do(loadImages) // 只加载一次

	// 2. API：返回所有图片的 URL
	ginServer.GET("/api/images", func(c *gin.Context) {
		c.Header("Cache-Control", "public, max-age=86400")
		c.Header("Expires", time.Now().AddDate(0, 0, 3).Format(http.TimeFormat))
		c.JSON(200, gin.H{"images": imageList})

	})

	// 静态文件服务（放在所有API路由之后）
	ginServer.StaticFS("/assets", http.FS(distFS))

	// Vue Router history 支持和静态文件服务
	ginServer.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.JSON(http.StatusNotFound, gin.H{"error": "API not found"})
			return
		}

		// 尝试从嵌入文件系统中获取请求的文件
		requestPath := strings.TrimPrefix(c.Request.URL.Path, "/")
		if requestPath == "" {
			requestPath = "index.html"
		}

		file, err := distFS.Open(requestPath)
		if err != nil {
			// 如果文件不存在，返回index.html（SPA支持）
			indexFile, err := distFS.Open("index.html")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "无法读取index.html"})
				return
			}
			defer indexFile.Close()

			indexContent, err := io.ReadAll(indexFile)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "无法读取index.html内容"})
				return
			}

			c.Data(http.StatusOK, "text/html; charset=utf-8", indexContent)
			return
		}
		defer file.Close()

		// 读取并返回请求的文件
		content, err := io.ReadAll(file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "无法读取文件内容"})
			return
		}

		// 根据文件扩展名设置Content-Type
		contentType := "application/octet-stream"
		if strings.HasSuffix(requestPath, ".html") {
			contentType = "text/html; charset=utf-8"
		} else if strings.HasSuffix(requestPath, ".css") {
			contentType = "text/css; charset=utf-8"
		} else if strings.HasSuffix(requestPath, ".js") {
			contentType = "application/javascript; charset=utf-8"
		} else if strings.HasSuffix(requestPath, ".json") {
			contentType = "application/json; charset=utf-8"
		} else if strings.HasSuffix(requestPath, ".png") {
			contentType = "image/png"
		} else if strings.HasSuffix(requestPath, ".jpg") || strings.HasSuffix(requestPath, ".jpeg") {
			contentType = "image/jpeg"
		} else if strings.HasSuffix(requestPath, ".ico") {
			contentType = "image/x-icon"
		}

		c.Data(http.StatusOK, contentType, content)
	})

	if len(os.Args) > 1 {
		if os.Args[1] == "oneLong" {
			task.OneLongTask()
			autoLog.Sugar.Infof("一条龙启动")
		}
	}

	//服务器端口
	post := config.Cfg.Post
	if post == "" {
		post = ":8082"
	}
	err = ginServer.Run(post)

	if err != nil {
		autoLog.Sugar.Errorf("启动失败:%v", err)
		return
	}

	//err = ginServer.RunTLS(post, "certFile/cert.pem", "certFile/key.pem")
	//if err != nil {
	//	autoLog.Sugar.Errorf("启动失败:%v", err)
	//}

}

//前端打包
//cd web
//npm run build

//后端打包：
//go build

//打包脚本
//  build.bat
