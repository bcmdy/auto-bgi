package bgiStatus

import (
	"auto-bgi/autoLog"
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

func compareDate(a, b string) int {
	ta, _ := time.Parse("2006/1/2", a)
	tb, _ := time.Parse("2006/1/2", b)
	if ta.Before(tb) {
		return -1
	} else if ta.After(tb) {
		return 1
	}
	return 0
}

func DeleteBag() {
	filePath := filepath.Clean(fmt.Sprintf("%s\\User\\JsScript\\背包材料统计\\latest_record.txt", Config.BetterGIAddress))

	// 打开文件读取内容
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	timeRegex := regexp.MustCompile(`^(\d{4}/\d{1,2}/\d{1,2}) \d{2}:\d{2}:\d{2}$`)

	var blocks [][]string
	var currentBlock []string
	var currentDate string
	var latestDate string

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		// 匹配时间戳行
		if match := timeRegex.FindStringSubmatch(line); match != nil {
			// 保存上一个 block
			if len(currentBlock) > 0 {
				blocks = append(blocks, append([]string{}, currentBlock...))
				currentBlock = nil
			}
			currentDate = match[1]

			if latestDate == "" || compareDate(currentDate, latestDate) > 0 {
				latestDate = currentDate
			}
		}
		currentBlock = append(currentBlock, line)
	}
	// 最后一块
	if len(currentBlock) > 0 {
		blocks = append(blocks, currentBlock)
	}

	// 过滤只保留最新日期的数据
	var result []string
	for _, block := range blocks {
		for _, line := range block {
			if match := timeRegex.FindStringSubmatch(line); match != nil {
				if match[1] == latestDate {
					result = append(result, strings.Join(block, "\n"))
				}
				break
			}
		}
	}

	// 直接覆盖原文件
	err = os.WriteFile(filePath, []byte(strings.Join(result, "\n\n")), 0644)
	if err != nil {
		panic(err)
	}

	fmt.Printf("成功：已将 %s 更新为仅包含 %s 的数据。\n", filePath, latestDate)

}

func DeleteMoLa() {

	filePath := filepath.Clean(fmt.Sprintf("%s\\User\\JsScript\\OCR读取当前摩拉记录并发送通知\\mora_log.txt", Config.BetterGIAddress))

	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// 日期行正则（匹配形如：2025/7/5 10:33:45）
	timeRegex := regexp.MustCompile(`^(\d{4}/\d{1,2}/\d{1,2}) \d{2}:\d{2}:\d{2}`)

	var allLines []string
	var dateMap = make(map[string][]string)
	var latestDate string

	for scanner.Scan() {
		line := scanner.Text()
		allLines = append(allLines, line)

		if match := timeRegex.FindStringSubmatch(line); match != nil {
			date := match[1]
			dateMap[date] = append(dateMap[date], line)

			if latestDate == "" || compareDate(date, latestDate) > 0 {
				latestDate = date
			}
		}
	}

	// 写回文件，仅保留最新日期
	if lines, ok := dateMap[latestDate]; ok {
		output := strings.Join(lines, "\n") + "\n"
		err = os.WriteFile(filePath, []byte(output), 0644)
		if err != nil {
			panic(err)
		}
		fmt.Printf("成功：%s 中只保留了 %s 的记录。\n", filePath, latestDate)
	} else {
		fmt.Println("未找到匹配的记录")
	}

}

func DeleteYuanShi() {
	filePath := filepath.Clean(fmt.Sprintf("%s\\User\\JsScript\\OCR读取当前抽卡资源并发送通知\\Resources_log.txt", Config.BetterGIAddress))
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// 日期行正则（匹配形如：2025/7/5 10:33:45）
	timeRegex := regexp.MustCompile(`^(\d{4}/\d{1,2}/\d{1,2}) \d{2}:\d{2}:\d{2}`)

	var allLines []string
	var dateMap = make(map[string][]string)
	var latestDate string

	for scanner.Scan() {
		line := scanner.Text()
		allLines = append(allLines, line)

		if match := timeRegex.FindStringSubmatch(line); match != nil {
			date := match[1]
			dateMap[date] = append(dateMap[date], line)

			if latestDate == "" || compareDate(date, latestDate) > 0 {
				latestDate = date
			}
		}
	}

	// 写回文件，仅保留最新日期
	if lines, ok := dateMap[latestDate]; ok {
		output := strings.Join(lines, "\n") + "\n"
		err = os.WriteFile(filePath, []byte(output), 0644)
		if err != nil {
			panic(err)
		}
		fmt.Printf("成功：%s 中只保留了 %s 的记录。\n", filePath, latestDate)
	} else {
		fmt.Println("未找到匹配的记录")
	}
}

func GetJsNowVersion(jsName string) string {
	version := ReadVersion(fmt.Sprintf("%s\\User\\JsScript\\%s", Config.BetterGIAddress, jsName))

	return version
}

func GetJsNewVersion(jsName string) (string, string) {
	repoDir := Config.BetterGIAddress + "/Repos/bettergi-scripts-list-git/repo/js"

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
		autoLog.Sugar.Errorf("GetJsNewVersion 解析 JSON 失败: %v", err)
	}
	// 提取版本号
	version, ok := data["version"].(string)
	if !ok {
		autoLog.Sugar.Errorf("GetJsNewVersion 版本号格式错误")
		return "未知", "未知"
	}
	//提取名字
	name, ok := data["name"].(string)
	if !ok {
		autoLog.Sugar.Errorf("GetJsNewVersion 名字格式错误")
		return "未知", "未知"
	}

	return version, name
}
