package control

import "C"
import (
	"auto-bgi/abgiConstant"
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"crypto/sha256"
	"fmt"
	"github.com/gin-gonic/gin"
	"image/jpeg"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/go-vgo/robotgo"
	"github.com/tidwall/gjson"
)

/*
#cgo LDFLAGS: -luser32 -lgdi32
#include <windows.h>

// 定义一个结构体方便返回
typedef struct {
    int width;
    int height;
} ScreenSize;

// 实现获取分辨率的函数
static ScreenSize GetPhysicalResolution() {
    // 关键：声明进程对高DPI感知
    SetProcessDPIAware();

    HDC hdc = GetDC(NULL);
    ScreenSize size;

    // 获取真实的物理像素宽度和高度
    size.width = GetDeviceCaps(hdc, 8);  // 8 是 HORZRES
    size.height = GetDeviceCaps(hdc, 10); // 10 是 VERTRES

    ReleaseDC(NULL, hdc);
    return size;
}

*/
import "C"

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
		autoLog.Sugar.Errorf("无法打开软件: %v", err)
	} else {
		fmt.Println("打开成功")
		autoLog.Sugar.Infof("打开成功")
	}

}

func GetSysTemUser() string {
	// 优先使用标准库获取
	u, err := user.Current()
	if err == nil {
		// 在 Windows 下，u.Username 会返回 "DOMAIN\User" 格式
		return u.Username
	}

	// 备选方案：环境变量拼接
	domain := os.Getenv("USERDOMAIN")
	user := os.Getenv("USERNAME")
	if domain != "" && user != "" {
		return domain + "\\" + user
	}

	return user // 实在拿不到就只返用户名
}

// 关闭软件
func CloseSoftware() {
	//关闭软件之前，停止当前脚本/独立任务
	CancelTaskHotkey()
	time.Sleep(5 * time.Second)

	username := GetSysTemUser()
	autoLog.Sugar.Infof("当前window用户是：%s", username)

	// 构建正确的命令
	cmd := exec.Command("cmd", "/c", "taskkill", "/F", "/IM", "BetterGI.exe", "/FI", fmt.Sprintf("USERNAME eq %s", username))

	// 创建命令
	//cmd := exec.Command("cmd", "/c", "taskkill", "/F", "/IM", "BetterGI.exe", "/FI", "USERNAME eq %USERNAME%")

	//cmd := exec.Command("taskkill", "/F", "/IM", "BetterGI.exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Stdout = nil
	cmd.Stderr = nil

	// 执行命令并获取输出
	_, err := cmd.CombinedOutput()

	if err != nil {
		autoLog.Sugar.Errorf("执行命令出错: %v\n", err)
	} else {
		autoLog.Sugar.Infof("BetterGI关闭成功")
	}

}

// CheckProcessRunning 检查指定进程名的进程是否在运行
func CheckProcessRunning(processName string) bool {
	cmd := exec.Command("tasklist", "/FI", fmt.Sprintf("IMAGENAME eq %s", processName))
	// 隐藏窗口
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := cmd.CombinedOutput()
	if err != nil {
		autoLog.Sugar.Errorf("检查进程失败: %v", err)
		return false
	}
	return strings.Contains(string(output), processName)
}

// CloseYuanShen 关闭软件
// CloseYuanShen 函数用于关闭原神游戏及其启动器
var CloseYuanShenWg sync.WaitGroup

func CloseYuanShen() {
	CloseYuanShenWg.Add(3)

	username := GetSysTemUser()
	autoLog.Sugar.Infof("当前window用户是：%s", username)

	go func() {
		defer CloseYuanShenWg.Done()
		// 创建命令用于强制关闭原神进程
		//cmd := exec.Command("taskkill", "/F", "/IM", "YuanShen.exe")
		cmd := exec.Command("cmd", "/c", "taskkill", "/F", "/IM", "YuanShen.exe", "/FI", fmt.Sprintf("USERNAME eq %s", username))
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

		// 执行命令并获取输出
		_, err := cmd.CombinedOutput()

		// 判断原神关闭结果
		if err != nil {
			autoLog.Sugar.Infof("原神已关闭")
		} else {
			autoLog.Sugar.Infof("原神关闭成功")
		}
	}()

	go func() {
		defer CloseYuanShenWg.Done()
		// 创建命令用于强制关闭原神启动器进程
		//cmd3 := exec.Command("taskkill", "/F", "/IM", "GenshinImpact.exe")
		cmd3 := exec.Command("cmd", "/c", "taskkill", "/F", "/IM", "GenshinImpact.exe", "/FI", fmt.Sprintf("USERNAME eq %s", username))
		// 执行命令并获取输出
		_, err3 := cmd3.CombinedOutput()

		// 判断启动器关闭结果
		if err3 != nil {
			autoLog.Sugar.Infof("原神国际已关闭")
		} else {
			autoLog.Sugar.Infof("原神国际关闭成功")
		}
	}()

	go func() {
		defer CloseYuanShenWg.Done()
		// 创建命令用于强制关闭原神启动器进程
		//cmd2 := exec.Command("taskkill", "/F", "/IM", "HYP.exe")
		cmd2 := exec.Command("cmd", "/c", "taskkill", "/F", "/IM", "HYP.exe", "/FI", fmt.Sprintf("USERNAME eq %s", username))

		// 执行命令并获取输出
		_, err2 := cmd2.CombinedOutput()

		// 判断启动器关闭结果
		if err2 != nil {
			autoLog.Sugar.Infof("原神启动器已关闭")
		} else {
			autoLog.Sugar.Infof("原神启动器关闭成功")
		}
	}()

	CloseYuanShenWg.Wait()

	// 等待3秒，确保原神完全关闭
	time.Sleep(2000 * time.Millisecond)

}

// 鼠标连续操作: type:类型 x:坐标  y   key:  DoubleClick time
func MouseAndKeyboardClicks(dataS string) {
	split := strings.Split(dataS, "*")
	for _, v := range split {
		data := strings.Split(v, "-")
		if data[0] == "Mouse" {
			autoLog.Sugar.Infof("电脑鼠标键盘操作---鼠标指令:%s", data)
			x, _ := strconv.Atoi(data[1])
			y, _ := strconv.Atoi(data[2])
			key := data[3]
			DoubleClick, _ := strconv.ParseBool(data[4])
			MouseClick(x, y, key, DoubleClick)
			d, _ := strconv.Atoi(data[5])
			time.Sleep(time.Duration(d) * time.Millisecond)
		} else if data[0] == "Keyboard" {
			autoLog.Sugar.Infof("电脑鼠标键盘操作---键盘指令:%s", data)
			pressKey(data[1])
			d, _ := strconv.Atoi(data[2])
			time.Sleep(time.Duration(d) * time.Millisecond)
		} else {
			autoLog.Sugar.Infof("电脑鼠标键盘操作---未知指令:%s", data)
		}

	}
}

func MouseClick(x, y int, key string, DoubleClick bool) {

	// 移动鼠标到指定位置
	robotgo.Move(x, y, 300)

	// 等待一小会儿，确保鼠标移动完成
	// 使用500毫秒的延迟，给系统足够时间完成鼠标移动
	time.Sleep(500 * time.Millisecond)

	// 模拟鼠标左键点击
	robotgo.Click(key, DoubleClick) // 第二个参数为 true 表示双击，false 表示单击
}

// ScreenShot 截图
func ScreenShot(imgName string) error {
	//screenWidth, screenHeight := robotgo.GetScreenSize()

	size := C.GetPhysicalResolution()
	//fmt.Println("--- 屏幕真实分辨率 ---")
	//fmt.Printf("宽度: %d px\n", int(size.width))
	//fmt.Printf("高度: %d px\n", int(size.height))

	imgScreen := robotgo.CaptureScreen(0, 0, int(size.width), int(size.height))
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

func mapHotkey(key string) string {
	switch key {
	// --- 重点补充：小键盘 (Robotgo 专用映射) ---
	case "NumPad0":
		return "numpad_0"
	case "NumPad1":
		return "numpad_1"
	case "NumPad2":
		return "numpad_2"
	case "NumPad3":
		return "numpad_3"
	case "NumPad4":
		return "numpad_4"
	case "NumPad5":
		return "numpad_5"
	case "NumPad6":
		return "numpad_6"
	case "NumPad7":
		return "numpad_7"
	case "NumPad8":
		return "numpad_8"
	case "NumPad9":
		return "numpad_9"
	case "Decimal":
		return "numpad_dot"
	case "Add":
		return "+"
	case "Subtract":
		return "-"
	case "Multiply":
		return "*"
	case "Divide":
		return "/"

	// --- OEM 符号键 ---
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
	case "Oem3":
		return "`"

	// --- 顶部数字键 ---
	case "D1":
		return "1"
	case "D2":
		return "2"
	case "D3":
		return "3"
	case "D4":
		return "4"
	case "D5":
		return "5"
	case "D6":
		return "6"
	case "D7":
		return "7"
	case "D8":
		return "8"
	case "D9":
		return "9"
	case "D0":
		return "0"

	// --- 特殊控制键 ---
	case "Return":
		return "enter"
	case "Back":
		return "backspace"
	case "Space":
		return "space"
	case "Tab":
		return "tab"
	case "Capital":
		return "capslock"

	default:
		return strings.ToLower(key)
	}
}

func pressKey(key string) {
	hotkey := mapHotkey(key)
	err := robotgo.KeyTap(hotkey)
	if err != nil {
		autoLog.Sugar.Errorf("按键失败: %v", err)
	}
}

func PressKeyWeb(context *gin.Context) {
	value := context.Query("key")
	PressKey(value)
	autoLog.Sugar.Infof("PressKeyWeb按键成功: %s", value)
	context.JSON(200, "按键成功")
}

func PressKey(key string) {
	hotkey := mapHotkey(key)
	err := robotgo.KeyTap(hotkey)
	if err != nil {
		autoLog.Sugar.Errorf("按键失败: %v", err)
	}
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

// 调用Python
func CallPython() {
	defer func() {
		if err := recover(); err != nil {
			autoLog.Sugar.Errorf("CallPython捕获异常: %v", err)
		}
	}()

	cmd := exec.Command("MihoyoBBSTools\\venv\\python.exe", "MihoyoBBSTools\\main.py")

	// --- 关键代码：隐藏黑窗口 ---
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,       // 隐藏窗口
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW 标志
	}
	// --------------------------

	stdout, _ := cmd.StdoutPipe()
	cmd.Start()

	buf := make([]byte, 1024)
	for {
		n, err := stdout.Read(buf)
		if n == 0 || err != nil {
			break
		}
		autoLog.Sugar.Infof(string(buf[:n]))
	}

	cmd.Wait()
}

// 执行系统命令并获取输出
func getWmic(target string, property string) string {
	cmd := exec.Command("wmic", target, "get", property)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	lines := strings.Split(string(out), "\n")
	if len(lines) > 1 {
		return strings.TrimSpace(lines[1])
	}
	return ""
}

func GetMachineFingerprint(Type, version string) {
	//捕获异常
	defer func() {
		if err := recover(); err != nil {
			autoLog.Sugar.Errorf("GetMachineFingerprint捕获异常: %v", err)
		}
	}()

	// 1. 获取主板 UUID
	uuid := getWmic("csproduct", "uuid")
	// 2. 获取 CPU ID
	cpuID := getWmic("cpu", "processorid")
	// 3. 获取硬盘序列号 (首个物理硬盘)
	diskID := getWmic("diskdrive", "serialnumber")

	// 组合原始字符串
	rawIdentifier := fmt.Sprintf("UUID:%s|CPU:%s|DISK:%s", uuid, cpuID, diskID)

	// SHA256 哈希处理
	hash := sha256.Sum256([]byte(rawIdentifier))

	UpdateLog := make(map[string]interface{})
	UpdateLog["machine_fingerprint"] = fmt.Sprintf("%x", hash)
	UpdateLog["Type"] = Type
	UpdateLog["version"] = version

	jsonResp, jsonStatus, jsonErr := abgiConstant.PostJSON(abgiConstant.ABgiInfoUrl+"/api/UpdateLog/Add", UpdateLog, nil)
	if jsonErr != nil {
		fmt.Printf("JSON 请求失败: %v\n", jsonErr)
	} else {
		fmt.Printf("JSON 请求状态码: %d\n", jsonStatus)
		fmt.Printf("JSON 响应内容: %s\n\n", jsonResp)

	}
	autoLog.Sugar.Infof("更新记录成功")

}
