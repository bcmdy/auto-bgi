package control

import (
	"auto-bgi/autoLog"
	"fmt"
	"github.com/go-vgo/robotgo"
	"github.com/pterm/pterm"
	"github.com/vcaesar/imgo"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
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
	// 检查当前操作系统

	// Windows 使用 "start" 命令
	cmd := exec.Command("cmd", "/C", "start", "", programPath)
	err := cmd.Start()
	if err != nil {
		fmt.Println("无法打开软件:", err)
	}
	fmt.Println("打开成功")

}

// 关闭软件
func CloseSoftware() {
	// 创建命令
	cmd := exec.Command("taskkill", "/F", "/IM", "BetterGI.exe")

	// 执行命令并获取输出
	output, err := cmd.CombinedOutput()

	if err != nil {
		//fmt.Printf("执行命令出错: %v\n", err)
		pterm.DefaultBasicText.Println("执行命令出错:", err)
	}
	pterm.DefaultBasicText.Println("命令输出:", string(output))
	//fmt.Printf("命令输出:\n%s\n", string(output))

}

// 打开软件
func CloseYuanShen() {
	// 创建命令
	cmd := exec.Command("taskkill", "/F", "/IM", "YuanShen.exe")

	// 执行命令并获取输出
	_, err := cmd.CombinedOutput()

	if err != nil {
		autoLog.Sugar.Infof("原神已关闭")
	} else {
		autoLog.Sugar.Infof("原神关闭成功")
	}

	time.Sleep(3000 * time.Millisecond)

	// 创建命令
	cmd2 := exec.Command("taskkill", "/F", "/IM", "HYP.exe")

	// 执行命令并获取输出
	_, err2 := cmd2.CombinedOutput()

	if err2 != nil {
		autoLog.Sugar.Infof("原神启动器已关闭")
	} else {
		autoLog.Sugar.Infof("原神启动器关闭成功")
	}

}

// 鼠标点击(x、y是鼠标坐标，key是键，是否双击)
func MouseClick(x, y int, key string, DoubleClick bool) {

	// 移动鼠标到指定位置
	robotgo.Move(x, y)

	// 等待一小会儿，确保鼠标移动完成
	time.Sleep(500 * time.Millisecond)

	// 模拟鼠标左键点击
	robotgo.Click(key, DoubleClick) // 第二个参数为 true 表示双击，false 表示单击
}

// 截图
func ScreenShot() error {
	// 获取当前屏幕大小
	screenWidth, screenHeight := robotgo.GetScreenSize()

	// 创建一个与屏幕大小相同的图像
	imgScreen := robotgo.CaptureScreen(0, 0, screenWidth, screenHeight)
	if imgScreen == nil {
		return fmt.Errorf("截图失败: 无法获取屏幕图像")
	}
	defer robotgo.FreeBitmap(imgScreen) // 确保释放资源

	img := robotgo.ToImage(imgScreen)
	imgo.Save("jt.png", img)

	time.Sleep(2000 * time.Millisecond)

	return nil
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

func GetMysQDXy() {
	currentDir, _ := os.Getwd()

	baseURL := "http://127.0.0.1:8888/qdXy"
	params := url.Values{}
	params.Add("te", currentDir+"\\modeImage\\qdModel.png")
	params.Add("ta", currentDir+"\\modeImage\\qdAll.png")
	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	// 发送 GET 请求
	resp, err := http.Get(fullURL)
	if err != nil {
		fmt.Println("请求失败:", err)
		return
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("读取响应失败:", err)
		return
	}

	fmt.Println("响应内容:", string(body))
}

func GetMysSignLog() string {
	readLogURL := "http://127.0.0.1:8888/read-log"
	resp, err := http.Get(readLogURL)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	return string(body)
}
