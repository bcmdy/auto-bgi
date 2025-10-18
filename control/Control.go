package control

import (
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"fmt"
	"github.com/go-vgo/robotgo"
	"github.com/tidwall/gjson"
	"image/jpeg"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
	"unsafe"
)

var (
	user32                   = syscall.NewLazyDLL("user32.dll")
	procFindWindow           = user32.NewProc("FindWindowW")
	procSetForegroundWnd     = user32.NewProc("SetForegroundWindow")
	procGetForeground        = user32.NewProc("GetForegroundWindow")
	procGetWindowTextW       = user32.NewProc("GetWindowTextW")
	procGetWindowTextLengthW = user32.NewProc("GetWindowTextLengthW")
)

// 打开软件
func OpenSoftware(programPath string) {

	cmd := exec.Command("cmd", "/C", "start", "", programPath)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	err := cmd.Start()
	if err != nil {
		fmt.Println("无法打开软件:", err)
	}
	fmt.Println("打开成功")

}

// 关闭软件
func CloseSoftware() {
	//关闭软件之前，停止当前脚本/独立任务
	CancelTaskHotkey()
	time.Sleep(5 * time.Second)

	// 创建命令
	cmd := exec.Command("taskkill", "/F", "/IM", "BetterGI.exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	// 执行命令并获取输出
	output, err := cmd.CombinedOutput()

	if err != nil {
		autoLog.Sugar.Errorf("执行命令出错: %v\n", err)
	}
	autoLog.Sugar.Infof("命令输出: %s", string(output))

}

// CloseYuanShen 关闭软件
// CloseYuanShen 函数用于关闭原神游戏及其启动器
func CloseYuanShen() {

	// 检查配置文件中是否设置了需要关闭原神
	if !config.Cfg.Control.IsCloseYuanShen {
		autoLog.Sugar.Infof("不需要关闭原神")
		return
	}
	autoLog.Sugar.Infof("需要关闭原神")
	// 创建命令用于强制关闭原神进程
	cmd := exec.Command("taskkill", "/F", "/IM", "YuanShen.exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	// 执行命令并获取输出
	_, err := cmd.CombinedOutput()

	// 判断原神关闭结果
	if err != nil {
		autoLog.Sugar.Infof("原神已关闭")
	} else {
		autoLog.Sugar.Infof("原神关闭成功")
	}

	// 等待3秒，确保原神完全关闭
	time.Sleep(3000 * time.Millisecond)

	// 创建命令用于强制关闭原神启动器进程
	cmd2 := exec.Command("taskkill", "/F", "/IM", "HYP.exe")

	// 执行命令并获取输出
	_, err2 := cmd2.CombinedOutput()

	// 判断启动器关闭结果
	if err2 != nil {
		autoLog.Sugar.Infof("原神启动器已关闭")
	} else {
		autoLog.Sugar.Infof("原神启动器关闭成功")
	}

}

// MouseClick 鼠标点击(x、y是鼠标坐标，key是键，是否双击)
func MouseClick(x, y int, key string, DoubleClick bool) {

	// 移动鼠标到指定位置
	robotgo.Move(x, y)

	// 等待一小会儿，确保鼠标移动完成
	time.Sleep(500 * time.Millisecond)

	// 模拟鼠标左键点击
	robotgo.Click(key, DoubleClick) // 第二个参数为 true 表示双击，false 表示单击
}

// ScreenShot 截图
func ScreenShot(imgName string) error {
	screenWidth, screenHeight := robotgo.GetScreenSize()
	imgScreen := robotgo.CaptureScreen(0, 0, screenWidth, screenHeight)
	if imgScreen == nil {
		return fmt.Errorf("截图失败: 无法获取屏幕图像")
	}
	defer robotgo.FreeBitmap(imgScreen)

	img := robotgo.ToImage(imgScreen)

	// 🔹确保目录存在
	dir := filepath.Dir(imgName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}

	file, err := os.Create(imgName)
	if err != nil {
		return err
	}
	defer file.Close()

	// 这里的 Quality 可以调节（1-100），一般 70 就够了
	opt := &jpeg.Options{Quality: 70}
	return jpeg.Encode(file, img, opt)
}

func findWindow(className, windowName *uint16) (hwnd uintptr, err error) {
	ret, _, err := procFindWindow.Call(
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(windowName)),
	)
	if ret == 0 {
		return 0, err
	}
	return ret, nil
}

func setForegroundWindow(hwnd uintptr) bool {
	ret, _, _ := procSetForegroundWnd.Call(hwnd)
	return ret != 0
}

// 切换屏幕
func SwitchingScreens(name string) {
	windowTitle, _ := syscall.UTF16PtrFromString(name)

	hwnd, err := findWindow(nil, windowTitle)
	if err != nil || hwnd == 0 {

		autoLog.Sugar.Infof("找不到指定窗口:", err)
		return
	}

	success := setForegroundWindow(hwnd)
	if success {

		autoLog.Sugar.Infof("成功切换到窗口: %s", name)
	} else {
		autoLog.Sugar.Errorf("切换窗口失败")
	}
}

func getForegroundWindow() (hwnd uintptr) {
	ret, _, _ := procGetForeground.Call()
	return ret
}

func getWindowText(hwnd uintptr) string {
	length, _, _ := procGetWindowTextLengthW.Call(hwnd)
	if length == 0 {
		return ""
	}

	buf := make([]uint16, length+1)
	ret, _, _ := procGetWindowTextW.Call(
		hwnd,
		uintptr(unsafe.Pointer(&buf[0])),
		length+1,
	)
	if ret == 0 {
		return ""
	}

	// 转换UTF-16到Go字符串
	return syscall.UTF16ToString(buf)
}

// 获取当前窗口标题
func GetWindows() string {
	hwnd := getForegroundWindow()
	if hwnd == 0 {
		autoLog.Sugar.Infof("无法获取活动窗口句柄")
		return ""
	}

	title := getWindowText(hwnd)
	if title == "" {
		autoLog.Sugar.Infof("无法获取窗口标题或窗口标题为空")
		return ""
	}

	autoLog.Sugar.Infof("当前活动窗口标题: %s", title)
	return title
}

func HttpGet(url string) error {
	// 目标 URL
	//url := "http://localhost:8888"
	// 发送 GET 请求
	resp, err := http.Get(url)
	if err != nil {
		autoLog.Sugar.Infof("发送请求失败: %v", err)
		return err
	}
	defer resp.Body.Close() // 延迟关闭响应体
	// 检查响应状态码
	autoLog.Sugar.Infof("状态码: %d", resp.StatusCode)
	if resp.StatusCode == 200 {
		return nil
	}
	autoLog.Sugar.Errorf("状态码: %d", resp.StatusCode)
	return fmt.Errorf("状态码: %d", resp.StatusCode)
}

func StartRecord() {
	//点击F12开始录屏
	err := robotgo.KeyTap("f12")
	if err != nil {
		autoLog.Sugar.Errorf("开始录屏失败: %v", err)
	}

}

func StopRecord() {
	//点击F12结束录屏
	err := robotgo.KeyTap("f12")
	if err != nil {
		autoLog.Sugar.Errorf("结束录屏失败: %v", err)
	}
}

// 按键
func mapHotkey(key string) string {
	switch key {
	case "OemOpenBrackets":
		return "["
	case "OemCloseBrackets":
		return "]"
	case "OemComma":
		return ","
	case "OemPeriod":
		return "."
	case "OemMinus":
		return "-"
	case "OemPlus":
		return "="
	case "Oem1":
		return ";"
	case "OemQuotes":
		return "'"
	case "Oem5":
		return "\\"
	case "OemQuestion":
		return "/"
	default:
		return key
	}
}

func pressKey(key string) {
	hotkey := mapHotkey(key)
	robotgo.KeyTap(hotkey)
}

// CancelTaskHotkey 停止当前脚本/独立任务
func CancelTaskHotkey() {
	// 构建配置文件路径，从配置中获取基础路径并拼接配置文件名
	configPath := config.Cfg.BetterGIAddress + "\\User\\Config.json"
	// 读取配置文件内容，忽略可能出现的错误
	configData, _ := os.ReadFile(configPath)

	// 从配置文件中获取取消任务的快捷键值
	value := gjson.Get(string(configData), "hotKeyConfig.cancelTaskHotkey")
	// 检查快捷键值是否存在，如果存在则执行按键操作
	if value.Exists() {
		autoLog.Sugar.Infof("取消任务快捷键: %s", value.String())
		pressKey(value.String())
	}
}

// 按ESC
func PressEsc() {
	autoLog.Sugar.Infof("按下ESC")
	pressKey("esc")
}

// 读取原神exe文件路径
func ReadGenShinExe() string {
	configPath := config.Cfg.BetterGIAddress + "\\User\\Config.json"

	configData, _ := os.ReadFile(configPath)

	value := gjson.Get(string(configData), "genshinStartConfig.installPath")

	autoLog.Sugar.Infof("原神exe文件路径: %s", value.String())

	return value.String()

}
