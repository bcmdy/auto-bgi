package ArtifactsBulkSupply

import (
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"bufio"
	"fmt"
	"os"
	"path/filepath"
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
		autoLog.Sugar.Infof("批发是多用户,查询文件：%s.txt", fileName)
		fileName = config.Cfg.Account.Uid
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
			if strings.Contains(line, "A") {
				autoLog.Sugar.Infof("今天的狗粮批发线路是A")
				return "A"
			} else if strings.Contains(line, "B") {
				autoLog.Sugar.Infof("今天的狗粮批发线路是B")
				return "B"
			} else {
				autoLog.Sugar.Errorf("今天的狗粮批发线路查询失败，未找到A或B")
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
