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
	filePath := filepath.Clean(fmt.Sprintf("%s\\User\\JsScript\\AAA-Artifacts-Bulk-Supply\\records\\默认账户.txt", config.Cfg.BetterGIAddress))
	file, err := os.Open(filePath)
	if err != nil {
		autoLog.Sugar.Errorf("狗粮批发默认账户txt打开失败，err:%v", err)
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
		autoLog.Sugar.Errorf("狗粮批发默认账户txt读取失败，err:%v", err)
		return ""
	}
	return ""

}
