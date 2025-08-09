package tools

import (
	"auto-bgi/config"
	"fmt"
	"net"
	"os"
	"regexp"
	"time"
)

// ExtractLogTime 从日志字符串中提取 [HH:MM:SS.mmm] 格式的时间，并解析为 time.Time 对象
func ExtractLogTime(logLine string) (string, error) {

	today := time.Now().Format("2006-01-02")

	re := regexp.MustCompile(`\[(\d{2}:\d{2}:\d{2}\.\d{3})\]`)
	matches := re.FindStringSubmatch(logLine)
	if len(matches) < 2 {
		//return time.Time{}, fmt.Errorf("未找到时间字段")
		return "", fmt.Errorf("未找到时间字段")
	}

	timeStr := matches[1]
	parsedTime, err := time.Parse("15:04:05", timeStr)
	if err != nil {
		return "", fmt.Errorf("解析时间失败: %w", err)
	}

	timeStr = today + " " + parsedTime.Format("15:04:05")

	return timeStr, nil
}

// ExtractLogTime 从日志字符串中提取 [HH:MM:SS.mmm] 格式的时间，并解析为 time.Time 对象
func ExtractLogTime2(today string, logLine string) (string, error) {

	re := regexp.MustCompile(`\[(\d{2}:\d{2}:\d{2}\.\d{3})\]`)
	matches := re.FindStringSubmatch(logLine)
	if len(matches) < 2 {
		//return time.Time{}, fmt.Errorf("未找到时间字段")
		return "", fmt.Errorf("未找到时间字段")
	}

	timeStr := matches[1]
	parsedTime, err := time.Parse("15:04:05", timeStr)
	if err != nil {
		return "", fmt.Errorf("解析时间失败: %w", err)
	}

	timeStr = today + " " + parsedTime.Format("15:04:05")

	return timeStr, nil
}

var timePattern = regexp.MustCompile(`\[\d{2}:\d{2}:\d{2}(\.\d{1,3})?\]`)

// HasTimestamp 判断一行日志是否包含时间戳
func HasTimestamp(line string) bool {
	return timePattern.MatchString(line)
}

// 计算执行时间
func CalculateDuration(start, end string) string {
	layout := "2006-01-02 15:04:05" // 根据日志时间格式调整
	startTime, err1 := time.Parse(layout, start)
	endTime, err2 := time.Parse(layout, end)

	if err1 == nil && err2 == nil {
		return endTime.Sub(startTime).String()
	}
	return ""
}

// ListSubDirsOnly 列出目录下的所有子目录
func ListSubDirsOnly(dirPath string) ([]string, error) {
	var subDirs []string

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			subDirs = append(subDirs, entry.Name())
		}
	}

	return subDirs, nil
}

func GetLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ip := ipnet.IP.To4(); ip != nil {
				// 跳过 169.254.x.x（APIPA）
				if ip[0] == 169 && ip[1] == 254 {
					continue
				}
				return ip.String() + config.Cfg.Post, nil
			}
		}
	}
	return "", fmt.Errorf("没有找到有效的局域网 IP")
}
