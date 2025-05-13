package bgiStatus

import (
	"auto-bgi/config"
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/robfig/cron/v3"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// 检查 BetterGI.exe 是否在运行
func IsWechatRunning() bool {
	cmd := exec.Command("tasklist", "/FI", "IMAGENAME eq BetterGI.exe")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error executing tasklist:", err)
		return false
	}
	return strings.Contains(string(output), "BetterGI.exe")
}

var Config = config.Cfg

// 向企业微信发送通知（文本）
func SendWeChatNotification(content string) {

	// 通知内容
	message := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			//"content": "BetterGI 已经关闭:\n" + Config.Content + "/test",
			"content": content,
		},
	}
	jsonData, err := json.Marshal(message)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}

	client := &http.Client{}

	req, err := http.NewRequest("POST", Config.WebhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("Response Status:", resp.Status)
	fmt.Println("BetterGI 已关闭，通知已发送")
}

// 向企业微信发送通知（图片）
func SendWeChatImage(path string) error {

	//获取本地文件
	// 读取图片文件
	imageData, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Printf("Error reading image file: %v\n", err)
		return err
	}
	// 计算 Base64 编码
	base64Data := base64.StdEncoding.EncodeToString(imageData)

	// 计算 MD5 哈希
	md5Hash := md5.Sum(imageData)
	md5String := hex.EncodeToString(md5Hash[:])

	// 通知内容
	message := map[string]interface{}{
		"msgtype": "image",
		"image": map[string]string{
			"base64": base64Data,
			"md5":    md5String,
		},
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return err
	}

	client := &http.Client{}

	req, err := http.NewRequest("POST", Config.WebhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return err
	}
	defer resp.Body.Close()

	return nil
}

var notified = false

func CheckBetterGIStatus() {

	cronTab := cron.New(cron.WithSeconds())

	// 定时任务,cron表达式
	spec := "*/30 * * * * *"

	task := func() {
		// 检查进程
		if IsWechatRunning() {
			fmt.Print("\rBetterGI 正在运行", time.Now().Format("2006-01-02 15:04:05"))
			notified = false // 清除通知状态
		} else {
			if !notified {
				SendWeChatNotification("BetterGI 已经关闭:\\n\" + Config.Content + \"/test")
				notified = true
			} else {
				fmt.Print("\rBetterGI 已关闭，已通知过", time.Now().Format("2006-01-02 15:04:05"))
			}
		}
	}
	// 添加定时任务
	cronTab.AddFunc(spec, task)
	// 启动定时器
	cronTab.Start()
	// 阻塞主线程停止
	select {}
}

func Progress(filename string, line string) (string, error) {

	start := strings.Index(line, `"`)
	end := strings.LastIndex(line, `"`)

	content := "0/0"
	// 检查是否找到了两个引号且位置有效
	if start == -1 || end == -1 || start >= end {
		content = line
	} else {
		content = line[start+1 : end]
	}

	// 1. 读取 JSON 文件
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("读取文件失败: %v", err)
	}
	// 2. 解析为 map[string]interface{}（保持原始结构）
	var jsonData map[string]interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		log.Fatalf("解析 JSON 失败: %v", err)
		return "", err
	}
	// 3. 获取 projects 数组
	projects, ok := jsonData["projects"].([]interface{})
	if !ok {
		log.Fatal("projects 字段不是数组或不存在")
		return "", err
	}
	//fmt.Println(len(projects))
	fmt.Println(content)
	pro := "0/0"
	for i, project := range projects {
		projectMap := project.(map[string]interface{})
		if projectMap["name"] == content {
			pro = fmt.Sprintf("当前进度:%d/%d", i, len(projects))
			break
		}
	}

	return pro, nil
}

// 根据配置组文件名字找到排序号
func GetGroupNum(filename string) (int, error) {

	// 1. 读取 JSON 文件
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("读取文件失败: %v", err)
		return 0, err
	}
	// 2. 解析为 map[string]interface{}（保持原始结构）
	var jsonData map[string]interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		log.Fatalf("解析 JSON 失败: %v", err)
		return 0, err
	}
	// 3. 获取 projects 数组
	index, ok := jsonData["index"].(interface{})
	if !ok {
		log.Fatal("projects 字段不是数组或不存在")
		return 0, err
	}

	return int(index.(float64)), nil
}

func TodayHarvest() (map[string]int, error) {
	fmt.Println("今日收获统计")
	re := regexp.MustCompile(`^交互或拾取："([^"]*)"`)

	// 生成日志文件名
	date := time.Now().Format("20060102")
	filename := filepath.Clean(fmt.Sprintf("%s\\log\\better-genshin-impact%s.log", Config.BetterGIAddress, date))

	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	// 初始化map用于存储物品和出现次数
	harvestStats := make(map[string]int)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		matches := re.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) > 1 {
				item := match[1]
				harvestStats[item]++
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("读取文件错误: %v", err)
	}

	for item, count := range harvestStats {
		fmt.Printf("%s: %d\n", item, count)
	}

	return harvestStats, nil
}

type Material struct {
	Data string
	Cl   string
}

func BagStatistics() ([]Material, error) {
	fmt.Println("背包统计")
	filename := filepath.Clean(fmt.Sprintf("%s\\User\\JsScript\\背包材料统计\\recognized_materials.txt", Config.BetterGIAddress))

	// 打开文件
	file, err := os.Open(filename) // 替换为你的文件路径
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 创建一个扫描器来读取文件
	scanner := bufio.NewScanner(file)

	// 创建一个正则表达式来匹配日期格式 "YYYY/M/D HH:MM:SS"
	re1 := regexp.MustCompile(`\b\d{4}/\d{1,2}/\d{1,2} \d{2}:\d{2}:\d{2}\b`)

	statistics := Config.BagStatistics

	split := strings.Split(statistics, ",")

	var bags []Material
	var bag Material

	for scanner.Scan() {
		for _, s := range split {
			// 创建一个正则表达式来匹配 "晶蝶：数字" 模式
			sprintf := fmt.Sprintf(`%s: (\d+)`, s)
			re := regexp.MustCompile(sprintf)

			line := scanner.Text()

			//日期匹配
			if re1.MatchString(line) {
				bag.Data = line
			}

			// 查找当前行中所有匹配
			match := re.FindString(line)
			if match != "" {
				// 提取数字部分并存储
				bag.Cl = match
				bags = append(bags, bag)
			}
		}

		// 检查扫描器是否有错误
		if err := scanner.Err(); err != nil {
			panic(err)
		}
	}

	return bags, nil
}

func DeleteBagStatistics() string {
	fmt.Println("背包统计")
	filePath := filepath.Clean(fmt.Sprintf("%s\\User\\JsScript\\背包材料统计\\recognized_materials.txt", Config.BetterGIAddress))
	// 删除文件
	err := os.Remove(filePath)
	if err != nil {
		fmt.Println("删除文件失败:", err)
		return "删除文件失败"
	}
	fmt.Println("文件删除成功")
	return "文件删除成功"
}
