package JsAPI

import (
	"auto-bgi/Notice"
	"auto-bgi/bgiStatus"
	"auto-bgi/config"
	"fmt"
	"github.com/gin-gonic/gin"
	"path/filepath"
	"sort"
)

func SendLogAnalysis(context *gin.Context) {

	logAnalysis := bgiStatus.LogAnalysis(filepath.Base(config.Cfg.BgiLog))

	// 创建结构体切片来存储键值对
	type kv struct {
		Key   string
		Value int
	}
	var sortedData []kv
	for k, v := range logAnalysis {
		sortedData = append(sortedData, kv{k, v})
	}
	// 按照值进行排序
	sort.Slice(sortedData, func(i, j int) bool {
		return sortedData[i].Value > sortedData[j].Value
	})

	data := "今日收获:\n"
	for _, i := range sortedData {
		data += fmt.Sprintf("   %s: %d\n", i.Key, i.Value)
	}
	Notice.SentText(data)
	context.JSON(200, data)

}
