package ArtifactsBulkSupply

import (
	"auto-bgi/Notice"
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"auto-bgi/tools"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// DogFood 狗粮批发js相关代码
type DogFood struct {
}

// 查询今天狗粮批发是什么线路
func (d *DogFood) DogFoodIsAOrB() string {
	autoLog.Sugar.Infof("狗粮批发线路查询")
	fileName := "默认账户"
	if config.Cfg.Account.IsMultiUser {
		fileName = config.Cfg.Account.Uid
		autoLog.Sugar.Infof("批发是多用户,查询文件：%s.txt", fileName)
	} else {
		autoLog.Sugar.Infof("不是多用户，查询默认账户")
	}

	filePath := filepath.Clean(fmt.Sprintf("%s\\User\\JsScript\\AAA-Artifacts-Bulk-Supply\\records\\%s.txt", config.Cfg.BetterGIAddress, fileName))
	file, err := os.Open(filePath)
	if err != nil {
		autoLog.Sugar.Errorf("狗粮批发txt打开失败，err:%v", err)
		autoLog.Sugar.Infof("获取当前目录下所有 .txt 文件")
		files, err := filepath.Glob(fmt.Sprintf("%s\\User\\JsScript\\AAA-Artifacts-Bulk-Supply\\records\\*.txt", config.Cfg.BetterGIAddress))
		if err != nil {
			autoLog.Sugar.Errorf("狗粮批发没有找到任何文件")
			return ""
		}
		if len(files) == 0 {
			autoLog.Sugar.Errorf("未找到任何txt文件")
			return ""
		}
		autoLog.Sugar.Infof("找到txt文件：%s", filepath.Base(files[0]))
		filePath = filepath.Clean(fmt.Sprintf("%s\\User\\JsScript\\AAA-Artifacts-Bulk-Supply\\records\\%s", config.Cfg.BetterGIAddress, filepath.Base(files[0])))
		file, err = os.Open(filePath)
		if err != nil {
			autoLog.Sugar.Errorf("狗粮批发txt打开失败，err:%v", err)
			return ""
		}
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)

	//格式化今天的日期
	today := time.Now().Format("2006/01/02")

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "日期:"+today) {
			autoLog.Sugar.Infof("今天的狗粮批发线路是:%s", line)
			//Notice.SentText(fmt.Sprintf("今天的狗粮批发线路是:%s", line))
			if strings.Contains(line, "A") {
				autoLog.Sugar.Infof("今天的狗粮批发线路是A")
				return "A"
			} else if strings.Contains(line, "B") {
				autoLog.Sugar.Infof("今天的狗粮批发线路是B")
				return "B"
			} else {
				autoLog.Sugar.Errorf("今天的狗粮批发线路查询失败，未找到A或B")
				//Notice.SentText("今天的狗粮批发线路查询失败，未找到A或B")
				return ""
			}
		}

	}
	if err := scanner.Err(); err != nil {
		autoLog.Sugar.Errorf("狗粮批发txt读取失败，err:%v", err)
		return ""
	}
	return ""

}

// 将识别的数量写入到狗粮批发
func (d *DogFood) WriteDogFoodNum(num string) string {

	autoLog.Sugar.Infof("批发写入数量：%s", num)
	fileName := "默认账户"
	if config.Cfg.Account.IsMultiUser {
		fileName = config.Cfg.Account.Uid
		autoLog.Sugar.Infof("批发是多用户,查询文件：%s.txt", fileName)
	} else {
		autoLog.Sugar.Infof("不是多用户，查询默认账户")
	}

	filePath := filepath.Clean(fmt.Sprintf("%s\\User\\JsScript\\AAA-Artifacts-Bulk-Supply\\records\\%s.txt", config.Cfg.BetterGIAddress, fileName))
	file, err := os.Open(filePath)
	if err != nil {
		autoLog.Sugar.Errorf("狗粮批发txt打开失败，err:%v", err)
		autoLog.Sugar.Infof("获取当前目录下所有 .txt 文件")
		files, err := filepath.Glob(fmt.Sprintf("%s\\User\\JsScript\\AAA-Artifacts-Bulk-Supply\\records\\*.txt", config.Cfg.BetterGIAddress))
		if err != nil {
			autoLog.Sugar.Errorf("狗粮批发没有找到任何文件")
			return ""
		}
		if len(files) == 0 {
			autoLog.Sugar.Errorf("未找到任何txt文件")
			return ""
		}
		autoLog.Sugar.Infof("找到txt文件：%s", filepath.Base(files[0]))
		filePath = filepath.Clean(fmt.Sprintf("%s\\User\\JsScript\\AAA-Artifacts-Bulk-Supply\\records\\%s", config.Cfg.BetterGIAddress, filepath.Base(files[0])))
		file, err = os.Open(filePath)
		if err != nil {
			autoLog.Sugar.Errorf("狗粮批发txt打开失败，err:%v", err)
			return ""
		}
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)

	//格式化今天的日期
	today := time.Now().Format("2006/01/02")
	var buf bytes.Buffer
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "日期:"+today) {
			autoLog.Sugar.Infof("批发:%s", line)
			re := regexp.MustCompile(`狗粮经验(\d+)`)
			//提取出数字
			match := re.FindStringSubmatch(line)
			if len(match) > 1 {
				autoLog.Sugar.Infof("找到数字：%s", match[1])
				//相加
				a, _ := strconv.Atoi(match[1])
				b, _ := strconv.Atoi(num)
				sum := strconv.Itoa(a + b)

				newText := re.ReplaceAllString(line, "狗粮经验"+sum)
				buf.WriteString(newText + "\n")
				//批发和联机狗粮相加
				autoLog.Sugar.Infof("批发:[%d]和联机狗粮:[%d]相加等于：%s", a, b, sum)
				Notice.SentText(fmt.Sprintf("批发:[%d]和联机狗粮:[%d]相加等于：%s", a, b, sum))

				go func() {
					if a+b >= 150000 && a+b <= 190000 {
						UpdateRevenue(sum)
					} else {
						autoLog.Sugar.Infof("a+b = %d，不在目标区间，未执行 UpdateRevenue", a+b)
					}
				}()

				continue
			} else {
				re2 := regexp.MustCompile(`狗粮经验NaN`)
				newText := re2.ReplaceAllString(line, "狗粮经验"+num)
				buf.WriteString(newText + "\n")
				//批发和联机狗粮相加
				autoLog.Sugar.Infof("批发:[%s]和联机狗粮:[%s]相加等于：%s", "识别错误", num, num)
				Notice.SentText(fmt.Sprintf("批发:[%s]和联机狗粮:[%s]相加等于：%s", "识别错误", num, num))

				go func() {
					UpdateRevenue(num)
				}()
				continue
			}
		}
		buf.WriteString(line + "\n")

	}
	if err := scanner.Err(); err != nil {
		autoLog.Sugar.Errorf("狗粮批发txt读取失败，err:%v", err)
		return ""
	}
	err = os.WriteFile(filePath, buf.Bytes(), 0644)
	if err != nil {
		autoLog.Sugar.Errorf("狗粮批发txt写入失败，err:%v", err)
	}
	return ""
}

func UpdateRevenue(num string) {
	decrypt, err2 := tools.Decrypt(config.Cfg.Account.SecretKey)
	if err2 != nil {
		return
	}

	resp, err := http.Post(fmt.Sprintf("http://%s/api/updateRevenue/"+config.Cfg.Account.Uid+"/"+num, decrypt), "application/json", nil)
	if err != nil {
		autoLog.Sugar.Error("更新收益失败:", err)
		return
	}
	//读取返回
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		autoLog.Sugar.Error("更新收益失败:", err)
	}
	autoLog.Sugar.Infof("更新收益:%s", string(body))

	defer resp.Body.Close()

}
