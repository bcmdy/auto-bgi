package menu

import (
	"auto-bgi/ArtifactsBulkSupply"
	"auto-bgi/BackpackStatistics"
	"auto-bgi/CDAwareAutoGather"
	"auto-bgi/CDCollectionManagement"
	"auto-bgi/JsAPI"
	"auto-bgi/Mihoyo"
	"auto-bgi/Notice"
	"auto-bgi/Ocr"
	"auto-bgi/OneLong"
	"auto-bgi/ScriptGroup"
	"auto-bgi/ScriptRepo"
	"auto-bgi/TaskCron"
	"auto-bgi/abgiFunction"
	"auto-bgi/abgiObs"
	"auto-bgi/abgiSSE"
	"auto-bgi/auth"
	"auto-bgi/autoLog"
	"auto-bgi/bgiStatus"
	"auto-bgi/config"
	"auto-bgi/control"
	"auto-bgi/models"
	"auto-bgi/task"
	"auto-bgi/tools"
	"auto-bgi/warehouse"
	"bufio"
	"crypto/md5"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/vcaesar/screenshot"
	"image/png"
	"io"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/getlantern/systray"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/go-toast/toast"
	"github.com/gorilla/websocket"
	"golang.org/x/sys/windows"
)

var userIP = "http://localhost" + config.Cfg.Post

//go:embed favicon.ico
var favicon embed.FS

var iconData []byte

func OnReady() {
	// 获取当前程序所在目录，计算哈希，确保同一目录下只启动一个
	exePath, _ := os.Executable()
	dir := filepath.Dir(exePath)
	hasher := md5.New()
	hasher.Write([]byte(dir))
	mutexName := "Global\\AutoBgiMutex_" + hex.EncodeToString(hasher.Sum(nil))

	_, errMutex := windows.CreateMutex(nil, false, windows.StringToUTF16Ptr(mutexName))
	if errMutex == windows.ERROR_ALREADY_EXISTS {
		notification := toast.Notification{
			AppID:   "AutoBGI",
			Title:   "重复启动",
			Message: "该目录下的程序已经在运行了！",
		}
		notification.Push()
		os.Exit(0)
	}

	// 设置托盘图标
	var err error
	iconData, err = favicon.ReadFile("favicon.ico") // 这里使用 embed.FS 的 ReadFile 方法
	if err != nil {
		log.Fatal(err)
	}
	systray.SetIcon(iconData)

	systray.SetTitle("我的Go程序")
	systray.SetTooltip("运行中...")

	// 菜单
	mHello := systray.AddMenuItem("打开网页", "打开网页")
	mQuit := systray.AddMenuItem("退出", "退出程序")

	go func() {
		for {
			select {
			case <-mHello.ClickedCh:
				autoLog.Sugar.Infof("正在打开浏览器")
				if err := tools.OpenBrowser(userIP); err != nil {
					autoLog.Sugar.Errorf("打开浏览器失败: %v", err)
				}
			case <-mQuit.ClickedCh:
				tools.CloseSoftware("frpc.exe")
				systray.Quit()
				return
			}
		}
	}()

	go StarGin()
}

func OnExit() {
	// 清理逻辑
	log.Println("程序退出")
}

//go:embed web/dist
var embeddedFiles embed.FS

func init() {
	// 初始化日志
	autoLog.Init()
	models.InitDB()
	abgiFunction.InitFunction()
	defer autoLog.Sync()

	//
	TaskCron.Task["更新aBgi"] = func(data string) {
		autoLog.Sugar.Infof("定时任务启动：更新aBgi-现在时间:%s 参数:[%s]", time.Now().Format("15:04:05"), data)
		err := Update()
		if err != nil {
			autoLog.Sugar.Errorf("更新aBgi失败: %v", err)
		} else {
			go func() {
				// 调用 run_auto_bgi.vbs 脚本来启动新的 auto-bgi.exe 程序
				if err := tools.RestartProgram(); err != nil {
					autoLog.Sugar.Error(err.Error())
					//return fmt.Errorf("重启程序失败: %v", err)
				}
			}()
		}

	}

	TaskCron.TmStart()

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

	entries, err := os.ReadDir(imageDir)
	if err != nil {
		autoLog.Sugar.Errorf("读取目录失败: %v", err)
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := filepath.Ext(entry.Name())
		switch ext {
		case ".jpg", ".jpeg", ".png", ".gif", ".webp":
			imageList = append(imageList, "/img/"+entry.Name())
		}
	}

	autoLog.Sugar.Infof("加载图片: %d", len(imageList))
}

func isSandbox() bool {
	if runtime.NumCPU() <= 1 {
		return true
	}

	return false
}

func isDebugged() bool {
	kernel32 := windows.NewLazySystemDLL("kernel32.dll")

	// IsDebuggerPresent
	isDebuggerPresent := kernel32.NewProc("IsDebuggerPresent")
	ret, _, _ := isDebuggerPresent.Call()
	if ret != 0 {
		return true
	}

	// CheckRemoteDebuggerPresent
	checkRemoteDebuggerPresent := kernel32.NewProc("CheckRemoteDebuggerPresent")

	hProcess := windows.CurrentProcess()
	var debug uint32

	r1, _, _ := checkRemoteDebuggerPresent.Call(
		uintptr(hProcess),
		uintptr(unsafe.Pointer(&debug)),
	)

	if r1 != 0 && debug != 0 {
		return true
	}

	return false
}

func StarGin() {

	//if isDebugged() {
	//	os.Exit(0)
	//}

	gin.SetMode(gin.ReleaseMode)

	//创建一个服务
	ginServer := gin.New()

	// 创建嵌入的文件系统
	distFS, err := fs.Sub(embeddedFiles, "web/dist")
	if err != nil {
		autoLog.Sugar.Fatalf("无法创建嵌入文件系统: %v", err)
	}

	// 配置CORS中间件 - 支持前后端分离
	ginServer.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	ginServer.SetTrustedProxies(nil)
	ginServer.Use(gzip.Gzip(gzip.DefaultCompression))

	authApi := ginServer.Group("/api/auth")
	{
		//登录
		authApi.POST("/login", auth.LoginService)

		//获取系统配置
		authApi.GET("/getSystemConfig", auth.GetSystemConfig)

	}

	var dogFood = ArtifactsBulkSupply.DogFood{}

	var OneLongService OneLong.OneLong

	var scriptGroupConfig ScriptGroup.ScriptGroupConfig

	var CDAwareAutoGatherService CDAwareAutoGather.UidInfo

	var pickBlackListsService bgiStatus.PickBlackLists

	var videoInfoService abgiObs.VideoInfo

	needAuth := ginServer.Group("/api", auth.AuthMiddleware())
	{
		/**
		 * 日志相关
		 */
		//查询所有日志文件
		needAuth.GET("/logFiles", bgiStatus.GetLogFiles)
		//读取日志
		needAuth.GET("/logInfo", bgiStatus.GetLogFileContent)

		/**
		 * 联机相关
		 */
		abgiWs := needAuth.Group("/abgiSSE")
		{
			//上线
			abgiWs.POST("/connect/:typeKey", abgiSSE.Online)

			//下线
			abgiWs.POST("/disconnect", func(c *gin.Context) {
				abgiSSE.ABgiSeeStatus = "手动下线"
				abgiSSE.Close()
			})

			//获取在线人员
			abgiWs.GET("/getOnlineUser", func(context *gin.Context) {
				onlineUser := abgiSSE.GroupsStatusHandler()
				context.JSON(http.StatusOK, onlineUser)
			})

			//联机清0
			abgiWs.POST("/clearNumberOfLaunches", func(context *gin.Context) {
				abgiSSE.RunCount = 0
				context.JSON(http.StatusOK, gin.H{"status": "received", "data": "清0成功"})
			})

			//获取联机状态
			abgiWs.GET("/getOnlineStatus", func(context *gin.Context) {

				onlineStatus := abgiSSE.ABgiSeeStatus
				context.JSON(http.StatusOK, onlineStatus)
			})

			//举报炸弹
			abgiWs.POST("/reportBomb", abgiSSE.ReportBomb)

			//发送消息
			abgiWs.POST("/sendMessage", func(context *gin.Context) {
				value := context.Query("msg")
				err := abgiSSE.Send(value)
				if err != nil {
					context.JSON(http.StatusBadRequest, gin.H{"status": "received", "data": err.Error()})
					return
				}
				context.JSON(http.StatusOK, gin.H{"status": "received", "data": "发送成功"})
			})

		}
		//ArtifactsBulkSupply.UpdateRevenue("174300")

		//获取上线次数
		ginServer.GET("/api/NumberOfLaunches", abgiSSE.NumberOfLaunches)

		needAuth.GET("/index", bgiStatus.GetIndex)

		//app获取信息
		needAuth.GET("/appInfo", GetAppInfo)

		//首页刷新
		needAuth.GET("/indexSX", func(c *gin.Context) {
			//bgiStatus.InitBgiLogStatus()
			if err := tools.RestartProgram(); err != nil {
				autoLog.Sugar.Error(err.Error())
			}
			c.JSON(http.StatusOK, "重启中")
		})

		//查询归档列表查询
		needAuth.GET("/archiveList", func(c *gin.Context) {
			// 调用函数获取数据
			archive := bgiStatus.ListArchive()

			c.JSON(http.StatusOK, archive)
		})

		// 删除归档记录
		needAuth.DELETE("/archive", func(c *gin.Context) {
			idStr := c.Query("id")
			id, err := strconv.Atoi(idStr)
			if err != nil {
				c.String(http.StatusBadRequest, "无效的ID")
				return
			}

			err = models.DeleteArchiveRecord(id)
			if err != nil {
				c.String(http.StatusInternalServerError, "删除失败")
				return
			}

			c.String(http.StatusOK, "删除成功")
		})

		// 删除全部归档记录
		needAuth.DELETE("/allArchives", func(c *gin.Context) {

			err := models.DeleteAllArchiveRecords()
			if err != nil {
				c.String(http.StatusInternalServerError, "全部删除失败")
				return
			}

			c.String(http.StatusOK, "全部删除成功")
		})

		needAuth.POST("/closeBgi", func(context *gin.Context) {

			control.CloseSoftware()

			control.CloseYuanShen()

			context.JSON(http.StatusOK, gin.H{"status": "received", "data": "BGI关闭成功"})
		})

		needAuth.POST("/closeYuanShen", func(context *gin.Context) {

			control.CloseYuanShen()

			context.JSON(http.StatusOK, gin.H{"status": "received", "data": "原神关闭成功"})
		})

		//发送截图
		needAuth.POST("/sendImage", func(c *gin.Context) {

			err := control.ScreenShot("jt.jpg")
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"status": "received", "data": "截图失败"})
				return
			} else {
				err2 := Notice.SentImage("jt.jpg")
				if err2 != nil {
					c.JSON(http.StatusBadRequest, gin.H{"status": "received", "data": err2})
					return
				}
				c.JSON(http.StatusOK, gin.H{"status": "received", "data": "发送成功"})
				return
			}

		})

		BagStatistics := needAuth.Group("/BagStatistics")
		{
			BagStatistics.GET("", func(context *gin.Context) {
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

			//查询米游社背包信息
			BagStatistics.GET("/getBagInfo", func(context *gin.Context) {
				getBagInfo := config.GetBagInfo()
				context.JSON(http.StatusOK, getBagInfo)
			})

			// 删除关注的材料
			BagStatistics.DELETE("/DELETE", BackpackStatistics.DeleteService)
			//添加关注的材料
			BagStatistics.POST("/ADD", BackpackStatistics.AddService)
			//清空所有
			BagStatistics.POST("/CLEAR", BackpackStatistics.ClearAllService)

			//查询摩拉收益
			BagStatistics.GET("/Morale", task.QueryMoraleRecord)

			//更新摩拉收益
			BagStatistics.POST("/updateMorale", task.UpdateMoraleRecord)

		}

		//吃药统计
		needAuth.GET("/EatStatistics", bgiStatus.EatStatisticsList)

		//检查背包材料是否超过8000
		needAuth.GET("/checkBag", func(context *gin.Context) {
			checkBag := bgiStatus.CheckBag()
			context.JSON(http.StatusOK, checkBag)
		})

		//删除背包统计记录
		needAuth.POST("/deleteBag", func(context *gin.Context) {
			isOk := bgiStatus.DeleteBagStatistics()

			data := gin.H{
				"message": isOk,
			}

			context.JSON(http.StatusOK, data)
		})

		//启动配置组
		needAuth.POST("/startGroups", func(context *gin.Context) {

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
		needAuth.GET("/getAutoArtifactsPro", func(context *gin.Context) {

			pro, err := bgiStatus.GetAutoArtifactsPro()

			//获取版本号
			version := bgiStatus.ReadVersion(fmt.Sprintf("%s\\User\\JsScript\\AAA-Artifacts-Bulk-Supply", config.Cfg.BetterGIAddress))

			//查询更新状态
			jsVersion := bgiStatus.JsVersion("AAA-Artifacts-Bulk-Supply", version)

			if err != nil {
				// 传递给模板
				context.JSON(http.StatusInternalServerError, gin.H{
					"title":     "狗粮批发-联机收益查询" + "【" + version + "】",
					"JsVersion": jsVersion,
					"items":     nil,
				})

				return
			}

			context.JSON(http.StatusOK, gin.H{
				"title":     "狗粮批发-联机收益查询" + "【" + version + "】",
				"JsVersion": jsVersion,
				"items":     pro,
			})

		})

		//查询狗粮日志
		needAuth.GET("/getAutoArtifactsPro2", func(context *gin.Context) {

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
		needAuth.GET("/logAnalysis", func(context *gin.Context) {
			fileName := context.Query("file")

			res := bgiStatus.LogAnalysis(fileName)

			context.JSON(200, res)

		})

		//备份文件
		needAuth.POST("/backup", func(context *gin.Context) {
			err := bgiStatus.Backup()
			if err != nil {
				context.JSON(http.StatusBadRequest, gin.H{"status": "received", "data": err})
				return
			}
			context.JSON(http.StatusOK, gin.H{"status": "received", "data": "备份成功"})
		})

		//获取仓库提交记录
		needAuth.GET("/gitLog", func(context *gin.Context) {

			gitLog := ScriptRepo.Read()
			context.JSON(http.StatusOK, gin.H{
				"gitLog": gitLog,
			})
		})

		// 统计配置组执行时间 - 返回JSON
		needAuth.GET("/other", func(context *gin.Context) {
			var otherGroup sync.WaitGroup
			otherGroup.Add(2)
			fileName := context.Query("file")

			var (
				GroupTime  []bgiStatus.LogAnalysis2Struct
				signLog    string
				groupPInfo string
			)

			//获取配置组执行时长
			go func() {
				defer otherGroup.Done()
				GroupTime = bgiStatus.GroupTime(fileName)
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

		needAuth.POST("/archive", func(c *gin.Context) {
			var req bgiStatus.LogAnalysis2Struct
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(400, gin.H{"error": "参数解析失败: " + err.Error()})
				return
			}
			bgiStatus.Archive(req)

			bgiStatus.InitBgiLogStatus()

			c.String(200, fmt.Sprintf("成功归档"))
		})

		//日志分析
		needAuth.GET("/LogAnalysis2Page", func(context *gin.Context) {
			fileName := context.Query("file")
			if fileName == "" {
				context.String(http.StatusBadRequest, "缺少 file 参数")
				return
			}

			results := bgiStatus.LogAnalysis2(fileName)

			context.JSON(http.StatusOK, gin.H{"status": "success", "data": results})
		})

		//查询脚本情况
		needAuth.GET("/jsNames", func(context *gin.Context) {

			jsNamesInfo := bgiStatus.JsNamesInfo()

			context.JSON(http.StatusOK, gin.H{"status": "success", "data": jsNamesInfo})
		})

		//查询所有脚本
		needAuth.GET("/jsNamesAll", ScriptRepo.QueryAllScript)

		//重置仓库
		needAuth.POST("/repo/resetRepo", warehouse.RepoReset)

		//订阅脚本
		needAuth.POST("/repo/subscribe", warehouse.SubscribeScript)

		//脚本Js更新
		needAuth.POST("/updateJs/:name", func(context *gin.Context) {
			name := context.Param("name")

			autoLog.Sugar.Infof("更新插件:%s", name)
			_, err := bgiStatus.UpdateJs(name)
			if err != nil {
				// 成功返回
				context.JSON(400, gin.H{"err": err})
				return
			}

			// 成功返回
			context.JSON(200, gin.H{"success": true})

		})

		//bgi配置
		bgiConfigCotroller := needAuth.Group("/bgiConfig")
		{
			//查询所有配置文件
			bgiConfigCotroller.GET("/allConfig", bgiStatus.ReadBgiConfigAll)
			//查询配置文件
			bgiConfigCotroller.GET("/findConfig", bgiStatus.ReadBgiConfig)
			//保存配置文件
			bgiConfigCotroller.POST("/saveConfig", bgiStatus.ModifyBgiConfig)

		}

		//查询配置文件
		needAuth.GET("/config", func(context *gin.Context) {
			cfg := config.Cfg
			context.JSON(http.StatusOK, gin.H{"status": "success", "data": cfg})
		})

		//aBgi配置保存
		needAuth.POST("/saveConfig", func(c *gin.Context) {
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

			//// 调用 run_auto_bgi.vbs 脚本来启动新的 auto-bgi.exe 程序
			//if err := tools.RestartProgram(); err != nil {
			//	autoLog.Sugar.Error(err.Error())
			//}

			c.JSON(http.StatusOK, gin.H{"status": "success", "message": "配置保存成功"})

		})

		oneLongController := needAuth.Group("/oneLong")
		{
			//启动一条龙
			oneLongController.POST("/startOneLong", func(c *gin.Context) {

				var name string
				if err := c.ShouldBindJSON(&name); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "缺少参数 name"})
					return
				}

				OneLongService.StartOneLong(name)

				c.JSON(http.StatusOK, gin.H{"status": "success", "msg": "启动成功"})
			})

			//查询没有跑完的一条龙配置组
			oneLongController.GET("/unfinishedOneLong", func(context *gin.Context) {
				unfinishedOneLong := bgiStatus.OneLongProgress.Order
				if unfinishedOneLong == nil {
					unfinishedOneLong = []string{}
				}
				context.JSON(http.StatusOK, gin.H{"status": "success", "data": unfinishedOneLong})
			})

		}

		//读取所有一条龙配置
		oneLongController.GET("/oneLongAllName", func(context *gin.Context) {
			oneLongInfo := OneLongService.OneLongAllName()
			context.JSON(http.StatusOK, gin.H{"status": "success", "data": oneLongInfo})
		})

		//读取js的md文件
		needAuth.GET("/md", func(c *gin.Context) {
			filePath := c.Query("filePath")

			jsMd := bgiStatus.ReadMd(filePath)
			c.JSON(http.StatusOK, gin.H{"status": "success", "data": jsMd})

		})

		//批量更新仓库
		needAuth.POST("/batchUpdate", func(c *gin.Context) {
			script := bgiStatus.BatchUpdateScript()
			if script != "" {
				c.JSON(http.StatusOK, gin.H{"status": "success", "message": script})
				return
			}
			c.JSON(http.StatusOK, gin.H{"status": "success", "message": "更新成功"})
		})

		//米游社手动签到
		needAuth.POST("/mysSignIn", func(c *gin.Context) {

			go func() {
				control.CallPython()
			}()

			c.JSON(http.StatusOK, gin.H{"status": "success", "message": "签到成功"})

		})

		//米游社通知配置
		needAuth.POST("/mysPush", Mihoyo.UpdatePushConfig)
		needAuth.GET("/ReadCk", Mihoyo.ReadCK)

		//定时任务
		taskCronController := needAuth.Group("/taskCron")
		{
			taskCronController.GET("/list", TaskCron.List)
			taskCronController.POST("/add", TaskCron.CronAdd)
			taskCronController.POST("/remove", TaskCron.CronRemove)
			//查询可以设置的定时的任务
			taskCronController.GET("/getTasks", TaskCron.GetTask)
			//更新定时任务
			taskCronController.POST("/update", TaskCron.Update)
			//暂停定时任务
			taskCronController.POST("/pause", TaskCron.Pause)
			//恢复定时任务
			taskCronController.POST("/resume", TaskCron.Resume)
			//立即执行任务
			taskCronController.POST("/AtOnceRun", TaskCron.AtOnceRun)
		}

		updateBgi := needAuth.Group("/UpdateBgi")
		{
			updateBgi.POST("/Upload", bgiStatus.UploadBgi)
			updateBgi.POST("/Download", bgiStatus.StartDownloadBgi)
			updateBgi.GET("/Download", bgiStatus.DownloadBgiProgress)
			updateBgi.GET("/DownloadStatus", bgiStatus.GetBgiDownloadStatus)
		}

		aBgiUpdate := needAuth.Group("/aBgiUpdate")
		{

			//获取最新版
			aBgiUpdate.POST("/GetLastVersion", GetLastVersion)
			//更新abgi
			aBgiUpdate.POST("/Update", UpdateABgi)

		}

		//配置组api
		scriptGroup := needAuth.Group("/scriptGroup")
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

			//地图追踪更新，无配置组
			scriptGroup.POST("/UpdatePathingNoConfig", scriptGroupConfig.UpdatePathingNoConfig)

			//查询地图追踪配置
			scriptGroup.GET("/ConfigPathing", func(c *gin.Context) {

				err := scriptGroupConfig.ListPathingUpdatePaths()
				if err != nil {
					return
				}

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

		//查询当前版本
		ginServer.GET("/api/aBgiUpdate/version", GetCurrentVersion)
		// GetBgiVersion 获取当前bgi版本和最新的bgi版本
		ginServer.GET("/api/aBgiUpdate/GetBgiVersion", GetBgiVersion)

		//CD-Aware-AutoGather - 带CD管理的自动采集
		CDAwareAutoGatherController := needAuth.Group("/CD-Aware-AutoGather")
		{
			CDAwareAutoGatherController.GET("/ReadInfo", func(context *gin.Context) {
				//状态
				status := context.Query("status")
				readInfo := CDAwareAutoGatherService.ReadInfo(status)
				context.JSON(http.StatusOK, readInfo)
			})
			CDAwareAutoGatherController.POST("/CDAllMaterial", func(context *gin.Context) {
				material := CDAwareAutoGatherService.CDAllMaterial()
				context.JSON(http.StatusOK, material)
			})

			//更新所有cd管理材料
			CDAwareAutoGatherController.POST("/UpdateAllCD", func(context *gin.Context) {
				CDAwareAutoGatherService.UpdateAllCD()
				context.JSON(http.StatusOK, gin.H{"status": "success", "message": "更新成功,具体请看日志:logs/"})
			})
		}

		CDCollectionController := needAuth.Group("/CDCollectionManagement")
		{
			CDCollectionController.GET("/AllUserFile", CDCollectionManagement.AllUserFile)
			CDCollectionController.GET("/list", CDCollectionManagement.CDCollectionRead)
			CDCollectionController.GET("/ReadPickup", CDCollectionManagement.ReadPickup)

		}

		bgiController := needAuth.Group("/betterGi")
		{
			//读取黑名单
			bgiController.GET("/blackList", func(c *gin.Context) {
				lists, err2 := pickBlackListsService.ReadPickBlackLists()
				if err2 != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "msg": err2.Error()})
					return
				}
				c.JSON(http.StatusOK, gin.H{"status": "success", "data": lists})
			})
			//添加黑名单
			bgiController.POST("/addBlackList", func(c *gin.Context) {
				var blackList []string
				if err := c.ShouldBindJSON(&blackList); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"message": "参数格式错误", "error": err.Error()})
					return
				}

				err := pickBlackListsService.AddPickBlackLists(blackList)
				if err != nil {
					autoLog.Sugar.Infof("添加黑名单失败:%s", err.Error())
					c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "msg": err.Error()})
					return
				}
				c.JSON(http.StatusOK, gin.H{"status": "success", "message": "添加成功"})

			})

			//删除某一个黑名单
			bgiController.POST("/deleteBlackList", func(c *gin.Context) {
				var blackName string
				if err := c.ShouldBindJSON(&blackName); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"message": "参数格式错误", "error": err.Error()})
					return
				}
				err := pickBlackListsService.DeletePickBlackLists(blackName)
				if err != nil {
					autoLog.Sugar.Infof("删除黑名单失败:%s", err.Error())
					c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "msg": err.Error()})
					return
				}

			})

		}

		//obs
		abgiObsController := needAuth.Group("/abgiObs")
		{
			abgiObsController.POST("/StartRecording", func(context *gin.Context) {
				err := abgiObs.StartRecording()
				if err != nil {
					context.JSON(http.StatusInternalServerError, gin.H{"status": "error", "msg": err.Error()})
					return
				}
				context.JSON(http.StatusOK, gin.H{"status": "success", "msg": "开始录制"})
			})
			//结束录制
			abgiObsController.POST("/StopRecording", func(context *gin.Context) {
				videoName := context.Query("videoName")
				err2 := abgiObs.StopRecording(videoName)
				if err2 != nil {
					context.JSON(http.StatusInternalServerError, gin.H{"status": "error", "msg": err2.Error()})
					return
				}
				context.JSON(http.StatusOK, gin.H{"status": "success", "msg": "结束录制"})
			})
			//查询录制状态
			abgiObsController.GET("/IsRecording", func(context *gin.Context) {
				time.Sleep(3 * time.Second)
				isRecording, err3 := abgiObs.GetRecordingStatus()
				if err3 != nil {
					context.JSON(http.StatusInternalServerError, gin.H{"status": "error", "msg": err3.Error()})
					return
				}
				context.JSON(http.StatusOK, gin.H{"status": "success", "msg": isRecording})
			})

			//开启回放缓存
			abgiObsController.POST("/StartReplayBuffer", func(context *gin.Context) {

				err := abgiObs.StartReplayBuffer()
				if err != nil {
					context.JSON(http.StatusInternalServerError, gin.H{"status": "error", "msg": err.Error()})
					return
				}
				context.JSON(http.StatusOK, gin.H{"status": "success", "msg": "开始回放缓冲区"})
			})

			//停止回放缓存
			abgiObsController.POST("/StopReplayBuffer", func(context *gin.Context) {
				err := abgiObs.StopReplayBuffer()
				if err != nil {
					context.JSON(http.StatusInternalServerError, gin.H{"status": "error", "msg": err.Error()})
					return
				}
				context.JSON(http.StatusOK, gin.H{"status": "success", "msg": "停止回放缓冲区"})
			})

			//获取重放缓冲区状态
			abgiObsController.GET("/GetReplayBufferStatus", func(context *gin.Context) {
				time.Sleep(3 * time.Second)
				status, err := abgiObs.GetReplayBufferStatus()
				if err != nil {
					context.JSON(http.StatusInternalServerError, gin.H{"status": "error", "msg": err.Error()})
					return
				}
				context.JSON(http.StatusOK, gin.H{"status": "success", "msg": status})
			})

			//保存重放缓冲区
			abgiObsController.POST("/SaveReplayBuffer", func(context *gin.Context) {

				_, err2 := abgiObs.SaveReplayBuffer("手动保存")
				if err2 != nil {
					context.JSON(http.StatusInternalServerError, gin.H{"status": "error", "msg": err2.Error()})
					return
				}
				context.JSON(http.StatusOK, gin.H{"status": "success", "msg": "保存成功"})
			})

			//获取指定目录下的视频信息列表
			abgiObsController.GET("/GetVideoInfo", func(context *gin.Context) {
				// 获取目录路径参数
				time.Sleep(3 * time.Second)
				info, err := videoInfoService.GetAllRecordingsInfo(config.Cfg.ScreenRecord.ObsSavePath)
				if err != nil {
					context.JSON(http.StatusInternalServerError, gin.H{"status": "error", "msg": err.Error()})
					return
				}
				context.JSON(http.StatusOK, gin.H{"status": "success", "msg": info})
			})

			//删除视频
			abgiObsController.POST("/DeleteVideo", func(context *gin.Context) {
				videoPath := context.Query("videoName") // 视频文件名
				if videoPath == "" {
					context.JSON(400, gin.H{"status": "error", "msg": "缺少视频文件名"})
					return
				}

				err := videoInfoService.DeleteVideo(videoPath)
				if err != nil {
					context.JSON(http.StatusInternalServerError, gin.H{"status": "error", "msg": err.Error()})
					return
				}
				context.JSON(http.StatusOK, gin.H{"status": "success", "msg": "删除成功"})
			})
			//清空所有视频
			abgiObsController.POST("/DeleteAllVideo", videoInfoService.DeleteAllVideo)

			//启动流
			abgiObsController.GET("/StartStream", func(context *gin.Context) {
				err := abgiObs.StartStream()
				if err != nil {
					context.JSON(http.StatusInternalServerError, gin.H{"status": "error", "msg": err.Error()})
					return
				}
				context.JSON(http.StatusOK, gin.H{"status": "success", "msg": "开始流"})
			})
		}

		//查看日志
		needAuth.GET("/autoLog", func(c *gin.Context) {

			data := c.Query("data")

			logs, err2 := autoLog.QueryLogs(data)
			if err2 != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "msg": err2.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"status": "success", "msg": logs})
		})

		ocrController := needAuth.Group("/ocr")
		{
			ocrController.POST("/dogFood", func(c *gin.Context) {

				var data map[string]string
				if err := c.ShouldBindJSON(&data); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"message": "参数格式错误", "error": err.Error()})
				}

				ocr, err2 := Ocr.BaiDuOcr(data["apiKey"], data["secretKey"])
				if err2 != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "msg": err2.Error()})
					return
				}
				dogFood.WriteDogFoodNum(ocr)
				c.JSON(http.StatusOK, gin.H{"status": "success", "data": ocr})

			})
		}

		//js-API
		jsController := needAuth.Group("/js")
		{

			//给指定区域截图
			jsController.POST("/screenShot", JsAPI.ScreenShot)
			//ai
			//jsController.POST("/abgiAiConversation", JsAPI.AbgiAiConversation)
		}
	}

	//测试
	ginServer.GET("/api/test", func(context *gin.Context) {

		oneLongProgress := bgiStatus.OneLongProgress

		context.JSON(http.StatusOK, gin.H{"status": "success", "msg": oneLongProgress})

	})

	//////妙妙屋
	ABGIHoui()

	ginServer.GET("/api/abgiObs/PlayVideoStream", func(c *gin.Context) {
		token := c.Query("tk")
		_, err2 := auth.ParseToken(token)
		if err2 != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "msg": err2.Error()})
			return
		}

		video := config.Cfg.ScreenRecord.ObsSavePath + "/" + c.Query("path")

		if video == "" {
			c.JSON(400, gin.H{"status": "error", "msg": "缺少视频路径"})
			return
		}

		// 检查文件是否存在
		if _, err := os.Stat(video); err != nil {
			c.JSON(404, gin.H{"status": "error", "msg": "视频不存在"})
			return
		}

		// 设置浏览器可以播放视频的 Header
		c.Header("Content-Type", "video/mkv") // 根据视频类型修改，如 .mkv/.flv
		c.Header("Content-Disposition", "inline; filename="+filepath.Base(video))
		c.File(video) // 返回视频文件流
	})

	//桌面图片
	needAuth.GET("/aBgiJt", func(c *gin.Context) {

		//截图
		//err := control.ScreenShot("./img/abgi/jt.jpg")
		//if err != nil {
		//	c.JSON(400, "截图失败")
		//}

		// 1. 获取显示器范围
		bounds := screenshot.GetDisplayBounds(0)

		// 2. 调用 Capture
		img, err := screenshot.Capture(bounds.Min.X, bounds.Min.Y, bounds.Dx(), bounds.Dy())
		if err != nil {
			panic(err)
		}

		// 3. 保存
		f, _ := os.Create("./img/abgi/jt.jpg")
		defer f.Close()
		png.Encode(f, img)

		fmt.Println("截图已保存至./img/abgi/jt.jpg")

		c.File("./img/abgi/jt.jpg") // 指定服务器上的图片路径
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

	type bgiWebhook struct {
		Event      string `json:"event" comment:"事件"`
		Result     int    `json:"result" comment:"结果"`
		Timestamp  string `json:"timestamp" comment:"时间"`
		Screenshot string `json:"screenshot" comment:"图片"`
		Message    string `json:"message" comment:"内容"`
		SendTo     string `json:"send_to" comment:"人员"`
	}

	//webhook
	ginServer.POST("/bgiWebhook", func(c *gin.Context) {
		var payload bgiWebhook

		if err := c.ShouldBindJSON(&payload); err != nil {
			fmt.Println(err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"message": "参数格式错误", "error": err.Error()})
			return
		}

		//fmt.Println(payload.Event)
		//fmt.Println(payload.Result)
		//fmt.Println(payload.Timestamp)
		//fmt.Println(payload.Screenshot)
		fmt.Println(payload.Message)
		fmt.Println(payload.SendTo)

		iconPath, _ := filepath.Abs("./img/ff.png") // 转为绝对路径

		notification := toast.Notification{
			AppID:   "autoBgi",
			Title:   payload.Event,
			Message: payload.Message,
			Icon:    iconPath, // 可选
		}
		err := notification.Push()
		if err != nil {
			autoLog.Sugar.Errorf("推送失败: %v", err)
		}

		//if payload.SendTo == "" {
		//	payload.SendTo = "老王"
		//}
		//
		//speech := htgotts.Speech{Folder: "audio", Language: voices.Chinese, Handler: &handlers.Native{}}
		//speech.Speak(payload.SendTo + "," + payload.Message)
		//
		//c.JSON(http.StatusOK, gin.H{"status": "success", "msg": payload})

	})

	ginServer.POST("/bot/bgiWebhook", func(c *gin.Context) {
		var data map[string]interface{}

		if err := c.ShouldBindJSON(&data); err != nil {
			fmt.Println(err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"message": "参数格式错误", "error": err.Error()})
			return
		}

		fmt.Println(data)
		c.JSON(http.StatusOK, gin.H{"status": "success", "msg": data})

	})

	// 1. 静态资源挂载（直接让前端可以访问图片）
	ginServer.Static("/img", "./img")

	imageListOnce.Do(loadImages) // 只加载一次

	// 2. API：返回所有图片的 URL
	ginServer.GET("/api/images", func(c *gin.Context) {
		c.Header("Cache-Control", "public, max-age=86400")
		c.Header("Expires", time.Now().AddDate(0, 0, 3).Format(http.TimeFormat))
		c.JSON(200, gin.H{"images": imageList})

	})

	//if config.Cfg.Control.AbgiScreen {
	//	autoLog.Sugar.Infof("屏幕捕获开启状态")
	//	// 设置最大CPU使用
	//	runtime.GOMAXPROCS(runtime.NumCPU())
	//
	//	// 启动屏幕捕获协程
	//	go abgiScreen.CaptureScreen()
	//
	//	// 启动广播协程
	//	go abgiScreen.BroadcastFrames()
	//
	//	abgiScreenController := ginServer.Group("/api/abgiScreen")
	//	{
	//		abgiScreenController.GET("/ws", abgiScreen.HandleWebSocket)
	//
	//	}
	//
	//} else {
	//	autoLog.Sugar.Infof("屏幕捕获关闭状态")
	//}

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
			c.Data(http.StatusOK, contentType, iconData)
			return
		}

		c.Data(http.StatusOK, contentType, content)
	})

	if len(os.Args) > 1 {
		if os.Args[1] == "OneLong" { // 3. 关闭软件（同步，后续任务依赖此步骤）
			control.CloseSoftware()
			autoLog.Sugar.Info("软件已关闭")
			OneLongService.OneLongTask(os.Args[2])
			autoLog.Sugar.Infof("一条龙启动:%s", os.Args[2])
		} else if os.Args[1] == "updateJs" {
			if err := bgiStatus.BatchUpdateScript(); err != "" {
				autoLog.Sugar.Errorf("批量更新脚本失败: %v", err)
			} else {
				autoLog.Sugar.Infof("批量更新脚本成功")
			}

		} else if os.Args[1] == "online" {
			//上线

			if config.Cfg.Account.Uid == "" {
				fmt.Println("账号配置错误")
				return
			}
			if config.Cfg.Account.Name == "" {
				fmt.Println("账号名称配置错误")
				return
			}
			if config.Cfg.Account.SecretKey == "" {
				fmt.Println("密钥配置错误")
				return
			}
			//解密
			decryptedKey, err3 := tools.Decrypt(config.Cfg.Account.SecretKey)
			if err3 != nil {
				fmt.Println("密钥错误")
				return
			}

			err := abgiSSE.Connect(fmt.Sprintf("ws://%s/api/abgiWs/%s/%s/%s", decryptedKey, "debug", config.Cfg.Account.Uid, config.Cfg.Account.Name), true, nil)
			if err != nil {
				fmt.Printf("连接失败: %v\n", err)
				return

			}
		}
	}

	//服务器端口
	post := config.Cfg.Post

	if post == "" || post == ":" {
		post = ":8082"
	}

	//判断端口是否被占用
	post = getAvailablePort(post)

	ABGIHouiPort(post)

	err = ginServer.Run(post)

	if err != nil {
		autoLog.Sugar.Errorf("服务器启动失败:%v", err)
		//退出程序
		os.Exit(1)
		return
	}

	//err = ginServer.RunTLS(post, "certFile/cert.pem", "certFile/key.pem")
	//if err != nil {
	//	autoLog.Sugar.Errorf("启动失败:%v", err)
	//}
}

// 检查端口是否可用，不行就+1
// getAvailablePort 函数用于获取一个可用的端口号
// 参数 start 是起始端口号的字符串形式
// 返回值是一个字符串形式的可用端口号
func getAvailablePort(start string) string {

	// 移除字符串中的冒号，只保留数字部分
	start = strings.ReplaceAll(start, ":", "")

	//转成数字
	port, _ := strconv.Atoi(start)

	for {
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err == nil {
			_ = ln.Close() // 关闭临时占用
			autoLog.Sugar.Infof("端口：【%d】可以使用", port)
			config.Cfg.Post = fmt.Sprintf(":%d", port)
			userIP = "http://localhost" + config.Cfg.Post

			ips, err := tools.GetLocalIPs()
			if err != nil {
				autoLog.Sugar.Infof("获取本机IP失败: %v", err)
			} else {
				autoLog.Sugar.Infof("浏览器使用本机局域网IP")
				for _, ip := range ips {
					if strings.Contains(ip, "192.168") {
						autoLog.Sugar.Infof("本机局域网IP: %s%s", ip, config.Cfg.Post)
						userIP = fmt.Sprintf("http://%s%s", ip, config.Cfg.Post)
					} else {
						autoLog.Sugar.Infof("本机其他IP: %s%s", ip, config.Cfg.Post)
					}

				}
			}

			return fmt.Sprintf(":%d", port)
		}
		autoLog.Sugar.Infof("端口：【%d】被占用", port)
		port++
	}
}
