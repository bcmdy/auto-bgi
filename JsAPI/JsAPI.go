package JsAPI

import (
	"auto-bgi/Ocr"
	"auto-bgi/abgiAi"
	"auto-bgi/autoLog"
	"auto-bgi/control"
	"github.com/gin-gonic/gin"
)

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

func AbgiAiConversation(context *gin.Context) {
	//query := context.Query("ask")
	var query struct {
		Ask string `json:"ask"`
	}
	err := context.ShouldBindJSON(&query)

	conversation, err := abgiAi.JsConversation(query.Ask)
	if err != nil {
		autoLog.Sugar.Errorf("JsConversation 失败: %v", err)
		return
	}
	context.JSON(200, conversation)
}
