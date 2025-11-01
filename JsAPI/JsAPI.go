package JsAPI

import (
	"auto-bgi/Notice"
	"auto-bgi/Ocr"
	"auto-bgi/autoLog"
	"auto-bgi/bgiStatus"
	"auto-bgi/config"
	"auto-bgi/control"
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

type ScreenShotImages struct {
	ImagesName string `json:"imagesName"` // 图片名称
	ImagesX    int    `json:"imagesX"`    // 图片x坐标
	ImagesY    int    `json:"imagesY"`    // 图片y坐标
	ImagesW    int    `json:"imagesW"`    // 图片宽度
	ImagesH    int    `json:"imagesH"`    // 图片高度
}

func ScreenShot(context *gin.Context) {

	var screenShotImages ScreenShotImages
	err := context.ShouldBindJSON(&screenShotImages)
	if err != nil {
		autoLog.Sugar.Errorf("ScreenShot 解析JSON失败: %v", err)
		context.JSON(400, "ScreenShot参数错误")
		return
	}

	err = control.ScreenShot("./img/abgi/" + screenShotImages.ImagesName + ".jpg")
	if err != nil {
		autoLog.Sugar.Errorf("截图失败: %v", err)
		context.JSON(400, "截图失败")
		return
	}

	err = Ocr.CropImage("./img/abgi/"+screenShotImages.ImagesName+".jpg", "./img/abgi/"+screenShotImages.ImagesName+"1.jpg", screenShotImages.ImagesX, screenShotImages.ImagesY, screenShotImages.ImagesW, screenShotImages.ImagesH)
	if err != nil {
		autoLog.Sugar.Errorf("裁剪图片失败: %v", err)
	}
}
