package bgiStatus

import (
	"auto-bgi/abgiConstant"
	"auto-bgi/autoLog"
	"bufio"
	"github.com/gin-gonic/gin"
	"os"
	"strconv"
	"strings"
)

type EatStatistics struct {
	Time  string
	Name  string
	Count int
}

func EatStatisticsList(context *gin.Context) {

	filePath := abgiConstant.JsPath("营养袋吃药统计", "assets", "默认账户.txt")
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		autoLog.Sugar.Errorf("营养袋吃药统计err: %v", err)
	}
	defer file.Close()
	// 读取文件内容
	scanner := bufio.NewScanner(file)
	mapData := make(map[string][]EatStatistics)
	for scanner.Scan() {
		var eatStatistics EatStatistics
		line := scanner.Text()
		split := strings.Split(line, "-")
		eatStatistics.Time = split[0]
		eatStatistics.Name = split[1]
		Count, err := strconv.Atoi(split[2])
		if err != nil {
			autoLog.Sugar.Errorf("营养袋吃药统计err: %v", err)
		}
		eatStatistics.Count = Count

		s := strings.Split(split[0], " ")
		Date := strings.ReplaceAll(s[0], "时间:", "")

		if _, ok := mapData[Date]; !ok {
			mapData[Date] = []EatStatistics{}
		}
		mapData[Date] = append(mapData[Date], eatStatistics)

	}
	context.JSON(200, mapData)
}
