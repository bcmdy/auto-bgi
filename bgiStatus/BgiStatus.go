package bgiStatus

import (
	"archive/zip"
	"auto-bgi/Notice"
	"auto-bgi/ScriptRepo"
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"auto-bgi/control"
	"auto-bgi/tools"
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/otiai10/copy"
	"github.com/robfig/cron/v3"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// 检查 程序 是否在运行
func IsWechatRunning(name string) bool {
	cmd := exec.Command("tasklist", "/FI", "IMAGENAME eq "+name)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := cmd.Output()
	if err != nil {

		autoLog.Sugar.Error("BetterGI.exe 是否在运行:", err)
		return false
	}

	return strings.Contains(string(output), "BetterGI.exe")
}

// 检查 YuanShen.exe 是否在运行
//func IsYuanShenRunning() bool {
//	cmd := exec.Command("tasklist", "/FI", "IMAGENAME eq YuanShen.exe")
//	output, err := cmd.Output()
//	if err != nil {
//
//		autoLog.Sugar.Error("YuanShen.exe 是否在运行:", err)
//		return false
//	}
//
//	return strings.Contains(string(output), "YuanShen.exe")
//}

//var notified = false
//var okInform = false
//var okRun = true
//
//func CheckBetterGIStatus() {
//
//	cronTab := cron.New(cron.WithSeconds())
//
//	// 定时任务,cron表达式
//	spec := "*/30 * * * * *"
//
//	task := func() {
//
//		// 检查进程
//		if IsWechatRunning() {
//
//			if okRun {
//				autoLog.Sugar.Infof("BetterGI 正在运行: %s", time.Now().Format("2006-01-02 15:04:05"))
//				BgiLogStatusInfo.Running = IsWechatRunning()
//				notified = false // 清除通知状态
//				okRun = false    // 清除通知状态
//			}
//		} else {
//			if !notified {
//				Notice.SentText("BetterGI 已经关闭:" + config.Cfg.Content)
//				control.CloseYuanShen()
//				notified = true
//				okRun = true
//				BgiLogStatusInfo.Running = IsWechatRunning()
//			} else if !okInform {
//				autoLog.Sugar.Infof("BetterGI 已关闭，已通知过: %s", time.Now().Format("2006-01-02 15:04:05"))
//				okInform = true
//				BgiLogStatusInfo.Running = IsWechatRunning()
//			}
//		}
//
//	}
//
//	// 添加定时任务
//	cronTab.AddFunc(spec, task)
//	// 启动定时器
//	cronTab.Start()
//	// 阻塞主线程停止
//	select {}
//}

var lastRunning = true

func CheckBetterGIStatus() {
	cronTab := cron.New(cron.WithSeconds())
	spec := "*/50 * * * * *"

	task := func() {
		running := IsWechatRunning("BetterGI.exe")
		if running != lastRunning {
			if running {
				autoLog.Sugar.Infof("BetterGI 正在运行: %s", time.Now().Format("2006-01-02 15:04:05"))
			} else {
				Notice.SentText("BetterGI 已经关闭:" + config.Cfg.Content)
				control.CloseYuanShen()
				autoLog.Sugar.Infof("BetterGI 已关闭: %s", time.Now().Format("2006-01-02 15:04:05"))
			}
			lastRunning = running
		}
		BgiLogStatusInfo.Running = running
	}

	if _, err := cronTab.AddFunc(spec, task); err != nil {
		autoLog.Sugar.Errorf("添加定时任务失败: %v", err)
		return
	}

	cronTab.Start()
	select {}
}

func JsProgress(filename string, patterns ...string) (string, error) {
	// 编译所有的正则表达式
	var regexps []*regexp.Regexp
	for _, p := range patterns {
		re, err := regexp.Compile(p)
		if err != nil {
			return "", fmt.Errorf("正则表达式编译失败: %v", err)
		}
		regexps = append(regexps, re)
	}

	// 打开文件
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// 扫描文件行并尝试匹配所有正则表达式
	var lastMatch string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		for _, re := range regexps {
			if re.MatchString(line) {
				lastMatch = line
				break // 当前行已经匹配，继续下一行
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	// 返回最后一行匹配结果
	if lastMatch != "" {
		return lastMatch, nil
	}
	return "", fmt.Errorf("没有找到匹配的行")
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
		return "", fmt.Errorf("进度读取文件失败:%s", filename)
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

func TodayHarvest(fileName string) (map[string]int, error) {

	autoLog.Sugar.Infof("今日收获统计")
	re := regexp.MustCompile(`^交互或拾取："([^"]*)"`)

	filename := filepath.Clean(fmt.Sprintf("%s\\log\\%s", config.Cfg.BetterGIAddress, fileName))

	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	// 初始化map用于存储物品和出现次数
	harvestStats := make(map[string]int)

	scanner := bufio.NewScanner(file)
	var jsonData strings.Builder
	var readingJSON bool

	for scanner.Scan() {
		line := scanner.Text()

		// 第一种匹配（交互或拾取）
		if matches := re.FindAllStringSubmatch(line, -1); matches != nil {
			for _, match := range matches {
				if len(match) > 1 {
					item := match[1]
					harvestStats[item]++
				}
			}
			continue // 匹配到后可跳过第二种，以免重复
		}

		// ✅ 检测到 “[主流程]总累积获取” 段落开始
		if strings.HasPrefix(line, "[主流程]总累积获取") {
			readingJSON = true
			jsonData.Reset() // 清空旧内容
			idx := strings.Index(line, "{")
			if idx != -1 {
				jsonData.WriteString(line[idx:]) // 把 { 后面部分写入
			}
			continue
		}

		if readingJSON {
			jsonData.WriteString(line)
			// 检测 JSON 结束
			if strings.HasSuffix(line, "}") {
				readingJSON = false
				// 尝试解析 JSON
				data := make(map[string]int)
				if err := json.Unmarshal([]byte(jsonData.String()), &data); err != nil {
					autoLog.Sugar.Warnf("解析总累积获取 JSON 失败: %v", err)
				} else {
					// 合并到 harvestStats
					for k, v := range data {
						harvestStats[k] += v
					}
				}
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
	filename := filepath.Clean(fmt.Sprintf("%s\\User\\JsScript\\背包材料统计\\latest_record.txt", config.Cfg.BetterGIAddress))

	// 打开文件
	file, err := os.Open(filename)
	if err != nil {
		autoLog.Sugar.Errorf("背包统计失败: %v", err)
	}
	defer file.Close()

	// 创建一个扫描器来读取文件
	scanner := bufio.NewScanner(file)

	// 创建一个正则表达式来匹配日期格式 "YYYY/M/D HH:MM:SS"
	re1 := regexp.MustCompile(`\b\d{4}/\d{1,2}/\d{1,2} \d{2}:\d{2}:\d{2}\b`)

	statistics := config.Cfg.BagStatistics

	split := strings.Split(statistics, ",")

	var bags []Material
	var bag Material

	bagMap := make(map[string]Material)

	layout := "2006/1/2 15:04:05"

	for scanner.Scan() {
		for _, s := range split {
			// 创建一个正则表达式来匹配 "晶蝶：数字" 模式
			sprintf := fmt.Sprintf(`(?:^|[,\s])%s: (\d+)`, s)

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
				bag.Cl = strings.Replace(split[0], ",", "", -1)
				bag.Num = split[1]

				//判断是否已经有了
				isNil := bagMap[bag.Cl]
				if isNil.Data != "" {
					//判断是否是同一天
					time1, err1 := time.Parse(layout, isNil.Data)
					time2, err2 := time.Parse(layout, bag.Data)
					if err1 != nil || err2 != nil {
						continue
					}

					y1, m1, d1 := time1.Date()
					y2, m2, d2 := time2.Date()

					fmt.Println("======", y1, m1, d1, bag.Cl)
					fmt.Println("======", y1, m1, d1, bag.Cl)

					if y1 == y2 && m1 == m2 && d1 == d2 {
						continue
					} else {
						bagMap[bag.Cl] = bag
						bags = append(bags, bag)
					}

				} else {
					bagMap[bag.Cl] = bag
					bags = append(bags, bag)
				}

			}
		}

		// 检查扫描器是否有错误
		if err := scanner.Err(); err != nil {
			panic(err)
		}
	}

	//摩拉统计
	morasStatistics, _ := MorasStatistics()
	bags = append(bags, morasStatistics...)

	//原石统计
	yuanShiStatistics, _ := YuanShiStatistics()
	bags = append(bags, yuanShiStatistics...)

	return bags, nil
}

func CheckBag() map[string]int {
	autoLog.Sugar.Infof("背包检查")
	filename := filepath.Clean(fmt.Sprintf("%s\\User\\JsScript\\背包材料统计\\latest_record.txt", config.Cfg.BetterGIAddress))

	file, err := os.Open(filename)
	if err != nil {
		autoLog.Sugar.Errorf("没有相关JS:背包材料统计")
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// 正则匹配物品名和数量
	itemRegex := regexp.MustCompile(`([\p{Han}「」·A-Za-z0-9]+):\s*([0-9?]+)`)

	allItems := make(map[string]int)

	// 遍历整个文件，逐行解析
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, ":") {
			matches := itemRegex.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				name := match[1]
				qtyStr := match[2]
				if qtyStr == "?" {
					continue
				}
				//判断是否已经有了
				isNil := allItems[name]
				if isNil != 0 {
					continue
				}

				qty, _ := strconv.Atoi(qtyStr)
				allItems[name] = qty
			}
		}
	}

	liquidationItems := make(map[string]int)

	// 检查是否有物品数量超过8000
	for item, qty := range allItems {
		if qty > 8000 {
			liquidationItems[item] = qty
		}
	}

	return liquidationItems

}

// 原石统计
func YuanShiStatistics() ([]Material, error) {
	autoLog.Sugar.Infof("原石统计")
	filename := filepath.Clean(fmt.Sprintf("%s\\User\\JsScript\\OCR读取当前抽卡资源并发送通知\\Resources_log.txt", config.Cfg.BetterGIAddress))
	file, err := os.Open(filename)
	if err != nil {
		autoLog.Sugar.Errorf("没有相关JS:OCR读取当前抽卡资源并发送通知")
		return nil, err
	}
	defer file.Close()
	var bags []Material
	// 创建一个扫描器来读取文件
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var bag Material
		line := scanner.Text()
		split := strings.Split(line, " —— ")
		if len(split) < 4 {
			continue
		}
		bag.Data = split[0]

		bag.Cl = "原石"

		yuanShiNum := split[3]
		//提取数字
		re := regexp.MustCompile(`\d+`)
		num := re.FindString(yuanShiNum)
		bag.Num = num

		bags = append(bags, bag)
	}
	return bags, nil
}

// 摩拉统计
func MorasStatistics() ([]Material, error) {

	autoLog.Sugar.Infof("摩拉统计")
	filename := filepath.Clean(fmt.Sprintf("%s\\User\\JsScript\\OCR读取当前摩拉记录并发送通知\\mora_log.txt", config.Cfg.BetterGIAddress))
	// 打开文件
	file, err := os.Open(filename)
	if err != nil {
		autoLog.Sugar.Infof("没有相关JS")
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

	autoLog.Sugar.Infof("清理背包统计")
	DeleteBag()

	autoLog.Sugar.Infof("清理摩拉统计")
	DeleteMoLa()

	autoLog.Sugar.Infof("清理原石统计")
	DeleteYuanShi()

	autoLog.Sugar.Infof("清理成功")
	return "清理成功"
}

type DogFood struct {
	FileName string
	Mark     string
	Detail   []string
}

// 获取当前配置组
func FindLastGroup(filename string) (group string, timestamp string, err error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var prevLine string
	for scanner.Scan() {
		line := scanner.Text()
		// 拼接上一行和当前行
		combined := prevLine + " " + line

		// 正则匹配时间和配置组
		pattern := `\[(\d{2}:\d{2}:\d{2}\.\d{3})\]\s+\[INF\].*?配置组 "(.*?)" 加载完成，共\d+个脚本，开始执行`
		re := regexp.MustCompile(pattern)

		matches := re.FindStringSubmatch(combined)
		if matches != nil {
			timestamp = matches[1]
			group = matches[2]
		}

		prevLine = line
	}

	if err := scanner.Err(); err != nil {
		return "", "", err
	}

	if group == "" {
		return "", "", fmt.Errorf("没有找到匹配的行")
	}

	return group, timestamp, nil
}

// 获取配置组进度
func GetGroupP(group string) string {
	file, err := os.Open("OneLongTask.txt")
	if err != nil {
		fmt.Println("打开文件失败：", err)
		return "未知"
	}
	defer file.Close()
	totalLines := 0
	gouliangLines := 0

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		totalLines++
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, group) {
			gouliangLines = totalLines
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("读取文件出错：", err)
		return "未知"
	}

	return fmt.Sprintf("%d/%d", gouliangLines, totalLines)
}

// 读取manifest.json的version号
func ReadVersion(filePath string) string {
	// 打开文件
	Path := filepath.Join(filePath, "manifest.json")
	file, err := os.Open(Path)
	if err != nil {
		fmt.Println("打开文件失败:", err)
	}
	defer file.Close()
	// 文件内容转map
	var data map[string]interface{}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&data)
	if err != nil {
		return "未知版本"
	}
	// 获取version
	version, ok := data["version"].(string)
	if !ok {
		return "未知版本"
	}
	return version

}

func GetAutoArtifactsPro() ([]DogFood, error) {
	// 获取当前目录下所有 .txt 文件
	files, err := filepath.Glob(fmt.Sprintf("%s\\User\\JsScript\\AAA-Artifacts-Bulk-Supply\\records\\*.txt", config.Cfg.BetterGIAddress))
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
				if strings.Contains(line, "上次运行收尾路线") {
					replace := strings.ReplaceAll(line, "上次运行收尾路线：", "")
					dogFood.Mark = replace

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
	filePath := filepath.Clean(fmt.Sprintf("%s\\User\\JsScript\\AAA-Artifacts-Bulk-Supply\\records\\%s", config.Cfg.BetterGIAddress, fileName))
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
			if strings.Contains(line, "上次运行收尾路线") {
				inHistory = true
			}
			continue
		}
		// 1. 分割字符串，获取日期部分
		parts := strings.Split(line, "，")
		if len(parts) != 4 {
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
		if number <= -1 {
			number = 0
		}

		// 摩拉
		MoraNum := strings.ReplaceAll(parts[3], "摩拉", "")
		number2, _ := strconv.Atoi(MoraNum)
		if number2 <= -1 {
			number2 = 0

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
	"沙漏", "绿花", "银冠", "鹰羽", "冒险", "游医的", "教官", "战狂", "流放"}

// analyseLog handles the /api/analyse GET request
func LogAnalysis(fileName string) map[string]int {
	autoLog.Sugar.Infof("日志分析")
	res, _ := TodayHarvest(fileName)

	var datas []KeyValue

	var syw = 0
	var xie = 0

	for item, count := range res {
		var data KeyValue

		if IsStringInDictionaryCategory(item, Relics) {
			syw += count
		} else if strings.Contains(item, "蟹") {
			xie += count
		} else if item == "调查" || item == "周查" {
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
	pattern := filepath.Join(dirPath, "*.log")

	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	// 保存文件名和时间
	type fileInfo struct {
		name string
		time time.Time
	}

	var fileInfos []fileInfo
	for _, f := range files {
		info, err := os.Stat(f)
		if err != nil {
			continue // 读取失败跳过
		}
		fileInfos = append(fileInfos, fileInfo{
			name: filepath.Base(f),
			time: info.ModTime(),
		})
	}

	// 按时间倒序排序
	sort.Slice(fileInfos, func(i, j int) bool {
		return fileInfos[i].time.After(fileInfos[j].time)
	})

	// 只返回文件名
	var filenames []string
	for _, fi := range fileInfos {
		filenames = append(filenames, fi.name)
	}

	return filenames, nil
}

func FindLogFiles1Remote(dirPath string) ([]string, error) {
	// 匹配 1Remote.log_20250810.md 这种文件名
	pattern := filepath.Join(dirPath, "1Remote.log_*.md")

	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	// 保存文件名和修改时间
	type fileInfo struct {
		name string
		time time.Time
	}

	var fileInfos []fileInfo
	for _, f := range files {
		info, err := os.Stat(f)
		if err != nil {
			continue // 读取失败跳过
		}
		fileInfos = append(fileInfos, fileInfo{
			name: filepath.Base(f), // 只保存文件名
			time: info.ModTime(),
		})
	}

	// 按修改时间倒序排序（最新的在前）
	sort.Slice(fileInfos, func(i, j int) bool {
		return fileInfos[i].time.After(fileInfos[j].time)
	})

	// 只返回文件名
	var filenames []string
	for _, fi := range fileInfos {
		filenames = append(filenames, fi.name)
	}

	return filenames, nil
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

// zipDir 压缩 sourceDir 到 zipFilePath
// keepRoot = true 时会在压缩包中保留 sourceDir 的目录名
func ZipDir(sourceDir, zipFilePath string, keepRoot bool) error {

	//清理历史备份
	_ = ClearDir("Users")

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

func Backup() error {

	autoLog.Sugar.Infof("开始备份user文件夹")
	err4 := ZipDir(config.Cfg.BetterGIAddress+"\\User\\", "Users\\User"+time.Now().Format("2006100215020405")+".zip", true)
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
}

// 提取文件名字日期
func GetFileNameDate(filePath string) string {

	fileName := filepath.Base(filePath)

	//提取文件名字的数字
	// 正则表达式匹配数字
	re := regexp.MustCompile(`\d+`)
	// 查找所有匹配项
	matches := re.FindAllString(fileName, -1)
	// 检查是否找到匹配项
	if len(matches) > 0 {
		//格式化转换
		formatted := matches[0][:4] + "-" + matches[0][4:6] + "-" + matches[0][6:]

		return formatted
	}
	return ""
}

// 日志分析
func GroupTime(fileName string) []LogAnalysis2Struct {
	filePath := filepath.Join(config.Cfg.BetterGIAddress, "log")
	fullPath := filepath.Join(filePath, fileName)
	date := GetFileNameDate(fileName)

	file, err := os.Open(fullPath)
	if err != nil {
		fmt.Println("无法打开日志文件:", err)
		return nil
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	// 🔹正则提前编译（避免循环里重复编译）
	startRegexp := regexp.MustCompile(`配置组 "(.*?)" 加载完成`)
	endRegexp := regexp.MustCompile(`配置组 "(.*?)" 执行结束`)

	var logAnalysis2Structs []LogAnalysis2Struct
	var currentStruct *LogAnalysis2Struct
	var lastLine string

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				autoLog.Sugar.Infof("分析完毕")
				// 🔹文件结束时检查是否有未结束的配置组
				if currentStruct != nil {
					logAnalysis2Structs = append(logAnalysis2Structs, *currentStruct)
				}
				break
			}
			autoLog.Sugar.Errorf("配置组分析文件出错: %v", err)
			break
		}

		// 默认时间戳
		timestampLine := lastLine
		if tools.HasTimestamp(line) {
			timestampLine = line
		}
		if timestampLine == "" {
			// 防止开头无时间戳导致空值
			timestampLine = date + " 00:00:00"
		}

		// 配置组开始
		if matches := startRegexp.FindStringSubmatch(line); len(matches) > 1 {
			currentStruct = &LogAnalysis2Struct{
				GroupName:    matches[1],
				ErrorSummary: make(map[string]int),
				SumIncome:    make(map[string]int),
			}
			if t, err := tools.ExtractLogTime2(date, timestampLine); err == nil {
				seg := ExecutionSegment{StartTime: t}
				currentStruct.Segments = append(currentStruct.Segments, seg)
			}
		}

		// 配置组结束
		if currentStruct != nil {
			if matches := endRegexp.FindStringSubmatch(line); len(matches) > 1 && matches[1] == currentStruct.GroupName {
				if t, err := tools.ExtractLogTime2(date, timestampLine); err == nil {
					segIdx := len(currentStruct.Segments) - 1
					currentStruct.Segments[segIdx].EndTime = t
					currentStruct.Segments[segIdx].Consuming =
						tools.CalculateDuration(currentStruct.Segments[segIdx].StartTime, t)
				}

				// 🔹统计错误（按子任务聚合）
				for _, subTask := range currentStruct.LogAnalysis2Json {
					for errStr, count := range subTask.Errors {
						currentStruct.ErrorSummary[errStr] += count
					}
				}

				logAnalysis2Structs = append(logAnalysis2Structs, *currentStruct)
				currentStruct = nil
			}
		}

		lastLine = line
	}

	// 🔹合并同名配置组（分段执行支持）
	merged := make(map[string]*LogAnalysis2Struct)
	for _, s := range logAnalysis2Structs {
		m, ok := merged[s.GroupName]
		if !ok {
			copyS := s
			if copyS.ErrorSummary == nil {
				copyS.ErrorSummary = make(map[string]int)
			}
			if copyS.SumIncome == nil {
				copyS.SumIncome = make(map[string]int)
			}
			merged[s.GroupName] = &copyS
			continue
		}

		// 🔹追加分段
		m.Segments = append(m.Segments, s.Segments...)

	}

	result := make([]LogAnalysis2Struct, 0, len(merged))
	for _, v := range merged {
		result = append(result, *v)
	}

	return result
}

//func GroupTime(fileName string) ([]GroupMap, error) {
//	layoutFull := "2006-01-02 15:04:05"
//
//	today := time.Now().Format("2006-01-02")
//
//	//提取文件名字的数字
//	// 正则表达式匹配数字
//	re := regexp.MustCompile(`\d+`)
//	// 查找所有匹配项
//	matches := re.FindAllString(fileName, -1)
//	// 检查是否找到匹配项
//	if len(matches) > 0 {
//		//格式化转换
//		formatted := matches[0][:4] + "-" + matches[0][4:6] + "-" + matches[0][6:]
//
//		today = formatted
//	}
//
//	filename := filepath.Clean(fmt.Sprintf("%s\\log\\%s", config.Cfg.BetterGIAddress, fileName))
//
//	file, err := os.Open(filename)
//	if err != nil {
//		return nil, err
//	}
//	defer file.Close()
//
//	timeRegexp := regexp.MustCompile(`\[(\d{2}:\d{2}:\d{2}\.\d{3})\]`)
//	startRegexp := regexp.MustCompile(`配置组 "(.*?)" 加载完成`)
//	endRegexp := regexp.MustCompile(`配置组 "(.*?)" 执行结束`)
//
//	type TempGroup struct {
//		GroupName string
//		StartTime time.Time
//		LineTime  string // 日志时间字符串
//	}
//
//	var results []GroupMap
//	var temp *TempGroup
//	scanner := bufio.NewScanner(file)
//	var prevLine string
//
//	var sunTime time.Duration
//
//	for scanner.Scan() {
//		line := scanner.Text()
//
//		if prevLine != "" {
//			// 开始记录
//			if startMatch := startRegexp.FindStringSubmatch(line); startMatch != nil {
//				if timeMatch := timeRegexp.FindStringSubmatch(prevLine); timeMatch != nil {
//					t, _ := time.Parse(layoutFull, today+" "+timeMatch[1])
//					temp = &TempGroup{
//						GroupName: startMatch[1],
//						StartTime: t,
//						LineTime:  timeMatch[1],
//					}
//				}
//			}
//
//			// 结束记录
//			if endMatch := endRegexp.FindStringSubmatch(line); endMatch != nil && temp != nil && endMatch[1] == temp.GroupName {
//				if timeMatch := timeRegexp.FindStringSubmatch(prevLine); timeMatch != nil {
//					endTime, _ := time.Parse(layoutFull, today+" "+timeMatch[1])
//					duration := endTime.Sub(temp.StartTime)
//
//					sunTime += duration
//
//					// 过滤收益
//					startStr := temp.StartTime.Format("2006-01-02 15:04:05")
//					endStr := endTime.Format("2006-01-02 15:04:05")
//
//					// 组装
//					results = append(results, GroupMap{
//						Title: temp.GroupName,
//						Detail: GroupDetail{
//							StartTime:   startStr,
//							EndTime:     endStr,
//							ExecuteTime: duration.String(),
//						},
//					})
//
//					// 重置临时变量
//					temp = nil
//				}
//			}
//		}
//		prevLine = line
//	}
//
//	if err := scanner.Err(); err != nil {
//		return nil, err
//	}
//
//	// 计算总时长
//	results = append(results, GroupMap{
//		Title: "合计",
//		Detail: GroupDetail{
//			StartTime:   "00:00:00",
//			EndTime:     "00:00:00",
//			ExecuteTime: sunTime.String(),
//		},
//	})
//
//	return results, nil
//}

// 判断配置文件是否正确
func CheckConfig() (bool, error) {
	fmt.Println("配置文件路径", config.Cfg.BetterGIAddress)
	_, err := os.Stat(config.Cfg.BetterGIAddress)
	if err == nil {
		fmt.Println("Bgi安装目录设置正确")
	}
	if os.IsNotExist(err) {
		return false, fmt.Errorf("Bgi安装目录设置错误目录设置错误，请检查配置文件BetterGIAddress：你有没有加双斜杠呀，没有看网站说明")
	}
	names := config.Cfg.ConfigNames
	if len(names) == 7 {
		fmt.Println("配置组configNames正确")
	} else {
		return false, fmt.Errorf("配置组configNames不正确")
	}
	return true, nil
}

func GetGroupPInfo() string {

	//读取文件内容
	file := "OneLongTask.txt"

	openFile, _ := os.OpenFile(file, os.O_RDWR, os.ModePerm)

	stat, _ := openFile.Stat()
	if stat == nil {
		return ""
	}

	defer openFile.Close()

	reader := bufio.NewReader(openFile)

	//读取
	s1 := make([]byte, stat.Size())
	_, err := reader.Read(s1)
	if err != nil {
		return ""
	}

	return string(s1)
}

type GitLogStruct struct {
	//提交时间
	CommitTime string
	//作者
	Author string
	//更新内容
	Message string
	//提交修改的文件
	Files []string
}

// 查询git日志
func GitLog(n int) ([]GitLogStruct, error) {
	localPath := config.Cfg.BetterGIAddress + "/Repos/bettergi-scripts-list-git"

	// 打开仓库
	r, err := git.PlainOpen(localPath)
	if err != nil {
		return nil, fmt.Errorf("打开仓库失败: %w", err)
	}

	// 获取 HEAD
	ref, err := r.Head()
	if err != nil {
		return nil, fmt.Errorf("获取 HEAD 失败: %w", err)
	}

	// 遍历日志
	cIter, err := r.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return nil, fmt.Errorf("获取日志失败: %w", err)
	}

	var logs []GitLogStruct
	count := 0
	err = cIter.ForEach(func(c *object.Commit) error {
		var files []string

		// 获取父提交，比较差异
		if c.NumParents() > 0 {
			parent, err := c.Parents().Next()
			if err == nil {
				patch, err := parent.Patch(c)
				if err == nil {
					for _, stat := range patch.Stats() {
						files = append(files, stat.Name)
					}
				}
			}
		} else {
			// 初始提交：直接获取树对象所有文件
			tree, _ := c.Tree()
			_ = tree.Files().ForEach(func(f *object.File) error {
				files = append(files, f.Name)
				return nil
			})
		}

		logs = append(logs, GitLogStruct{
			CommitTime: c.Author.When.Format("2006-01-02 15:04:05"),
			Author:     c.Author.Name,
			Message:    c.Message,
			Files:      files,
		})

		count++
		if count >= n {
			return storer.ErrStop
		}
		return nil
	})

	if err != nil && err != storer.ErrStop {
		return nil, fmt.Errorf("遍历提交日志失败: %w", err)
	}

	return logs, nil
}

// git拉取代码
func GitPull() {
	// 从配置文件中获取仓库URL
	repoUrl := config.Cfg.RepoUrl
	if repoUrl == "" {
		repoUrl = "https://gitcode.com/huiyadanli/bettergi-scripts-list.git"
	}
	//_, _, err := ScriptRepo.UpdateCenterRepoByGit("https://github.com/babalae/bettergi-scripts-list.git")
	_, _, err := ScriptRepo.UpdateCenterRepoByGit(repoUrl)
	if err != nil {
		if strings.Contains(err.Error(), "worktree contains unstaged changes") {
			autoLog.Sugar.Info("仓库没有更新")
		} else {
			autoLog.Sugar.Errorf("仓库更新失败,请去bgi重置仓库:%s", err.Error())
		}
	}

	//localPath := config.Cfg.BetterGIAddress + "/Repos/bettergi-scripts-list-git"
	//
	//// 尝试打开本地仓库
	//repo, err := git.PlainOpen(localPath)
	//if err == git.ErrRepositoryNotExists {
	//	// 本地不存在，克隆
	//	autoLog.Sugar.Info("仓库不存在，请先去bgi重置或者更新仓库")
	//
	//} else if err == nil {
	//	// 已存在，拉取最新
	//	autoLog.Sugar.Info("仓库存在，拉取最新代码...")
	//	w, err := repo.Worktree()
	//	if err != nil {
	//		return fmt.Errorf("获取工作区失败: %v", err)
	//	}
	//	// 强制还原本地更改
	//	err = w.Reset(&git.ResetOptions{
	//		Mode: git.HardReset,
	//	})
	//	if err != nil {
	//		autoLog.Sugar.Errorf("重置工作区失败: %v", err)
	//		return fmt.Errorf("重置工作区失败: %v", err)
	//	}
	//	autoLog.Sugar.Info("本地更改已清除，准备拉取")
	//
	//	// 拉取更新
	//	err = w.Pull(&git.PullOptions{
	//		RemoteName:    "origin",
	//		ReferenceName: plumbing.NewBranchReferenceName("main"),
	//		Force:         false,
	//	})
	//	if err != nil && err != git.NoErrAlreadyUpToDate {
	//		autoLog.Sugar.Errorf("拉取失败: %v", err)
	//
	//		return fmt.Errorf("拉取失败: %v", err)
	//	}
	//	autoLog.Sugar.Info("拉取完成或已是最新")
	//} else {
	//	return fmt.Errorf("打开仓库失败: %v", err)
	//}
	//return nil
}

func UpdateJs(jsName string) (string, error) {

	repoDir := filepath.Join(config.Cfg.BetterGIAddress, "Repos", "bettergi-scripts-list-git", "repo", "js")

	// 仓库中 js 脚本目录
	subFolderPath, err := findSubFolder(repoDir, jsName)
	if err != nil {
		autoLog.Sugar.Errorf("查找子文件夹失败: %v", err)
		return fmt.Sprintf("未找到子文件夹: %s", jsName), err
	}

	// 本地 js 脚本目录
	targetPath := filepath.Join(config.Cfg.BetterGIAddress, "User", "JsScript", jsName)

	// manifest 中指定的待备份文件或目录
	manifest, err := config.ReadManifest(subFolderPath)
	if err != nil {
		return err.Error(), err
	}
	files := manifest.SavedFiles

	// 备份路径
	backupRoot := filepath.Join("backups", jsName)

	// 开始备份
	for _, pattern := range files {
		fullPattern := filepath.Join(targetPath, pattern)
		matches, err := filepath.Glob(fullPattern)
		if err != nil {
			autoLog.Sugar.Warnf("路径匹配失败: %s, 错误: %v", fullPattern, err)
			continue
		}

		for _, match := range matches {
			relPath, _ := filepath.Rel(targetPath, match)
			dstPath := filepath.Join(backupRoot, relPath)

			err := copy.Copy(match, dstPath)
			if err != nil {
				autoLog.Sugar.Warnf("备份失败: %s -> %s, 错误: %v", match, dstPath, err)
			} else {
				autoLog.Sugar.Infof("备份成功: %s -> %s", match, dstPath)
			}
		}
	}

	// 删除原 js 脚本目录
	os.RemoveAll(targetPath)

	// 拷贝更新的 js 脚本目录
	err = copy.Copy(subFolderPath, targetPath)
	if err != nil {
		return err.Error(), err
	}

	// 4. 还原备份内容到新脚本目录
	for _, pattern := range files {
		backupPattern := filepath.Join(backupRoot, pattern)
		matches, err := filepath.Glob(backupPattern)
		if err != nil {
			autoLog.Sugar.Warnf("还原匹配失败: %s, 错误: %v", backupPattern, err)
			continue
		}

		for _, backupItem := range matches {
			relPath, _ := filepath.Rel(backupRoot, backupItem)
			restorePath := filepath.Join(targetPath, relPath)

			_ = os.MkdirAll(filepath.Dir(restorePath), os.ModePerm)

			if err := copy.Copy(backupItem, restorePath); err != nil {
				autoLog.Sugar.Warnf("还原失败: %s -> %s, 错误: %v", backupItem, restorePath, err)
			} else {
				autoLog.Sugar.Infof("还原成功: %s -> %s", backupItem, restorePath)
			}
		}
	}

	autoLog.Sugar.Infof("Js脚本: %s 已更新并还原备份内容", jsName)
	return "更新并还原成功", nil
}

// 查找 repo 目录下是否存在名为 targetFolder 的子文件夹
func findSubFolder(root string, targetFolder string) (string, error) {
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

func Archive(data LogAnalysis2Struct) string {
	title := data.GroupName
	//合并时间
	executeTime, _ := time.ParseDuration("0s")
	for _, segment := range data.Segments {
		if segment.Consuming != "" {
			duration, err := time.ParseDuration(segment.Consuming)
			if err != nil {
				autoLog.Sugar.Errorf("解析时间失败: %v", err)
				continue
			}
			executeTime += duration

		}
	}

	//如果时间小于30s就不归档
	if executeTime.Seconds() < 30 {
		autoLog.Sugar.Infof("%s执行时间小于30s，不归档", title)
		return "执行时间小于30s，不归档"
	}

	if title == "" || executeTime.String() == "" || executeTime.String() == "0s" {
		autoLog.Sugar.Errorf("归档数据字段缺失或格式错误: %s, %s", title, executeTime.String())
		return "归档数据字段缺失或格式错误"
	}

	// 检查是否已经归档
	stmt, err := config.DB.Prepare(`SELECT COUNT(*) FROM archive_records WHERE title = ?`)
	if err != nil {
		fmt.Println("预处理失败:", err)
		return "预处理失败"
	}
	defer stmt.Close()

	var count int
	err = stmt.QueryRow(title).Scan(&count)
	if err != nil {
		fmt.Println("查询数据库失败:", err)
		return "查询数据库失败"
	}

	autoLog.Sugar.Infof("查询数据库是否存在归档记录：%d", count)

	if count > 0 {
		autoLog.Sugar.Infof("存在归档记录，执行删除操作")

		// 删除已存在的归档记录
		delStmt, err := config.DB.Prepare(`DELETE FROM archive_records WHERE title = ?`)
		if err != nil {
			autoLog.Sugar.Errorf("删除预处理失败: %v", err)
			return "删除预处理失败"
		}
		defer delStmt.Close()

		_, err = delStmt.Exec(title)
		if err != nil {
			autoLog.Sugar.Errorf("删除数据库记录失败: %v", err)
			return "删除数据库记录失败"
		}

		autoLog.Sugar.Infof("删除归档记录成功")
	}

	autoLog.Sugar.Infof("执行新增归档记录")

	// 插入新归档记录
	insertStmt, err := config.DB.Prepare(`INSERT INTO archive_records(title, execute_time) VALUES (?, ?)`)
	if err != nil {
		fmt.Println("预处理失败:", err)
		return "预处理失败"
	}
	defer insertStmt.Close()

	_, err = insertStmt.Exec(title, executeTime.String())
	if err != nil {
		autoLog.Sugar.Errorf("写入数据库失败: %v", err)
		return "写入数据库失败"
	}

	autoLog.Sugar.Infof("成功归档：%s (%s)", title, executeTime)
	return "归档成功"
}

type ArchiveRecords struct {
	Id          int    `json:"id"`
	Title       string `json:"title"`
	ExecuteTime string `json:"execute_time"`
	CreatedAt   string `json:"created_at"`
}

// 时间计算
func CalculateTime(filename, groupName, startTime string) (string, error) {
	// 解析文件名中的日期
	fileDate := GetFileNameDate(filename)

	// 查询数据库配置组时长
	stmt, err := config.DB.Prepare(`SELECT execute_time FROM archive_records WHERE title = ?`)
	if err != nil {
		return "", err
	}
	defer stmt.Close()

	rows, err := stmt.Query(groupName)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var archiveRecords ArchiveRecords
	for rows.Next() {
		err = rows.Scan(&archiveRecords.ExecuteTime)
		if err != nil {
			return "", err
		}
	}

	// 解析起始时间，例如 09:06:24.391
	start, err := time.Parse("2006-01-02 15:04:05", fileDate+" "+startTime)
	if err != nil {
		return "", err
	}

	// 将执行时长字符串 "HH:MM:SS" 转为 Duration
	duration, err := time.ParseDuration(archiveRecords.ExecuteTime)
	if err != nil {
		return "", err
	}

	// 计算预计结束时间
	expectedEnd := start.Add(duration)

	// 返回格式化为 "15:04:05.000"
	startTime = start.Format("15:04:05")
	return "【开始时间：" + fileDate + " " + startTime + "】\n" +
		"【上次时长：" + archiveRecords.ExecuteTime + "】\n" +
		"【预计结束时间：" + fileDate + " " + expectedEnd.Format("15:04:05") + "】", nil
}

// ListArchive 归档查询
func ListArchive() []ArchiveRecords {
	stmt, err := config.DB.Prepare(`SELECT id, title, execute_time, created_at FROM archive_records`)
	if err != nil {
		return []ArchiveRecords{}
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		return []ArchiveRecords{}
	}
	defer rows.Close()

	//var archiveRecords []ArchiveRecords
	archiveRecords := make([]ArchiveRecords, 0)
	for rows.Next() {
		var record ArchiveRecords
		err = rows.Scan(&record.Id, &record.Title, &record.ExecuteTime, &record.CreatedAt)
		if err != nil {
			continue // 或者记录日志
		}
		archiveRecords = append(archiveRecords, record)
	}

	return archiveRecords
}

// JsVersion 读取脚本的版本号
func JsVersion(jsName, nowVersion string) string {

	repoDir := config.Cfg.BetterGIAddress + "/Repos/bettergi-scripts-list-git/repo/js"

	filePath := filepath.Join(repoDir, jsName, "manifest.json")
	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		autoLog.Sugar.Errorf("读取文件失败: %v", err)
	}
	// 解析 JSON
	var data map[string]interface{}
	err = json.Unmarshal(content, &data)
	if err != nil {
		autoLog.Sugar.Errorf("JsVersion 解析 JSON 失败: %v", err)
	}
	// 提取版本号
	version, ok := data["version"].(string)
	if !ok {
		autoLog.Sugar.Errorf("JsVersion 版本号格式错误")
		return "未知"
	}

	if nowVersion == version {
		return "最新"
	}
	return "有更新[" + version + "]"

}

var aa string
var i int

func ReadLog() {
	filePath := filepath.Clean(fmt.Sprintf("%s\\log", config.Cfg.BetterGIAddress))
	files, err := FindLogFiles(filePath)
	if err != nil || len(files) == 0 {
		fmt.Println("找不到日志文件")
		return
	}
	fileLog := files[0]
	file, err := os.Open(filepath.Join(filePath, fileLog))
	if err != nil {
		fmt.Println("无法打开日志文件:", err)
		return
	}
	defer file.Close()

	// 定位到文件末尾
	file.Seek(0, io.SeekEnd)

	reader := bufio.NewReader(file)
	for {
		line, _ := reader.ReadString('\n')

		if aa == line {
			if i < 30 {
				i++
				aa = line
				time.Sleep(1000 * time.Millisecond)
				continue
			} else if i == 30 {
				autoLog.Sugar.Info("bgi" + strconv.Itoa(i) + "秒没有动静")
				Notice.SentText("bgi30秒没有动静")
				i++
			}
		} else {
			aa = line
			i = 0
		}

	}
}

var errorKeywords = []string{
	"未识别到突发任务",
	"OCR 识别失败",
	"此路线出现3次卡死，重试一次路线或放弃此路线！",
	"检测到复苏界面，存在角色被击败",
	"执行路径时出错",
	"传送点未激活或不存在",
	"疑似卡死，尝试脱离...",
	"此追踪脚本未正常走完！",
}

func isErrorLine(line string) (matched string, ok bool) {
	for _, keyword := range errorKeywords {
		if strings.Contains(line, keyword) {
			return keyword, true
		}
	}
	return "", false
}

// 🔹执行分段
type ExecutionSegment struct {
	StartTime string
	EndTime   string
	Consuming string
}

type LogAnalysis2Struct struct {
	GroupName        string
	Segments         []ExecutionSegment // 每段执行的开始/结束/耗时
	LogAnalysis2Json []LogAnalysis2Json
	ErrorSummary     map[string]int // 所有分段合并的错误统计
	SumIncome        map[string]int // 所有分段合并的收入统计
}

type LogAnalysis2Json struct {
	JsonName   string
	StartTime  string
	EndTime    string
	Income     map[string]int // 收入项及其数量
	Errors     map[string]int // 错误项及其数量
	ErrorsMark map[string]int
	Consuming  string
}

// 日志分析
func LogAnalysis2(fileName string) []LogAnalysis2Struct {
	filePath := filepath.Join(config.Cfg.BetterGIAddress, "log")
	fullPath := filepath.Join(filePath, fileName)
	date := GetFileNameDate(fileName)

	file, err := os.Open(fullPath)
	if err != nil {
		fmt.Println("无法打开日志文件:", err)
		return nil
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	// 🔹正则提前编译（避免循环里重复编译）
	startRegexp := regexp.MustCompile(`配置组 "(.*?)" 加载完成`)
	endRegexp := regexp.MustCompile(`配置组 "(.*?)" 执行结束`)
	pickupRegexp := regexp.MustCompile(`交互或拾取："(.*?)"`)

	var logAnalysis2Structs []LogAnalysis2Struct
	var currentStruct *LogAnalysis2Struct
	var lastLine string
	var xy string

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				autoLog.Sugar.Infof("分析完毕")
				// 🔹文件结束时检查是否有未结束的配置组
				if currentStruct != nil {
					logAnalysis2Structs = append(logAnalysis2Structs, *currentStruct)
				}
				break
			}
			autoLog.Sugar.Errorf("配置组分析文件出错: %v", err)
			break
		}

		// 默认时间戳
		timestampLine := lastLine
		if tools.HasTimestamp(line) {
			timestampLine = line
		}
		if timestampLine == "" {
			// 防止开头无时间戳导致空值
			timestampLine = date + " 00:00:00"
		}

		// 配置组开始
		if matches := startRegexp.FindStringSubmatch(line); len(matches) > 1 {
			currentStruct = &LogAnalysis2Struct{
				GroupName:    matches[1],
				ErrorSummary: make(map[string]int),
				SumIncome:    make(map[string]int),
			}
			if t, err := tools.ExtractLogTime2(date, timestampLine); err == nil {
				seg := ExecutionSegment{StartTime: t}
				currentStruct.Segments = append(currentStruct.Segments, seg)
			}
		}

		// 配置组结束
		if currentStruct != nil {
			if matches := endRegexp.FindStringSubmatch(line); len(matches) > 1 && matches[1] == currentStruct.GroupName {
				if t, err := tools.ExtractLogTime2(date, timestampLine); err == nil {
					segIdx := len(currentStruct.Segments) - 1
					currentStruct.Segments[segIdx].EndTime = t
					currentStruct.Segments[segIdx].Consuming =
						tools.CalculateDuration(currentStruct.Segments[segIdx].StartTime, t)
				}

				// 🔹统计错误（按子任务聚合）
				for _, subTask := range currentStruct.LogAnalysis2Json {
					for errStr, count := range subTask.Errors {
						currentStruct.ErrorSummary[errStr] += count
					}
				}

				logAnalysis2Structs = append(logAnalysis2Structs, *currentStruct)
				currentStruct = nil
			}
		}

		// 地图追踪任务开始
		if currentStruct != nil && strings.HasPrefix(line, "→ 开始执行地图追踪任务") {
			subTask := LogAnalysis2Json{JsonName: line}
			if t, err := tools.ExtractLogTime2(date, timestampLine); err == nil {
				subTask.StartTime = t
			}
			currentStruct.LogAnalysis2Json = append(currentStruct.LogAnalysis2Json, subTask)
		}

		// 脚本执行结束（地图追踪/JS 脚本共用）
		if currentStruct != nil && strings.HasPrefix(line, "→ 脚本执行结束") {
			n := len(currentStruct.LogAnalysis2Json)
			if n > 0 {
				current := &currentStruct.LogAnalysis2Json[n-1]
				if t, err := tools.ExtractLogTime2(date, timestampLine); err == nil {
					current.EndTime = t
					current.Consuming = tools.CalculateDuration(current.StartTime, current.EndTime)
				}
			}
		}

		// JS脚本开始
		if currentStruct != nil && strings.HasPrefix(line, "→ 开始执行JS脚本") {
			subTask := LogAnalysis2Json{JsonName: line}
			if t, err := tools.ExtractLogTime2(date, timestampLine); err == nil {
				subTask.StartTime = t
			}
			currentStruct.LogAnalysis2Json = append(currentStruct.LogAnalysis2Json, subTask)
		}

		// 收入情况
		if currentStruct != nil && pickupRegexp.MatchString(line) {
			matches := pickupRegexp.FindStringSubmatch(line)
			if len(matches) > 1 {
				item := matches[1]
				n := len(currentStruct.LogAnalysis2Json)
				if n > 0 {
					current := &currentStruct.LogAnalysis2Json[n-1]
					if current.Income == nil {
						current.Income = make(map[string]int)
					}
					current.Income[item]++
					currentStruct.SumIncome[item]++
				}
			}
		}

		// 错误记录
		if currentStruct != nil {
			if matched, ok := isErrorLine(line); ok {
				n := len(currentStruct.LogAnalysis2Json)
				if n > 0 {
					current := &currentStruct.LogAnalysis2Json[n-1]
					if current.Errors == nil {
						current.Errors = make(map[string]int)
					}
					if current.ErrorsMark == nil {
						current.ErrorsMark = make(map[string]int)
					}
					current.Errors[matched]++
					if xy != "" {
						current.ErrorsMark[xy]++
					}
				}
			}
		}

		// 坐标记录
		if strings.Contains(line, "粗略接近途经点，位置") {
			xy = line
		}

		lastLine = line
	}

	// 🔹合并同名配置组（分段执行支持）
	merged := make(map[string]*LogAnalysis2Struct)
	for _, s := range logAnalysis2Structs {
		m, ok := merged[s.GroupName]
		if !ok {
			copyS := s
			if copyS.ErrorSummary == nil {
				copyS.ErrorSummary = make(map[string]int)
			}
			if copyS.SumIncome == nil {
				copyS.SumIncome = make(map[string]int)
			}
			merged[s.GroupName] = &copyS
			continue
		}

		// 🔹追加分段
		m.Segments = append(m.Segments, s.Segments...)

		// 🔹合并子任务
		m.LogAnalysis2Json = append(m.LogAnalysis2Json, s.LogAnalysis2Json...)

		// 🔹合并错误统计
		for k, v := range s.ErrorSummary {
			m.ErrorSummary[k] += v
		}

		// 🔹合并收入统计
		for k, v := range s.SumIncome {
			m.SumIncome[k] += v
		}
	}

	result := make([]LogAnalysis2Struct, 0, len(merged))
	for _, v := range merged {
		result = append(result, *v)
	}

	var sumLog LogAnalysis2Struct
	sumLog.GroupName = "合计"
	//总耗时
	sumSeg, _ := time.ParseDuration("0s")
	var seg ExecutionSegment
	// 🔹累加总耗时
	for _, s := range result {
		for _, seg := range s.Segments {
			dur, _ := time.ParseDuration(seg.Consuming)
			sumSeg += dur
		}
	}
	seg.Consuming = sumSeg.String()
	sumLog.Segments = append(sumLog.Segments, seg)
	result = append(result, sumLog)

	// 🔹按配置组最早开始时间排序
	sort.Slice(result, func(i, j int) bool {
		if result[i].GroupName == "合计" {
			return false
		}
		if result[j].GroupName == "合计" {
			return true
		}
		if len(result[i].Segments) == 0 || len(result[j].Segments) == 0 {
			return result[i].GroupName < result[j].GroupName
		}
		return result[i].Segments[0].StartTime < result[j].Segments[0].StartTime
	})

	return result
}

type JsNamesInfoStruct struct {
	Name        string
	ChineseName string
	NowVersion  string
	NewVersion  string
	Mark        string
}

func JsNamesInfo() []JsNamesInfoStruct {

	GitPull()
	time.Sleep(1)

	// 获取本地所有订阅脚本目录
	scriptDir := filepath.Join(config.Cfg.BetterGIAddress, "User", "JsScript")
	subDirs, err := tools.ListSubDirsOnly(scriptDir)
	if err != nil {
		autoLog.Sugar.Errorf("获取本地脚本失败: %v", err)
		return nil
	}

	jsNamesInfoStructs := make([]JsNamesInfoStruct, 0, len(subDirs))

	for _, name := range subDirs {
		nowVersion := getJsNowVersion(scriptDir, name)
		newVersion, chineseName, err := GetJsNewVersion(name)
		if err != nil {
			continue
		}

		mark := "无更新"
		//版本比较
		if tools.CompareVersion(nowVersion, newVersion) == -1 {
			mark = "有更新"
		}

		//if nowVersion != newVersion {
		//	mark = "有更新"
		//}

		jsNamesInfoStructs = append(jsNamesInfoStructs, JsNamesInfoStruct{
			Name:        name,
			NowVersion:  nowVersion,
			NewVersion:  newVersion,
			ChineseName: chineseName,
			Mark:        mark,
		})
	}

	return jsNamesInfoStructs
}

func getJsNowVersion(basePath, jsName string) string {
	return readVersion(filepath.Join(basePath, jsName, "manifest.json"))
}

//func GetMysSignLog() string {
//
//	url := config.Cfg.MySign.Url
//	readLogURL := url + "/read-log"
//	resp, err := http.Get(readLogURL)
//	if err != nil {
//		return ""
//	}
//	defer resp.Body.Close()
//	body, err := io.ReadAll(resp.Body)
//	if err != nil {
//		return ""
//	}
//	return string(body)
//}

func readVersion(manifestPath string) string {
	file, err := os.Open(manifestPath)
	if err != nil {
		autoLog.Sugar.Warnf("打开文件失败: %v", err)
		return "未知版本"
	}
	defer file.Close()

	var data map[string]interface{}
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		autoLog.Sugar.Warnf("解析JSON失败: %d%v", manifestPath, err)
		return "未知版本"
	}

	if version, ok := data["version"].(string); ok {
		return version
	}
	return "未知版本"
}

var monitor *LogMonitor
var currentLogFile string

// 监控日志（支持每天变化的日志文件）
func LogM() {
	logDir := filepath.Clean(fmt.Sprintf("%s\\log", config.Cfg.BetterGIAddress))

	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for {
		files, err := FindLogFiles(logDir)
		if err != nil || len(files) == 0 {
			fmt.Println("找不到日志文件")
			<-ticker.C
			continue
		}

		newLogFile := filepath.Join(logDir, files[0])

		if newLogFile != currentLogFile {
			fmt.Printf("检测到新日志文件: %s\n", newLogFile)
			currentLogFile = newLogFile

			go func() {
				config.Cfg.BgiLog = newLogFile
				InitBgiLogStatus()
			}()

			if monitor != nil {
				monitor.Stop()
			}

			monitor = NewLogMonitor(newLogFile, config.Cfg.LogKeywords, 5)
			go monitor.Monitor()
		}

		<-ticker.C
	}
}

func Log1Remote() {

	var Log1RemoteCurrentLogFile string
	var Log1RemoteMonitor *LogMonitor

	// 关键字
	keywords := []string{
		"OnRdpClientDisconnected",
	}

	// 每 30 分钟检查一次最新日志文件
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for {
		files, err := FindLogFiles1Remote(config.Cfg.OneRemote.LogFilePath)
		if err != nil || len(files) == 0 {
			fmt.Println("找不到 1Remote 日志文件")
			<-ticker.C
			continue
		}

		newLogFile := filepath.Join(config.Cfg.OneRemote.LogFilePath, files[0]) // 最新文件

		if newLogFile != Log1RemoteCurrentLogFile {
			fmt.Printf("检测到新的 1Remote 日志文件: %s\n", newLogFile)
			Log1RemoteCurrentLogFile = newLogFile

			if Log1RemoteMonitor != nil {
				Log1RemoteMonitor.Stop()
			}

			Log1RemoteMonitor = NewLogMonitor(newLogFile, keywords, 5)
			go Log1RemoteMonitor.Monitor()
		}

		<-ticker.C
	}
}

// 将今日所有配置组归档
func ArchiveConfig() []LogAnalysis2Struct {
	// 生成日志文件名
	date := time.Now().Format("20060102")
	filename := fmt.Sprintf("better-genshin-impact%s.log", date)
	//获取今日所有配置组
	groupTime := GroupTime(filename)
	for _, groupMap := range groupTime {

		Archive(groupMap)

		autoLog.Sugar.Infof("归档配置组 %s", groupMap.GroupName)
	}
	return groupTime

}
