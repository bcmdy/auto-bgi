package bgiStatus

import (
	"archive/zip"
	"auto-bgi/BackpackStatistics"
	"auto-bgi/Notice"
	"auto-bgi/ScriptRepo"
	"auto-bgi/abgiConstant"
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"auto-bgi/control"
	"auto-bgi/models"
	"auto-bgi/tools"
	"bufio"
	"encoding/json"
	"fmt"
	"io"
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

	"github.com/agnivade/levenshtein"
	"github.com/fsnotify/fsnotify"
	"github.com/otiai10/copy"
	"github.com/robfig/cron/v3"
	"github.com/tidwall/gjson"
)

// 检查 程序 是否在运行
func IsWechatRunning(name string) bool {
	cmd := exec.Command("tasklist", "/FI", "IMAGENAME eq "+name)
	//autoLog.Sugar.Infof("执行命令：%s", cmd.String())
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := cmd.Output()
	if err != nil {

		autoLog.Sugar.Error("BetterGI.exe 是否在运行:", err)
		return false
	}

	return strings.Contains(string(output), "BetterGI.exe")
}

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
				autoLog.Sugar.Infof("BetterGI 已关闭: %s", time.Now().Format("2006-01-02 15:04:05"))
				// 检查配置文件中是否设置了需要关闭原神
				if config.Cfg.Control.IsCloseYuanShen {
					autoLog.Sugar.Infof("需要关闭原神")
					control.CloseYuanShen()

				}
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

func TodayHarvest(fileName string) (map[string]int, error) {

	autoLog.Sugar.Infof("今日收获统计")
	re := regexp.MustCompile(`^交互或拾取："([^"]*)"`)

	filename := filepath.Clean(fmt.Sprintf("%s\\log\\%s", config.Cfg.BetterGIAddress, fileName))

	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("TodayHarvest打开文件失败: %v", err)
	}
	defer file.Close()

	// 初始化map用于存储物品和出现次数
	harvestStats := make(map[string]int)

	scanner := bufio.NewScanner(file)

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

	//statistics := config.Cfg.BagStatistics
	//
	//split := strings.Split(statistics, ",")
	statistics := BackpackStatistics.List()

	var bags []Material
	var bag Material

	bagMap := make(map[string]Material)

	layout := "2006/1/2 15:04:05"

	for scanner.Scan() {
		for _, statistic := range statistics {
			// 创建一个正则表达式来匹配 "晶蝶：数字" 模式
			sprintf := fmt.Sprintf(`(?:^|[,\s])%s: (\d+)`, statistic.Material)

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

					if y1 == y2 && m1 == m2 && d1 == d2 {
						continue
					} else {
						bagMap[bag.Cl] = bag
						//判断数据是否一致
						if isNil.Num != bag.Num {
							bags = append(bags, bag)
						}

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

	if len(bags) == 0 {
		bag.Data = "没有数据"
		bag.Cl = "没有数据"
		bag.Num = "没有数据"
		bags = append(bags, bag)
	}

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
	var bags []Material
	filename := filepath.Clean(fmt.Sprintf("%s\\User\\JsScript\\OCR读取当前抽卡资源并发送通知\\Resources_log.txt", config.Cfg.BetterGIAddress))
	file, err := os.Open(filename)
	if err != nil {
		autoLog.Sugar.Errorf("没有相关JS:OCR读取当前抽卡资源并发送通知")
		return bags, err
	}
	defer file.Close()

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
	var bags []Material
	autoLog.Sugar.Infof("摩拉统计")
	filename := filepath.Clean(fmt.Sprintf("%s\\User\\JsScript\\OCR读取当前摩拉记录并发送通知\\mora_log.txt", config.Cfg.BetterGIAddress))
	// 打开文件
	file, err := os.Open(filename)
	if err != nil {
		autoLog.Sugar.Infof("没有相关JS")
		return bags, err
	}
	defer file.Close()

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
	FileName     string
	Mark         string
	ActivateTime string
	Detail       []string
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
		Asum := 0

		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "上次激活收尾路线时间: ") {
				timeStr := strings.ReplaceAll(line, "上次激活收尾路线时间: ", "")
				parsedTime, err := time.Parse(time.RFC3339, timeStr)
				if err != nil {
					fmt.Println("Error parsing time:", err)
				}
				//加8小时
				parsedTime = parsedTime.Add(8 * time.Hour)
				dogFood.ActivateTime = "上次激活收尾路线时间: " + parsedTime.Format("2006-01-02 15:04:05")
			}

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

		for _, s := range dogFood.Detail {
			if strings.Contains(s, "额外A") {
				Asum++
			} else {
				break
			}
		}
		dogFood.Mark = dogFood.Mark + fmt.Sprintf("(连续%d个A)", Asum)

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

// LogAnalysis TodayHarvest 获取今日收获
// LogAnalysis 函数用于分析日志文件并统计材料数量
// 参数 fileName: 日志文件名
// 返回值: map[string]int - 材料名称及其数量的映射，只返回数量超过10的材料
func LogAnalysis(fileName string) []map[string]interface{} {
	// 记录日志分析开始
	autoLog.Sugar.Infof("日志分析")
	// 获取今日收获的数据，忽略错误
	res, _ := TodayHarvest(fileName)

	var syw = 0 // 圣遗物计数器
	var xie = 0 // 螃蟹计数器

	// 创建一个map用于存储材料统计结果
	data := make(map[string]int)

	// 遍历今日收获的每个项目及其数量
	for item, count := range res {

		// 清理项目名称，移除特殊字符
		item = strings.ReplaceAll(item, "·", "")
		item = strings.ReplaceAll(item, "。", "")

		// 判断项目是否属于圣遗物类别
		if IsStringInDictionaryCategory(item, Relics) {
			syw += count // 增加圣遗物计数
		} else if strings.Contains(item, "蟹") {
			xie += count // 增加螃蟹计数
		} else {
			// 处理普通材料
			material := abgiConstant.Material
			var name = item           // 最匹配的材料名称
			var bestSim float64 = 0.0 // 最高相似度

			// 遍历所有材料进行匹配
			for _, m := range material {
				// 特殊处理：月萤虫视为晶蝶
				if item == "月萤虫" {
					name = "晶蝶"
					bestSim = 1.0
					break
				}

				// 计算当前材料与日志项目的相似度
				_, f := ComputeDistance(item, m)
				if f > bestSim { // 找到更相似的匹配
					bestSim = f
					name = m
				}
			}

			// 如果相似度达不到阈值，则视为未知
			if bestSim < 0.1 {
				autoLog.Sugar.Infof("未判断出的材料:%s", item)
			}
			data[name] += count
		}

	}

	data["圣遗物"] = syw
	data["螃蟹"] = xie

	statistics := BackpackStatistics.List()
	dd := make(map[string]uint)
	for _, s := range statistics {
		dd[s.Material] = s.ID
	}

	// 取出数量超过10的材料
	var mapDatas []map[string]interface{}
	for s, i2 := range data {
		mapData := make(map[string]interface{})
		if i2 >= 10 {
			mapData["name"] = s
			mapData["count"] = i2
			//判断是否已经关注
			if _, ok := dd[s]; ok {
				mapData["isFocus"] = true
			} else {
				mapData["isFocus"] = false
			}
			mapDatas = append(mapDatas, mapData)
		}
	}

	return mapDatas

}

// 相似的计算
func ComputeDistance(a, b string) (int, float64) {
	distance := levenshtein.ComputeDistance(a, b)
	//fmt.Println("编辑距离:", distance)

	// 转化为相似度（0~1之间）
	similarity := 1 - float64(distance)/float64(max(len([]rune(a)), len([]rune(b))))
	//fmt.Printf("相似度: %.2f\n", similarity)
	return distance, similarity
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func findLogFiles(dirPath string) ([]string, error) {
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
func UnzipRepo(zipPath, outputDir, targetPrefix string) error {
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

func Backup() error {

	autoLog.Sugar.Infof("开始备份user文件夹")
	err4 := ZipDir(config.Cfg.BetterGIAddress+"\\User\\", "Users\\User"+time.Now().Format("2006-01-02-15-04-05")+".zip", true)
	if err4 != nil {
		autoLog.Sugar.Errorf("备份失败: %v", err4)
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
// GroupTime 函数用于分析日志文件并按配置组进行分组统计
// 参数:
//
//	fileName - 日志文件名
//
// 返回值:
//
//	[]LogAnalysis2Struct - 分析结果的结构体切片
func GroupTime(fileName string) []LogAnalysis2Struct {
	// 构建完整的日志文件路径
	filePath := filepath.Join(config.Cfg.BetterGIAddress, "log")
	fullPath := filepath.Join(filePath, fileName)
	date := GetFileNameDate(fileName) // 从文件名中提取日期

	// 打开日志文件
	file, err := os.Open(fullPath)
	if err != nil {
		fmt.Println("GroupTime无法打开日志文件:", err)
		return nil
	}
	defer file.Close() // 确保文件在函数结束时关闭

	// 创建文件读取器
	reader := bufio.NewReader(file)

	// 🔹正则提前编译（避免循环里重复编译）
	startRegexp := regexp.MustCompile(`配置组 "(.*?)" 加载完成`)
	endRegexp := regexp.MustCompile(`配置组 "(.*?)" 执行结束`)

	// 初始化变量
	var logAnalysis2Structs []LogAnalysis2Struct
	var currentStruct *LogAnalysis2Struct
	var lastLine string

	// 逐行读取文件内容
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				//autoLog.Sugar.Infof("分析完毕")
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
		// 将合并后的结果转换为切片

	}

	result := make([]LogAnalysis2Struct, 0, len(merged))
	for _, v := range merged {
		result = append(result, *v)
	}

	return result
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

// git拉取代码
func GitPull() {
	// 从配置文件中获取仓库URL
	//_, _, err := ScriptRepo.UpdateCenterRepoByGit("https://github.com/babalae/bettergi-scripts-list.git")
	_, _, err := ScriptRepo.UpdateRepoGit()
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
	ScriptRepo.RepoLock.Lock()
	defer ScriptRepo.RepoLock.Unlock()

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

	//先删除旧备份文件
	err = os.RemoveAll(backupRoot)
	if err != nil {
		autoLog.Sugar.Errorf("旧文件删除失败：%s,错误：%s", jsName, err)
		return "", err
	}
	autoLog.Sugar.Infof("旧文件删除成功")

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

	count := models.DB.Where("title = ?", title).Find(&models.ArchiveRecords{}).RowsAffected

	autoLog.Sugar.Infof("查询数据库是否存在归档记录：%d", count)

	if count > 0 {
		autoLog.Sugar.Infof("存在归档记录，执行删除操作")

		// 删除已存在的归档记录
		err := models.DeleteArchiveRecordByTitle(title)
		if err != nil {
			autoLog.Sugar.Errorf("删除归档记录失败: %v", err)
			return "删除归档记录失败"
		}

		autoLog.Sugar.Infof("删除归档记录成功")
	}

	autoLog.Sugar.Infof("执行新增归档记录")

	err := models.InsertArchiveRecord(title, executeTime.String())
	if err != nil {
		autoLog.Sugar.Errorf("写入数据库失败: %v", err)
		return "写入数据库失败"
	}
	autoLog.Sugar.Infof("成功归档：%s (%s)", title, executeTime)
	return "归档成功"
}

// 时间计算
func CalculateTime(filename, groupName, startTime string) (string, error) {
	// 解析文件名中的日期
	fileDate := GetFileNameDate(filename)

	//根据title查询数据库
	archiveRecords := models.GetArchiveRecordByTitle(groupName)

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

	//// 计算预计结束时间
	//expectedEnd := start.Add(duration)

	//查询已经运行了多少
	groupTime := GroupTime(filepath.Base(filename))
	sumExecuteTime, _ := time.ParseDuration("0s")
	for _, groupMap := range groupTime {
		if groupMap.GroupName != groupName {
			continue
		}
		executeTime, _ := time.ParseDuration("0s")
		for _, segment := range groupMap.Segments {
			if segment.Consuming != "" {
				duration, err := time.ParseDuration(segment.Consuming)
				if err != nil {
					autoLog.Sugar.Errorf("解析时间失败: %v", err)
					continue
				}
				executeTime += duration

			}

		}
		sumExecuteTime += executeTime
	}

	// 返回格式化为 "15:04:05.000"
	startTime = start.Format("15:04:05")
	//字符串转时间
	//减去运行时长
	d := duration - sumExecuteTime
	//expectedEnd := start.Add(d)

	parse, _ := time.Parse("2006-01-02 15:04:05", fileDate+" "+startTime)
	expectedEnd := parse.Add(d)

	return "【时间：" + fileDate + " " + startTime + "】\n" +
		"【时长：" + archiveRecords.ExecuteTime + "】\n" +
		"【已经运行：" + sumExecuteTime.String() + "】\n" +
		"【预计：" + expectedEnd.Format("2006-01-02 15:04:05") + "】", nil
}

// ListArchive 归档查询
func ListArchive() []models.ArchiveRecords {
	//查询所有归档
	archiveRecords := models.ListArchiveRecords()
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
	files, err := findLogFiles(filePath)
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
	"执行超时，放弃此次追踪",
	"路径点执行超时，放弃整条路径",
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
	Mola      int64
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
	ErrorTime  string
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
				//autoLog.Sugar.Infof("分析完毕")
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
					logByTimeSUM := models.MoraLogByTimeSUM(currentStruct.Segments[segIdx].StartTime, t)
					currentStruct.Segments[segIdx].Mola = logByTimeSUM
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
					time2, _ := tools.ExtractLogTime2(date, timestampLine)

					current.ErrorTime = strings.ReplaceAll(time2, date, "")
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
	LastUpdated string // 最后更新时间
}

type RepoStruct struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	LastUpdated string `json:"lastUpdated"`
}

func JsNamesInfo() []JsNamesInfoStruct {

	GitPull()
	time.Sleep(1)

	ScriptRepo.RepoLock.Lock()
	defer ScriptRepo.RepoLock.Unlock()

	// 获取本地所有订阅脚本目录
	scriptDir := filepath.Join(config.Cfg.BetterGIAddress, "User", "JsScript")
	subDirs, err := tools.ListSubDirsOnly(scriptDir)
	if err != nil {
		autoLog.Sugar.Errorf("获取本地脚本失败: %v", err)
		return nil
	}

	jsNamesInfoStructs := make([]JsNamesInfoStruct, 0, len(subDirs))

	for _, name := range subDirs {
		nowVersion, nowChineseName := GetJsNowVersion(scriptDir, name)
		newVersion, chineseName, err := GetJsNewVersion(name)
		if err != nil {
			jsNamesInfoStructs = append(jsNamesInfoStructs, JsNamesInfoStruct{
				Name:        name,
				NowVersion:  nowVersion,
				NewVersion:  newVersion,
				ChineseName: nowChineseName,
				Mark:        "未知",
			})
			continue
		}

		mark := "无更新"
		//版本比较
		if tools.CompareVersion(nowVersion, newVersion) == -1 {
			mark = "有更新"
		}

		jsNamesInfoStructs = append(jsNamesInfoStructs, JsNamesInfoStruct{
			Name:        name,
			NowVersion:  nowVersion,
			NewVersion:  newVersion,
			ChineseName: chineseName,
			Mark:        mark,
		})
	}

	// 读取
	filePath := filepath.Join(config.Cfg.BetterGIAddress, "Repos", "bettergi-scripts-list-git", "repo.json")
	repo, err := os.ReadFile(filePath)
	if err != nil {
		//panic(err)
		return jsNamesInfoStructs
	}
	newData := repo
	data := gjson.Get(string(newData), "indexes.1.children")
	dataMap := make(map[string]RepoStruct)
	for _, repo := range data.Array() {
		dataMap[repo.Get("name").String()] = RepoStruct{
			Name:        repo.Get("name").String(),
			Version:     repo.Get("version").String(),
			LastUpdated: repo.Get("lastUpdated").String(),
		}

	}
	for i, jsNamesInfoStruct := range jsNamesInfoStructs {
		if v, ok := dataMap[jsNamesInfoStruct.Name]; ok {
			jsNamesInfoStructs[i].LastUpdated = v.LastUpdated
		}
	}

	return jsNamesInfoStructs
}

func GetJsNowVersion(basePath, jsName string) (string, string) {
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

func readVersion(manifestPath string) (string, string) {
	//捕获异常
	defer func() {
		if r := recover(); r != nil {
			autoLog.Sugar.Warnf("捕获异常: %v", r)
			return
		}
	}()

	file, err := os.Open(manifestPath)
	if err != nil {
		autoLog.Sugar.Warnf("readVersion打开文件失败: %v", err)
		return "未知版本", "未知"
	}
	defer file.Close()

	var data map[string]interface{}
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		autoLog.Sugar.Warnf("解析JSON失败: %s %v", manifestPath, err)
		return "未知版本", data["name"].(string)
	}

	if version, ok := data["version"].(string); ok {

		return version, data["name"].(string)
	}
	return "未知版本", data["name"].(string)
}

var monitor *LogMonitor
var currentLogFile string

func switchLogFile(newLogFile string) {
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

func LogM() {
	logDir := filepath.Clean(fmt.Sprintf("%s\\log", config.Cfg.BetterGIAddress))

	files, err := findLogFiles(logDir)
	if err != nil || len(files) == 0 {
		fmt.Println("找不到日志文件")
	} else {
		newLogFile := filepath.Join(logDir, files[0])
		if newLogFile != currentLogFile {
			switchLogFile(newLogFile)
		}
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("创建日志目录监控失败:", err)
		return
	}
	defer watcher.Close()

	if err := watcher.Add(logDir); err != nil {
		fmt.Println("监听日志目录失败:", err)
		return
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			if event.Name == "" {
				continue
			}

			ext := strings.ToLower(filepath.Ext(event.Name))
			if ext != ".log" {
				continue
			}

			if event.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Rename|fsnotify.Remove) == 0 {
				continue
			}

			files, err := findLogFiles(logDir)
			if err != nil || len(files) == 0 {
				fmt.Println("日志事件触发，但未找到日志文件")
				continue
			}

			newLogFile := filepath.Join(logDir, files[0])
			if newLogFile != currentLogFile {
				switchLogFile(newLogFile)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			fmt.Println("日志目录监控错误:", err)
		}
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

// 查询今日执行配置组
func TodayGroupsInfo() {
	//查询规定记录
	ArchiveRecordMap, _ := ArchiveRecordList()

	date := time.Now().Format("20060102")
	filename := fmt.Sprintf("better-genshin-impact%s.log", date)
	//获取今日所有配置组
	groupTime := GroupTime(filename)
	NoticeData := "今日配置组执行情况\n"
	sumExecuteTime, _ := time.ParseDuration("0s")
	for _, groupMap := range groupTime {

		executeTime, _ := time.ParseDuration("0s")
		for _, segment := range groupMap.Segments {
			if segment.Consuming != "" {
				duration, err := time.ParseDuration(segment.Consuming)
				if err != nil {
					autoLog.Sugar.Errorf("解析时间失败: %v", err)
					continue
				}
				executeTime += duration
			}

		}
		Archive(groupMap)
		//计算差值
		var diff time.Duration
		if archiveRecord, ok := ArchiveRecordMap[groupMap.GroupName]; ok {
			duration, err := time.ParseDuration(archiveRecord)
			if err != nil {
				fmt.Println("解析错误:", err)
				return
			}
			diff = executeTime - duration
		}

		NoticeData += fmt.Sprintf("【%s--%s】(%s)\n", groupMap.GroupName, executeTime, diff)
		sumExecuteTime += executeTime
	}
	NoticeData += fmt.Sprintf("【%s--%s】\n", "合计", sumExecuteTime)
	NoticeData += "\n"

	Notice.SentText(NoticeData)

}
