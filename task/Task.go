package task

import (
	"auto-bgi/Notice"
	"auto-bgi/abgiSSE"
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"auto-bgi/control"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
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

	//暂停500毫秒
	time.Sleep(500 * time.Millisecond)

	betterGIPath := filepath.Join(config.Cfg.BetterGIAddress, "BetterGI.exe")

	// 检查文件是否存在
	if _, err := os.Stat(betterGIPath); err != nil {
		autoLog.Sugar.Errorf("BetterGI.exe 不存在: %v", err)
		return err
	}

	args := append([]string{"--startGroups"}, names...) // 每个组名单独参数

	maxRetries := 3
	var err error

	for i := 0; i < maxRetries; i++ {
		cmd := exec.Command(betterGIPath, args...)
		cmd.Dir = config.Cfg.BetterGIAddress // 设置工作目录，确保能读取到配置文件
		cmd.Stdout = nil
		cmd.Stderr = nil

		err = cmd.Start()
		if err != nil {
			autoLog.Sugar.Errorf("启动命令执行失败 (第 %d 次尝试): %v", i+1, err)
		} else {
			autoLog.Sugar.Infof("启动命令已下发，等待 20 秒检查进程状态 (第 %d 次尝试)...", i+1)
		}

		// 等待20秒检查进程
		time.Sleep(20 * time.Second)

		if control.CheckProcessRunning("BetterGI.exe") {
			autoLog.Sugar.Infof("检测到 BetterGI 进程运行正常，启动成功: %v", names)
			return nil
		}

		autoLog.Sugar.Warnf("未检测到 BetterGI 进程 (第 %d 次尝试)，将重试...", i+1)
	}

	return fmt.Errorf("启动配置组失败，已重试 %d 次: %w", maxRetries, err)
}

func StartOneDragon(name string) {

	autoLog.Sugar.Infof("准备启动一条龙：%s", name)

	// 延迟确保关闭完成
	delay := 2 * time.Second
	autoLog.Sugar.Infof("等待 %v 后启动...", delay)
	time.Sleep(delay)

	betterGIPath := filepath.Join(config.Cfg.BetterGIAddress, "BetterGI.exe")
	if _, err := os.Stat(betterGIPath); err != nil {
		autoLog.Sugar.Errorf("BetterGI.exe 不存在: %v", err)
		return
	}

	maxRetries := 3
	var err error

	for i := 0; i < maxRetries; i++ {
		cmd := exec.Command(betterGIPath, "--startOneDragon", name)
		cmd.Dir = config.Cfg.BetterGIAddress
		err = cmd.Start()
		if err != nil {
			autoLog.Sugar.Errorf("启动一条龙命令执行失败 (第 %d 次尝试): %v", i+1, err)
		} else {
			autoLog.Sugar.Infof("启动一条龙命令已下发，等待 20 秒检查进程状态 (第 %d 次尝试)...", i+1)
		}

		time.Sleep(20 * time.Second)

		if control.CheckProcessRunning("BetterGI.exe") {
			autoLog.Sugar.Infof("检测到 BetterGI 进程运行正常，一条龙启动成功: %s", name)
			return
		}
		autoLog.Sugar.Warnf("未检测到 BetterGI 进程 (第 %d 次尝试)，将重试...", i+1)
	}

	autoLog.Sugar.Errorf("启动一条龙失败，已重试 %d 次: %v", maxRetries, err)
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

// 发送摩拉排行榜
func SendMoraleRankTask() {
	cronTab := cron.New(cron.WithSeconds())

	// 定时任务,cron表达式
	//每1个小时执行一次
	//spec := fmt.Sprintf("0 0 0,3,6,9,12,15,18,21 * * ? ")
	spec := fmt.Sprintf("0 12 9 * * ? ")
	// 定义定时器调用的任务函数
	task := func() {
		now := time.Now().Format("2006-01-02 15:04:05")
		autoLog.Sugar.Infof("定时任务启动-摩拉排行榜 %v", now)
		rank := Rank()
		autoLog.Sugar.Infof("摩拉记录更新结果： %v", rank)

	}

	// 添加定时任务
	cronTab.AddFunc(spec, task)
	// 启动定时器
	cronTab.Start()
	// 阻塞主线程停止
	select {}

}

// 清空联机次数
func ClearRunCount() {

	cronTab := cron.New(cron.WithSeconds())

	// 定时任务,cron表达式
	//每1个小时执行一次
	spec := fmt.Sprintf("0 30 23 * * *")

	// 定义定时器调用的任务函数
	task := func() {
		autoLog.Sugar.Infof("清空联机次数 %v", time.Now().Format("2006-01-02 15:04:05"))
		abgiSSE.RunCount = 0
		Rank()
	}

	// 添加定时任务
	cronTab.AddFunc(spec, task)
	// 启动定时器
	cronTab.Start()
	// 阻塞主线程停止
	select {}

}

func StartOneDragonPlan(PlanName string) {
	autoLog.Sugar.Infof("准备启动计划表：%s", PlanName)

	// 延迟确保关闭完成
	delay := 2 * time.Second
	autoLog.Sugar.Infof("等待 %v 后启动...", delay)
	time.Sleep(delay)

	betterGIPath := filepath.Join(config.Cfg.BetterGIAddress, "BetterGI.exe")
	if _, err := os.Stat(betterGIPath); err != nil {
		autoLog.Sugar.Errorf("BetterGI.exe 不存在: %v", err)
		return
	}

	maxRetries := 3
	var err error

	for i := 0; i < maxRetries; i++ {
		cmd := exec.Command(betterGIPath, "--startContinuousOneDragon", PlanName)
		cmd.Dir = config.Cfg.BetterGIAddress
		err = cmd.Start()
		if err != nil {
			autoLog.Sugar.Errorf("启动计划表命令执行失败 (第 %d 次尝试): %v", i+1, err)
		} else {
			autoLog.Sugar.Infof("启动计划表命令已下发，等待 20 秒检查进程状态 (第 %d 次尝试)...", i+1)
		}

		time.Sleep(20 * time.Second)

		if control.CheckProcessRunning("BetterGI.exe") {
			autoLog.Sugar.Infof("检测到 BetterGI 进程运行正常，计划表启动成功: %s", PlanName)
			return
		}
		autoLog.Sugar.Warnf("未检测到 BetterGI 进程 (第 %d 次尝试)，将重试...", i+1)
	}

	autoLog.Sugar.Errorf("启动计划表失败，已重试 %d 次: %v", maxRetries, err)
}
