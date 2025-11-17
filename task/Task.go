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
	"fmt"
	"github.com/robfig/cron/v3"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

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
	cmdArgs := append([]string{"/C", "start", "", betterGIPath}, args...)

	exec.Command("cmd", cmdArgs...).Start()

	autoLog.Sugar.Infof("执行命令:cmd /C start %s %s", betterGIPath, fmt.Sprintf("--startGroups %s", strings.Join(names, " ")))

	return nil
}

func StartOneDragon(name string) {
	autoLog.Sugar.Infof("准备启动一条龙：%s", name)

	// 关闭软件
	control.CloseSoftware()

	// 延迟确保关闭完成
	delay := 3 * time.Second
	autoLog.Sugar.Infof("等待 %v 后启动...", delay)
	time.Sleep(delay)

	betterGIPath := filepath.Join(config.Cfg.BetterGIAddress, "BetterGI.exe")
	if _, err := os.Stat(betterGIPath); err != nil {
		autoLog.Sugar.Errorf("BetterGI.exe 不存在: %v", err)
		return
	}
	exec.Command("cmd", "/C", "start", "", betterGIPath, "--startOneDragon", name).Start()

	autoLog.Sugar.Infof("执行命令：cmd /C start   %s %s %s", betterGIPath, "--startOneDragon", name)

}

// 调用bat脚本
func CallBat(batPath string) {
	res := exec.Command("cmd", "/C", "", batPath)
	out, err := res.CombinedOutput()
	if err != nil {
		autoLog.Sugar.Errorf("执行bat脚本失败：%v", err)
		return
	}
	autoLog.Sugar.Infof("执行bat脚本成功：%s", string(out))

	autoLog.Sugar.Infof("执行命令：cmd /C start   %s", batPath)

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
