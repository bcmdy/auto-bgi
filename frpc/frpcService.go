package frpc

import (
	"auto-bgi/abgiConstant"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

//go:embed frpc.exe
var embeddedFiles embed.FS

// 1. 内置的 Server 配置 (写死在代码里，用户看不到也改不了)
// 注意：确保最后有一个换行符 \n，否则拼接时会出错

// 用户自定义配置文件的文件名
const userConfigFileName = "frpc_user.toml"

func InitFrp() {
	// --- 第一步：检查并读取用户的外部配置文件 ---
	userConfigBytes, err := os.ReadFile(userConfigFileName)
	if err != nil {
		// 如果用户没有这个文件，我们可以帮他生成一个模板，或者报错
		fmt.Printf("未找到 %s，请确保配置文件在当前目录下。\n", userConfigFileName)
		createTemplate() // 这是一个可选功能，帮用户生成模板
		return
	}

	// --- 第二步：合并配置 (内置 + 外置) ---
	// 拼接逻辑：Server配置 + 换行 + 用户配置
	fullConfigContent := abgiConstant.ServerConfig + "\n" + string(userConfigBytes)

	// --- 第三步：准备临时环境 ---
	tempDir, err := os.MkdirTemp("", "frpc-runner")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tempDir) // 退出时清理

	fmt.Println("正在初始化运行环境...")

	// 1. 释放 frpc.exe 到临时目录
	frpcExePath, err := extractFile("frpc.exe", "frpc.exe", tempDir)
	if err != nil {
		panic(err)
	}

	// 2. 将合并后的配置写入临时文件 (merged_frpc.toml)
	tempConfigPath := filepath.Join(tempDir, "merged_frpc.toml")
	err = os.WriteFile(tempConfigPath, []byte(fullConfigContent), 0644)
	if err != nil {
		fmt.Printf("写入临时配置失败: %v\n", err)
		return
	}

	// --- 第四步：启动 FRPC ---
	fmt.Println("正在启动穿透服务...")
	fmt.Printf("加载配置: %s\n", userConfigFileName)

	cmd := exec.Command(frpcExePath, "-c", tempConfigPath)

	// 隐藏窗口 (Windows)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}

	// 可选：将 frpc 的日志输出到当前控制台，方便调试
	// 如果你想完全静默，可以注释掉下面两行
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		fmt.Printf("启动失败: %v\n", err)
		return
	}

	fmt.Println("服务启动成功！PID:", cmd.Process.Pid)
	fmt.Println("请勿关闭此窗口...")

	// 阻塞等待
	cmd.Wait()
}

// 辅助：从 embed 提取文件
func extractFile(embedPath string, fileName string, targetDir string) (string, error) {
	data, err := embeddedFiles.ReadFile(embedPath)
	if err != nil {
		return "", err
	}
	targetPath := filepath.Join(targetDir, fileName)
	err = os.WriteFile(targetPath, data, 0755)
	return targetPath, err
}

// 辅助：如果用户没有配置文件，自动生成一个模板
func createTemplate() {
	tpl := `[[proxies]]
name = "demo"
type = "tcp"
localIP = "127.0.0.1"
localPort = 80
remotePort = 8080`
	os.WriteFile(userConfigFileName, []byte(tpl), 0644)
	fmt.Println("已自动生成示例配置文件，请修改后重新运行程序。")
}
