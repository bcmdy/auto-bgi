package OneRemote

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// GbkToUtf8 将 GBK 字节流转换为 UTF-8 字符串
func GbkToUtf8(s []byte) (string, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	d, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(d), nil
}

type SessionInfo struct {
	SessionName string
	Username    string
	ID          string
	State       string
}

// GetDetailedSessions 获取所有多用户
func GetDetailedSessions() ([]SessionInfo, error) {
	// 使用完整路径调用 qwinsta
	cmd := exec.Command("C:\\Windows\\System32\\qwinsta.exe")

	// 获取混合输出（包含标准输出和错误输出）
	output, err := cmd.CombinedOutput()

	if len(output) == 0 {
		if err != nil {
			return nil, fmt.Errorf("未获取到任何输出: %v", err)
		}
		return nil, fmt.Errorf("未获取到任何输出")
	}

	// 【关键步骤】解码 GBK 输出为 UTF-8
	utf8Str, err := GbkToUtf8(output)
	if err != nil {
		return nil, fmt.Errorf("编码转换失败: %v", err)
	}

	lines := strings.Split(utf8Str, "\r\n")
	var results []SessionInfo

	// 正则匹配：匹配 2 个及以上的空格
	re := regexp.MustCompile(`\s{2,}`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.Contains(line, "会话名") || strings.Contains(line, "SESSIONNAME") {
			continue
		}

		// 处理当前会话标志 '>'
		line = strings.TrimPrefix(line, ">")
		line = strings.TrimSpace(line)

		fields := re.Split(line, -1)

		// 动态解析列
		s := SessionInfo{}
		if len(fields) >= 2 {
			// 如果第一列是数字，说明没有会话名（SessionName 为空）
			if regexp.MustCompile(`^\d+$`).MatchString(fields[0]) {
				s.ID = fields[0]
				if len(fields) >= 2 {
					s.State = fields[1]
				}
			} else {
				s.SessionName = fields[0]
				// 正常包含用户名的行通常有 4 列：会话名 用户名 ID 状态
				if len(fields) >= 4 {
					s.Username = fields[1]
					s.ID = fields[2]
					s.State = fields[3]
				} else if len(fields) == 3 {
					// 只有 会话名 ID 状态
					s.ID = fields[1]
					s.State = fields[2]
				}
			}

			if strings.EqualFold(strings.TrimSpace(s.Username), "Administrator") {
				continue
			}

			if s.ID != "" {
				results = append(results, s)
			}
		}
	}
	return results, nil
}

func FindSessionIDByUsername(username string) (string, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return "", fmt.Errorf("username is empty")
	}

	sessions, err := GetDetailedSessions()
	if err != nil {
		return "", err
	}

	for _, s := range sessions {
		if s.ID == "" {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(s.Username), username) {
			return s.ID, nil
		}
	}
	return "", fmt.Errorf("未找到用户会话: %s", username)
}

// 启动1Remote用户
func startOneRemoteUser(sessionName string) error {
	// 1Remote.exe 的路径，如果在环境变量中可直接写名
	// 注意：根据你的安装路径修改
	appPath := "D:\\1R\\1Remote.exe"

	if _, err := os.Stat(appPath); err != nil {
		return fmt.Errorf("1Remote.exe 不存在或不可访问: %s", appPath)
	}

	// 构造命令：1Remote.exe --launcher "MyWindowsServer"
	cmd := exec.Command(appPath, "--launcher", sessionName)

	// 运行命令
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("无法启动 1Remote: %v", err)
	}

	log.Printf("已成功触发 1Remote 连接: %s", sessionName)
	return nil
}

// ForceLogoff 注销1Remote用户
// ForceLogoff 函数用于强制注销指定的远程会话
// 参数:
//
//	sessionID: 要注销的会话ID字符串
func ForceLogoff(sessionID string) error {
	// 格式: logoff [ID] /server:[IP]
	cmd := exec.Command("logoff", sessionID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("强制注销远程会话失败: %v, output: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}
