package bgiStatus

import (
	"archive/zip"
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"auto-bgi/control"
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/otiai10/copy"
	"github.com/robfig/cron/v3"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// 检查 BetterGI.exe 是否在运行
func IsWechatRunning() bool {
	cmd := exec.Command("tasklist", "/FI", "IMAGENAME eq BetterGI.exe")
	output, err := cmd.Output()
	if err != nil {

		autoLog.Sugar.Error("BetterGI.exe 是否在运行:", err)
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
		autoLog.Sugar.Error("Error marshaling JSON:", err)
		return
	}

	client := &http.Client{}

	req, err := http.NewRequest("POST", Config.WebhookURL, bytes.NewBuffer(jsonData))
	if err != nil {

		autoLog.Sugar.Error("Error creating request:", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		autoLog.Sugar.Error("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	autoLog.Sugar.Info("企业微信机器人配置错误:", resp.Status)
	autoLog.Sugar.Info("BetterGI 已关闭，通知已发送")
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
		autoLog.Sugar.Error("Error marshaling JSON:", err)
		return err
	}

	client := &http.Client{}

	req, err := http.NewRequest("POST", Config.WebhookURL, bytes.NewBuffer(jsonData))
	if err != nil {

		autoLog.Sugar.Error("Error creating request:", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {

		autoLog.Sugar.Error("Error sending request:", err)
		return err
	}
	defer resp.Body.Close()

	return nil
}

var notified = false
var okInform = false

func CheckBetterGIStatus() {

	cronTab := cron.New(cron.WithSeconds())

	// 定时任务,cron表达式
	spec := "*/30 * * * * *"

	task := func() {

		// 检查进程
		if IsWechatRunning() {
			//fmt.Print("\rBetterGI 正在运行", time.Now().Format("2006-01-02 15:04:05"))

			autoLog.Sugar.Infof("BetterGI 正在运行: %s", time.Now().Format("2006-01-02 15:04:05"))
			notified = false // 清除通知状态
		} else {
			if !notified {
				SendWeChatNotification("BetterGI 已经关闭:" + Config.Content)
				control.CloseYuanShen()
				notified = true
			} else if !okInform {
				autoLog.Sugar.Infof("BetterGI 已关闭，已通知过: %s", time.Now().Format("2006-01-02 15:04:05"))
				okInform = true
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

func JsProgress(filename string, pattern string) (string, error) {

	// 1. 读取 JSON 文件

	re := regexp.MustCompile(pattern)

	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// 用于存储最后匹配的行和配置组名称
	var lastMatch string

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		matches := re.FindStringSubmatch(line)
		if matches != nil {
			lastMatch = line

		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	// 输出结果
	if lastMatch != "" {
		//autoLog.Sugar.Infof("最后匹配的行: %s", lastMatch)
	} else {
		errs := fmt.Errorf("没有找到匹配的行", 500)
		return "", errs
	}
	return lastMatch, nil
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

		autoLog.Sugar.Errorf("解析 JSON 失败: %v", err)
		return "", err
	}
	// 3. 获取 projects 数组
	projects, ok := jsonData["projects"].([]interface{})
	if !ok {
		log.Fatal("projects 字段不是数组或不存在")
		return "", err
	}
	pro := "0/0"
	for i, project := range projects {
		projectMap := project.(map[string]interface{})
		if projectMap["name"] == content {
			pro = fmt.Sprintf("%d/%d", i, len(projects))
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

	autoLog.Sugar.Infof("今日收获统计")
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

	return harvestStats, nil
}

type Material struct {
	Data string
	Cl   string
	Num  string
}

func BagStatistics() ([]Material, error) {
	autoLog.Sugar.Infof("背包统计")
	filename := filepath.Clean(fmt.Sprintf("%s\\User\\JsScript\\背包材料统计\\latest_record.txt", Config.BetterGIAddress))

	// 打开文件
	file, err := os.Open(filename) // 替换为你的文件路径
	if err != nil {
		autoLog.Sugar.Errorf("背包统计失败: %v", err)
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
				split := strings.Split(match, ":")

				bag.Cl = split[0]
				bag.Num = split[1]

				bags = append(bags, bag)
			}
		}

		// 检查扫描器是否有错误
		if err := scanner.Err(); err != nil {
			panic(err)
		}
	}
	morasStatistics, _ := MorasStatistics()
	bags = append(bags, morasStatistics...)

	return bags, nil
}

// 摩拉统计
func MorasStatistics() ([]Material, error) {

	autoLog.Sugar.Infof("摩拉统计")
	filename := filepath.Clean(fmt.Sprintf("%s\\User\\JsScript\\OCR读取当前摩拉记录并发送通知\\mora_log.txt", Config.BetterGIAddress))
	// 打开文件
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var bags []Material

	// 创建一个扫描器来读取文件
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var bag Material
		line := scanner.Text()
		split := strings.Split(line, " - ")
		bag.Data = split[0]

		ClNum := strings.Split(split[1], ":")
		bag.Cl = ClNum[0]
		bag.Num = ClNum[1]
		bags = append(bags, bag)
	}
	return bags, nil
}

// 删除背包统计
func DeleteBagStatistics() string {

	autoLog.Sugar.Infof("删除背包统计")
	filePath := filepath.Clean(fmt.Sprintf("%s\\User\\JsScript\\背包材料统计\\latest_record.txt", Config.BetterGIAddress))
	// 删除文件
	err := os.Remove(filePath)
	if err != nil {
		fmt.Println("背包统计删除文件失败:", err)

	}

	autoLog.Sugar.Infof("删除摩拉统计")
	filePath2 := filepath.Clean(fmt.Sprintf("%s\\User\\JsScript\\OCR读取当前摩拉记录并发送通知\\mora_log.txt", Config.BetterGIAddress))
	// 删除文件
	err2 := os.Remove(filePath2)
	if err2 != nil {
		autoLog.Sugar.Errorf("删除摩拉统计失败")

	}
	autoLog.Sugar.Infof("文件删除成功")
	return "文件删除成功"
}

type DogFood struct {
	FileName string
	Detail   []string
}

func GetAutoArtifactsPro() ([]DogFood, error) {
	// 获取当前目录下所有 .txt 文件
	files, err := filepath.Glob(fmt.Sprintf("%s\\User\\JsScript\\AutoArtifactsPro\\records\\*.txt", Config.BetterGIAddress))
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("未找到任何txt文件")
	}
	var data []DogFood
	for _, filename := range files {
		file, err := os.Open(filename)

		if err != nil {

			autoLog.Sugar.Errorf("打开文件失败: %s, 错误: %v\n", filename, err)
			continue
		}
		defer file.Close()

		var dogFood DogFood

		dogFood.FileName = filepath.Base(filename)

		scanner := bufio.NewScanner(file)
		inHistory := false

		for scanner.Scan() {
			line := scanner.Text()
			if !inHistory {
				if strings.HasPrefix(line, "历史收益：") {
					inHistory = true
				}
				continue
			}
			dogFood.Detail = append(dogFood.Detail, line)

		}

		data = append(data, dogFood)

		if err := scanner.Err(); err != nil {

			autoLog.Sugar.Errorf("读取文件出错: %s, 错误: %v\n", filename, err)
		}

	}

	return data, nil
}

type EarningsData struct {
	Dates  []string `json:"dates"`
	Line   []string `json:"line"`
	DogExp []int    `json:"dogExp"`
	Mora   []int    `json:"mora"`
}

func GetAutoArtifactsPro2(fileName string) (*EarningsData, error) {

	autoLog.Sugar.Infof("狗粮查询")
	filePath := filepath.Clean(fmt.Sprintf("%s\\User\\JsScript\\AutoArtifactsPro\\records\\%s", Config.BetterGIAddress, fileName))
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	data := &EarningsData{}
	inHistory := false

	for scanner.Scan() {

		line := scanner.Text()
		if !inHistory {
			if strings.HasPrefix(line, "历史收益：") {
				inHistory = true
			}
			continue
		}
		// 1. 分割字符串，获取日期部分
		parts := strings.Split(line, "，")
		if len(parts) < 1 {
			autoLog.Sugar.Errorf("字符串格式不正确，无法提取日期。")
			continue
		}
		//日期

		// 路线
		re := regexp.MustCompile(`[a-zA-Z]`)
		letters := re.FindAllString(parts[1], -1)

		// 狗粮
		DogExpNum := strings.ReplaceAll(parts[2], "狗粮经验", "")
		number, _ := strconv.Atoi(DogExpNum)
		if number == -1 {
			continue
		}

		// 摩拉
		MoraNum := strings.ReplaceAll(parts[3], "摩拉", "")
		number2, _ := strconv.Atoi(MoraNum)
		if number2 == 0 {
			continue

		}

		date := strings.ReplaceAll(parts[0], "日期:", "")
		data.Dates = append(data.Dates, date)
		data.Line = append(data.Line, letters[0])
		data.DogExp = append(data.DogExp, number)
		data.Mora = append(data.Mora, number2)

	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return data, nil
}

// IsStringInDictionaryCategory 检查一个字符串是否包含字典数组中的任何词语
func IsStringInDictionaryCategory(target string, dictionary []string) bool {
	for _, word := range dictionary {
		if strings.Contains(target, word) {
			return true // 如果找到任何一个词语，就返回 true
		}
	}
	return false // 遍历完所有词语都没有找到，则返回 false
}

// 定义一个结构体来存储键值对
type KeyValue struct {
	Key   string
	Value int
}

// 创建一个数组
var Relics = []string{"冒险家", "游医", "幸运儿", "险家", "医的", "运儿", "家",
	"方巾", "枭羽", "怀钟", "药壶", "银莲", "怀表", "尾羽", "头带", "金杯", "之花", "之杯",
	"沙漏", "绿花", "银冠", "鹰羽"}

// analyseLog handles the /api/analyse GET request
func LogAnalysis() map[string]int {
	autoLog.Sugar.Infof("日志分析")
	res, _ := TodayHarvest()

	var datas []KeyValue

	var syw = 0
	var xie = 0

	for item, count := range res {
		var data KeyValue

		if IsStringInDictionaryCategory(item, Relics) {
			syw += count
		} else if strings.Contains(item, "蟹") {
			xie += count
		} else if item == "调查" {
			continue
		} else {
			data.Key = item
			data.Value = count
			//autoLog.Sugar.Infof("物品: %s, 数量: %d", item, count)
		}
		datas = append(datas, data)
	}
	var data KeyValue
	data.Key = "圣遗物"
	data.Value = syw
	datas = append(datas, data)

	var dataXie KeyValue
	dataXie.Key = "螃蟹"
	dataXie.Value = xie
	datas = append(datas, dataXie)

	// 按值从大到小排序
	sort.Slice(datas, func(i, j int) bool {
		return datas[i].Value > datas[j].Value
	})

	// 取出前 5 个元素，考虑长度不足 5 的情况
	mapData := make(map[string]int)
	for i := 0; i < 10 && i < len(datas); i++ {

		mapData[datas[i].Key] = datas[i].Value
	}

	return mapData

}

func FindLogFiles(dirPath string) ([]string, error) {

	date := time.Now().Format("20060102")
	prefix := fmt.Sprintf("better-genshin-impact%s", date)
	pattern := dirPath + "\\" + prefix + "*.log" // logs 为日志目录

	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	var filenames []string
	for _, f := range files {
		filenames = append(filenames, filepath.Base(f))
	}

	return filenames, nil
}

func UpdateJsAndPathing() error {
	autoLog.Sugar.Infof("开始更新脚本和地图仓库")
	autoLog.Sugar.Infof("开始备份user文件夹")

	err4 := zipDir(Config.BetterGIAddress+"\\User\\", "Users\\User"+time.Now().Format("20060102")+".zip")
	if err4 != nil {
		return fmt.Errorf("备份失败")
	}

	autoLog.Sugar.Info("备份成功")

	url := "https://github.com/babalae/bettergi-scripts-list/archive/refs/heads/main.zip"
	zipFile := "main.zip"
	targetPrefix := "repo/"
	outputDir := "repo"
	// 下载 zip 文件
	if err := downloadFile(zipFile, url); err != nil {
		autoLog.Sugar.Info("下载失败")
		return err
	}

	autoLog.Sugar.Info("下载完成")
	// 解压指定目录
	if err := unzipRepo(zipFile, outputDir, targetPrefix); err != nil {
		autoLog.Sugar.Errorf("解压失败")
		return err
	}

	autoLog.Sugar.Info("已提取 repo 文件夹")

	_ = os.Remove(zipFile)

	autoLog.Sugar.Info("已删除压缩包")
	autoLog.Sugar.Info("开始备份指定文件")
	for _, path := range Config.Backups {

		file := fmt.Sprintf("%s\\User\\%s", Config.BetterGIAddress, path)

		err := copy.Copy(file, "./backups/"+path)
		if err != nil {

			autoLog.Sugar.Error("备份文件失败", err)
			return err
		}
		autoLog.Sugar.Info("已复制文件:", path)
	}

	autoLog.Sugar.Info("开始更新脚本文件")
	err := copy.Copy("./repo/js", Config.BetterGIAddress+"\\User\\JsScript")
	if err != nil {
		return err
	}

	autoLog.Sugar.Info("已更新脚本文件")
	autoLog.Sugar.Info("开始更新地图追踪文件")

	err2 := os.RemoveAll(Config.BetterGIAddress + "\\User\\AutoPathing")
	if err2 != nil {
		return err2
	}
	err3 := copy.Copy("./repo/pathing", Config.BetterGIAddress+"\\User\\AutoPathing")
	if err3 != nil {
		return err3
	}

	autoLog.Sugar.Info("开始还原备份文件配置文件")
	autoLog.Sugar.Info("开始还原备份文件配置文件")

	for _, path := range Config.Backups {

		file := fmt.Sprintf("%s\\User\\%s", Config.BetterGIAddress, path)

		err := copy.Copy("./backups/"+path, file)
		if err != nil {
			return err
		}

		autoLog.Sugar.Info("已还原文件", file)
	}

	autoLog.Sugar.Info("还原备份文件配置文件成功")
	os.RemoveAll("./repo")
	autoLog.Sugar.Info("脚本和地图已经更新成功")
	return nil
}

// 解压 zip 中 repo 文件夹的内容
func unzipRepo(zipPath, outputDir, targetPrefix string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	rootPrefix := ""
	if len(r.File) > 0 {
		parts := strings.SplitN(r.File[0].Name, "/", 2)
		if len(parts) > 1 {
			rootPrefix = parts[0] + "/"
		}
	}

	fullTarget := rootPrefix + targetPrefix

	for _, f := range r.File {
		if !strings.HasPrefix(f.Name, fullTarget) {
			continue // 跳过不在 repo/ 下的内容
		}

		relPath := strings.TrimPrefix(f.Name, fullTarget)
		fpath := filepath.Join(outputDir, relPath)

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		outFile, err := os.Create(fpath)
		if err != nil {
			return err
		}
		defer outFile.Close()

		_, err = io.Copy(outFile, rc)
		if err != nil {
			return err
		}
	}

	return nil
}

// 下载文件
func downloadFile(filename, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func zipDir(sourceDir, zipFilePath string) error {
	fmt.Println(sourceDir)
	fmt.Println(zipFilePath)

	// 创建一个新的 zip 文件
	zipFile, err := os.Create(zipFilePath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	// 创建一个新的 zip 写入器
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// 遍历文件夹中的所有文件和子文件夹
	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 使用 filepath.Rel 获取相对路径
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		// 获取文件在压缩包中的相对路径
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// 如果是目录，则不需要写入内容，只需创建对应的目录条目
		if info.IsDir() {
			header.Name = relPath + "/"
			header.Method = zip.Store
			if _, err := zipWriter.CreateHeader(header); err != nil {
				return err
			}
			return nil
		}

		// 文件的相对路径
		header.Name = relPath

		// 设置压缩方法
		header.Method = zip.Deflate

		// 创建新的文件写入器
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		// 打开文件进行读取
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		// 将文件内容复制到压缩包中
		_, err = io.Copy(writer, file)
		return err
	})

	return err
}

func Backup() error {
	for _, path := range Config.Backups {

		file := fmt.Sprintf("%s\\User\\%s", Config.BetterGIAddress, path)

		copy.Copy(file, "./backups/"+path)

		autoLog.Sugar.Infof("已备份文件: %s\n", path)
	}
	autoLog.Sugar.Infof("开始备份user文件夹")
	err4 := zipDir(Config.BetterGIAddress+"\\User\\", "Users\\User"+time.Now().Format("2006100215020405")+".zip")
	if err4 != nil {
		autoLog.Sugar.Errorf("备份失败: %v")
		return fmt.Errorf("备份失败")
	}

	autoLog.Sugar.Info("备份成功")
	return nil
}

type GroupMap struct {
	//标题
	Title  string
	Detail GroupDetail
}

type GroupDetail struct {
	// 开始时间
	StartTime string
	// 结束时间
	EndTime string
	// 执行时间
	ExecuteTime string
	//摩拉
	MoLa int
}

func GroupTime() ([]GroupMap, error) {
	nowTime := time.Now()
	today := nowTime.Format("2006-01-02")
	layoutFull := "2006-01-02 15:04:05"

	date := time.Now().Format("20060102")
	filename := filepath.Clean(fmt.Sprintf("%s\\log\\better-genshin-impact%s.log", Config.BetterGIAddress, date))
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	timeRegexp := regexp.MustCompile(`\[(\d{2}:\d{2}:\d{2}\.\d{3})\]`)
	startRegexp := regexp.MustCompile(`配置组 "(.*?)" 加载完成`)
	endRegexp := regexp.MustCompile(`配置组 "(.*?)" 执行结束`)

	type TempGroup struct {
		GroupName string
		StartTime time.Time
		LineTime  string // 日志时间字符串
	}

	var results []GroupMap
	var temp *TempGroup
	scanner := bufio.NewScanner(file)
	var prevLine string

	var asyncList []config.TravelsDiaryDetailList

	fmt.Println(Config.IsMoLaSum)

	if Config.IsMoLaSum {
		async, err := config.GetTravelsDiaryDetailAsync(int(nowTime.Month()), 2, 1)
		if err == nil {
			time.Sleep(3 * time.Second)
			async1, _ := config.GetTravelsDiaryDetailAsync(int(nowTime.Month()), 2, 2)
			time.Sleep(3 * time.Second)
			async2, _ := config.GetTravelsDiaryDetailAsync(int(nowTime.Month()), 2, 3)

			asyncList = append(async.List, async1.List...)
			asyncList = append(async.List, async2.List...)
		}
	}

	var sunTime time.Duration
	var sumMoLa int

	for scanner.Scan() {
		line := scanner.Text()

		if prevLine != "" {
			// 开始记录
			if startMatch := startRegexp.FindStringSubmatch(line); startMatch != nil {
				if timeMatch := timeRegexp.FindStringSubmatch(prevLine); timeMatch != nil {
					t, _ := time.Parse(layoutFull, today+" "+timeMatch[1])
					temp = &TempGroup{
						GroupName: startMatch[1],
						StartTime: t,
						LineTime:  timeMatch[1],
					}
				}
			}

			// 结束记录
			if endMatch := endRegexp.FindStringSubmatch(line); endMatch != nil && temp != nil && endMatch[1] == temp.GroupName {
				if timeMatch := timeRegexp.FindStringSubmatch(prevLine); timeMatch != nil {
					endTime, _ := time.Parse(layoutFull, today+" "+timeMatch[1])
					duration := endTime.Sub(temp.StartTime)

					sunTime += duration

					// 过滤收益
					startStr := temp.StartTime.Format("2006-01-02 15:04:05")
					endStr := endTime.Format("2006-01-02 15:04:05")
					filtered := config.FilterByTime(asyncList, startStr, endStr)
					var totalMoLa int
					for _, item := range filtered {
						totalMoLa += item.Num
					}
					sumMoLa += totalMoLa

					// 组装
					results = append(results, GroupMap{
						Title: temp.GroupName,
						Detail: GroupDetail{
							StartTime:   startStr,
							EndTime:     endStr,
							ExecuteTime: duration.String(),
							MoLa:        totalMoLa,
						},
					})

					// 重置临时变量
					temp = nil
				}
			}
		}
		prevLine = line
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// 计算总时长
	results = append(results, GroupMap{
		Title: "合计",
		Detail: GroupDetail{
			StartTime:   "00:00:00",
			EndTime:     "00:00:00",
			ExecuteTime: sunTime.String(),
			MoLa:        sumMoLa,
		},
	})

	return results, nil
}

// 判断配置文件是否正确
func CheckConfig() (bool, error) {
	_, err := os.Stat(Config.BetterGIAddress)
	if err == nil {
		fmt.Println("Bgi安装目录设置正确")
	}
	if os.IsNotExist(err) {
		return false, fmt.Errorf("Bgi安装目录设置错误目录设置错误，请检查配置文件BetterGIAddress：例子：D:\\subject\\BetterGI")
	}
	names := Config.ConfigNames
	if len(names) == 7 {
		fmt.Println("配置组configNames正确")
	} else {
		return false, fmt.Errorf("配置组configNames不正确")
	}
	return true, nil
}
