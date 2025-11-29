package tools

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
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

func GetLocalIPs() ([]string, error) {
	var ips []string

	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range ifaces {
		// 跳过未启用和回环接口
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		// 跳过 Docker、VMWare、VPN 等虚拟网卡
		name := strings.ToLower(iface.Name)
		if strings.Contains(name, "docker") ||
			strings.Contains(name, "vmnet") ||
			strings.Contains(name, "vbox") {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ip := ipnet.IP.To4(); ip != nil {
					// 跳过 APIPA 地址 169.254.x.x
					if ip[0] == 169 && ip[1] == 254 {
						continue
					}
					ips = append(ips, ip.String())
				}
			}
		}
	}

	if len(ips) == 0 {
		return nil, fmt.Errorf("没有找到有效的局域网 IP")
	}
	return ips, nil
}

// 安全提取日志时间，如果 line 为空或提取失败，则返回默认时间
func ExtractLogTime2Safe(date string, line string) string {
	if line == "" {
		// 如果没有行信息，直接返回当日最后时间
		return fmt.Sprintf("%s 23:59:59", date)
	}
	t, err := ExtractLogTime2(date, line)
	if err != nil || t == "" {
		// 提取失败，则返回当日最后时间
		return fmt.Sprintf("%s 23:59:59", date)
	}
	return t
}

// 查找 repo 目录下是否存在名为 targetFolder 的子文件夹
func FindSubFolder(root string, targetFolder string) (string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return "", err
	}

	for _, entry := range entries {
		if entry.IsDir() && entry.Name() == targetFolder {
			return filepath.Join(root, entry.Name()), nil
		}
	}

	return "", fmt.Errorf("未找到子文件夹: %s", targetFolder)
}

// FindJSONFiles 查找指定目录及其子目录中的所有JSON文件
func FindJSONFiles(rootDir string) ([]string, error) {
	var jsonFiles []string

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录，只处理文件
		if info.IsDir() {
			return nil
		}

		// 检查文件扩展名是否为.json
		if strings.EqualFold(filepath.Ext(path), ".json") {
			jsonFiles = append(jsonFiles, path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("遍历目录时出错: %v", err)
	}

	return jsonFiles, nil
}

// 查询指定目录下的文件夹
func ListDirectories(dirPath string) ([]string, error) {
	var directories []string

	// 读取目录内容
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	// 只选择目录
	for _, entry := range entries {
		if entry.IsDir() {
			directories = append(directories, entry.Name())
		}
	}

	return directories, nil
}

// CompareVersion 版本比较
// 返回值：1：v1大于v2，-1：v1小于v2，0：相等
func CompareVersion(v1, v2 string) int {
	// 按 . 拆分
	s1 := strings.Split(v1, ".")
	s2 := strings.Split(v2, ".")

	// 补齐长度
	n := len(s1)
	if len(s2) > n {
		n = len(s2)
	}

	for i := 0; i < n; i++ {
		var num1, num2 int
		if i < len(s1) {
			num1, _ = strconv.Atoi(s1[i])
		}
		if i < len(s2) {
			num2, _ = strconv.Atoi(s2[i])
		}
		if num1 > num2 {
			return 1 // v1 > v2
		} else if num1 < num2 {
			return -1 // v1 < v2
		}
	}
	return 0 // 相等
}

// 打开指定网页
func OpenBrowser(url string) error {
	var cmd *exec.Cmd

	cmd = exec.Command("cmd", "/c", "start", url)

	return cmd.Start()
}

// 字符串数组怎么去重
func UniqueStrings(arr []string) []string {
	seen := make(map[string]struct{})
	result := []string{}
	for _, v := range arr {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}
	return result
}

// CompareDirs 判断两个文件夹内容是否一致
func CompareDirs(dir1, dir2 string) (bool, error) {
	files1 := make(map[string][16]byte)
	files2 := make(map[string][16]byte)

	err := filepath.Walk(dir1, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			rel, _ := filepath.Rel(dir1, path)
			hash, err := hashFileMD5(path)
			if err != nil {
				return err
			}
			files1[rel] = hash
		}
		return nil
	})
	if err != nil {
		return false, err
	}

	err = filepath.Walk(dir2, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			rel, _ := filepath.Rel(dir2, path)
			hash, err := hashFileMD5(path)
			if err != nil {
				return err
			}
			files2[rel] = hash
		}
		return nil
	})
	if err != nil {
		return false, err
	}

	if len(files1) != len(files2) {
		return false, nil
	}

	for k, v := range files1 {
		if hv, ok := files2[k]; !ok || v != hv {
			return false, nil
		}
	}

	return true, nil
}

// hashFileMD5 计算文件的 MD5 哈希
func hashFileMD5(filePath string) ([16]byte, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return [16]byte{}, err
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return [16]byte{}, err
	}
	var result [16]byte
	copy(result[:], h.Sum(nil))
	return result, nil
}

// 解密
func Decrypt(encryptedText, key string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return "", errors.New("key length must be 16, 24, or 32 bytes")
	}

	cipherText, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", err
	}

	if len(cipherText) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}

	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(cipherText, cipherText)

	return string(cipherText), nil
}

// 字符串转数字
func StringToInt(i string) int {

	atoi, err := strconv.Atoi(i)
	if err != nil {
		log.Fatal(err)
	}
	return atoi

}
