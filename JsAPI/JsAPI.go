package JsAPI

import (
	"auto-bgi/Ocr"
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

// ScreenShot 处理截图请求的函数
// 接收一个gin上下文指针作为参数，用于处理HTTP请求和响应
func ScreenShot(context *gin.Context) {

	// 定义截图图片结构体变量
	var screenShotImages ScreenShotImages
	// 将请求中的JSON数据绑定到结构体变量中
	err := context.ShouldBindJSON(&screenShotImages)
	// 检查JSON解析是否出错
	if err != nil {
		// 记录JSON解析失败的错误日志
		autoLog.Sugar.Errorf("ScreenShot 解析JSON失败: %v", err)
		// 返回400状态码和错误信息
		context.JSON(400, "ScreenShot参数错误")
		return
	}

	// 调用控制层方法进行截图，传入图片保存路径
	err = control.ScreenShot("./img/abgi/" + screenShotImages.ImagesName + ".jpg")
	// 检查截图是否出错
	if err != nil {
		// 记录截图失败的错误日志
		autoLog.Sugar.Errorf("截图失败: %v", err)
		// 返回400状态码和错误信息
		context.JSON(400, "截图失败")
		return
	}

	// 调用OCR服务裁剪图片，传入原图路径、保存路径和裁剪参数
	err = Ocr.CropImage("./img/abgi/"+screenShotImages.ImagesName+".jpg", "./img/abgi/"+screenShotImages.ImagesName+"1.jpg", screenShotImages.ImagesX, screenShotImages.ImagesY, screenShotImages.ImagesW, screenShotImages.ImagesH)
	// 检查图片裁剪是否出错
	if err != nil {
		// 记录裁剪图片失败的错误日志
		autoLog.Sugar.Errorf("裁剪图片失败: %v", err)
	}
}

//func AbgiAiConversation(context *gin.Context) {
//	//query := context.Query("ask")
//	var query struct {
//		Ask string `json:"ask"`
//	}
//	err := context.ShouldBindJSON(&query)
//
//	conversation, err := abgiAi.JsConversation(query.Ask)
//	if err != nil {
//		autoLog.Sugar.Errorf("JsConversation 失败: %v", err)
//		return
//	}
//	context.JSON(200, conversation)
//}
