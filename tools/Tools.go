package tools

import (
	"archive/zip"
	"auto-bgi/abgiConstant"
	"auto-bgi/autoLog"
	"auto-bgi/control"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/agnivade/levenshtein"
	"github.com/getlantern/systray"
	"github.com/go-resty/resty/v2"
	"io"
	"log"
	"math"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
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

// DetermineFileType 检查一个字符串是否包含字典数组中的任何词语
func DetermineFileType(target string, typeS ...string) bool {
	for _, word := range typeS {

		if strings.HasSuffix(target, word) {
			return true // 如果找到任何一个词语，就返回 true
		}
	}
	return false // 遍历完所有词语都没有找到，则返回 false
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

func Encrypt(plainText string) (string, error) {
	block, err := aes.NewCipher([]byte(abgiConstant.ABgiKey))
	if err != nil {
		return "", err
	}

	cipherText := make([]byte, aes.BlockSize+len(plainText))
	iv := cipherText[:aes.BlockSize]

	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], []byte(plainText))

	return base64.StdEncoding.EncodeToString(cipherText), nil
}

// 解密
func Decrypt(encryptedText string) (string, error) {
	block, err := aes.NewCipher([]byte(abgiConstant.ABgiKey))
	if err != nil {
		return "", err
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

// Unzip 解压 zip 文件到指定目录
func Unzip(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		// 构建解压后的完整路径
		fpath := filepath.Join(destDir, f.Name)

		// 判断是否是目录
		if f.FileInfo().IsDir() {
			// 创建目录
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// 确保父目录存在
		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		// 打开 zip 内文件
		rc, err := f.Open()
		if err != nil {
			return err
		}

		// 创建目标文件
		outFile, err := os.Create(fpath)
		if err != nil {
			rc.Close()
			return err
		}

		// 拷贝内容
		_, err = io.Copy(outFile, rc)

		// 关闭文件句柄
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

// 重启程序的函数
func RestartProgram() error {

	// 调用 run_auto_bgi.vbs 来启动更新后的程序
	cmd := exec.Command("wscript", "run_auto_bgi.vbs")
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("启动 VBS 脚本失败: %v", err)
	}
	systray.Quit() // 或确保内部调用 NIM_DELETE

	// 退出当前进程
	os.Exit(0)
	return nil
}

func computeDistance(a, b string) (int, float64) {
	distance := levenshtein.ComputeDistance(a, b)
	//fmt.Println("编辑距离:", distance)

	// 转化为相似度（0~1之间）
	similarity := 1 - float64(distance)/float64(max(len([]rune(a)), len([]rune(b))))
	//fmt.Printf("相似度: %.2f\n", similarity)
	return distance, similarity
}

// GetStandardName 根据输入名称返回匹配度最高的标准材料名
// rawName: 原始名称 (例如 "月萤虫" 或 OCR 识别出的 "琉鳞石")
func GetStandardName(rawName string) string {
	// 1. 特殊别名映射 (如果以后别名多了，建议提取为全局 map)
	if rawName == "月萤虫" {
		return "晶蝶"
	}

	// 初始化最佳匹配
	bestMatch := rawName // 默认回退到原始名称
	maxScore := 0.0      // 记录最高相似度

	// 2. 遍历标准库 (abgiConstant.Material 需在包级可见)
	for _, stdName := range abgiConstant.Material {
		// 性能优化：如果完全相等，直接返回，不做复杂计算
		if rawName == stdName {
			return stdName
		}

		// 计算相似度
		_, score := computeDistance(rawName, stdName)

		// 更新最佳匹配
		if score > maxScore {
			maxScore = score
			bestMatch = stdName

			// 性能优化：如果相似度已经是 1.0 (完美匹配)，没必要继续找了
			if score >= 1.0 {
				return stdName
			}
		}
	}

	// 3. 阈值检查与日志记录
	// 注意：虽然记录了日志，但为了程序的连贯性，通常还是返回在这个阈值下找到的“最像”的那个，或者原名
	//if maxScore < 0.1 {
	//	// 这里假设 autoLog 是全局变量，如果不是，需要作为参数传入或通过回调处理
	//	// autoLog.Sugar.Infof("未知材料: %s (最高相似度: %.2f, 匹配项: %s)", rawName, maxScore, bestMatch)
	//	fmt.Printf("[Log] 未知材料: %s (最高相似度: %.2f)\n", rawName, maxScore)
	//}

	return bestMatch
}

// 保留指定位小数
func RoundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

var httpClient = &http.Client{
	Timeout: 15 * time.Second, // 默认超时时间 15 秒
}

func PostHttp[T any, R any](url string, body T, res *R) (*R, int, error) {
	client := resty.New()

	resp, err := client.R().
		SetHeaders(map[string]string{
			"Content-Type": "application/json",
		}).
		SetBody(body).
		SetResult(res).
		Post(url)

	if err != nil {
		return nil, 0, err
	}

	// 检查 HTTP 状态码（可选，根据需要调整）
	if resp.StatusCode() >= 400 {
		return nil, resp.StatusCode(), fmt.Errorf("HTTP error: %d", resp.StatusCode())
	}

	// 返回反序列化后的结果
	return res, resp.StatusCode(), nil
}

// IsDateLess 判断 date1 是否早于 date2
// 支持常见日期格式：
// 2006-01-02
// 2006-01-02 15:04:05
// 2006/01/02
// 2006/01/02 15:04:05
func IsDateLess(date1, date2 string) (bool, error) {
	//日期相同直接返回
	if date1 == date2 {
		return true, nil
	}

	layouts := []string{
		"2006-01-02",
		"2006-01-02 15:04:05",
		"2006/01/02",
		"2006/01/02 15:04:05",
	}

	parse := func(dateStr string) (time.Time, error) {
		for _, layout := range layouts {
			if t, err := time.ParseInLocation(layout, dateStr, time.Local); err == nil {
				return t, nil
			}
		}
		return time.Time{}, errors.New("不支持的日期格式: " + dateStr)
	}

	t1, err := parse(date1)
	if err != nil {
		return false, err
	}

	t2, err := parse(date2)
	if err != nil {
		return false, err
	}

	return t1.Before(t2), nil
}

// 关闭软件
func CloseSoftware(name string) {
	//关闭软件之前，停止当前脚本/独立任务

	time.Sleep(5 * time.Second)

	//// 获取当前用户名
	//username := os.Getenv("USERNAME")
	//if username == "" {
	//	username = os.Getenv("USER")
	//}

	username := control.GetSysTemUser()

	autoLog.Sugar.Infof("当前window用户是：%s", username)
	// 构建正确的命令
	cmd := exec.Command("cmd", "/c", "taskkill", "/F", "/IM", name, "/FI", fmt.Sprintf("USERNAME eq %s", username))

	// 创建命令
	//cmd := exec.Command("cmd", "/c", "taskkill", "/F", "/IM", name, "/FI", "USERNAME eq %USERNAME%")

	//cmd := exec.Command("taskkill", "/F", "/IM", "BetterGI.exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Stdout = nil
	cmd.Stderr = nil

	// 执行命令并获取输出
	_, err := cmd.CombinedOutput()

	if err != nil {
		autoLog.Sugar.Errorf("执行命令出错: %v\n", err)
	}
	autoLog.Sugar.Infof("%s关闭成功", name)

}

// IsSameDay 判断两个时间是否在同一天
func IsSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func ZipDir(sourceDir, zipFilePath string, keepRoot bool) error {

	////清理历史备份
	//_ = ClearDir("Users")

	fmt.Println("压缩目录:", sourceDir)
	fmt.Println("输出文件:", zipFilePath)

	zipFile, err := os.Create(zipFilePath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	base := filepath.Clean(sourceDir)
	parent := filepath.Dir(base)

	err = filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			// ✅ 不写入目录条目，让解压自动生成
			return nil
		}
		if !info.Mode().IsRegular() {
			return nil
		}

		// 计算压缩包内路径
		var relPath string
		if keepRoot {
			relPath, _ = filepath.Rel(parent, path) // 保留根目录
		} else {
			relPath, _ = filepath.Rel(base, path) // 去掉根目录
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(relPath) // ✅ 统一分隔符
		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})

	return err
}
