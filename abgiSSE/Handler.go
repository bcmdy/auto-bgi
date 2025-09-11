package abgiSSE

import (
	"auto-bgi/Notice"
	"auto-bgi/autoLog"
	"encoding/base64"
	"fmt"
	"os"
)

func HandleImg(info Information) {
	if info.Img.Base64Str != "" {
		// 这里 Img 是 Base64 数据，你可以保存成文件，或者直接通知客户端
		fileName := fmt.Sprintf(info.Img.Path)
		imgBytes, err := base64.StdEncoding.DecodeString(info.Img.Base64Str)
		if err != nil {
			autoLog.Sugar.Errorf("解码图片失败: %v", err)
		} else {
			err = os.WriteFile(fileName, imgBytes, 0644)
			if err != nil {
				autoLog.Sugar.Errorf("保存图片失败: %v", err)
			} else {
				autoLog.Sugar.Infof("收到一张图片，已保存为 %s", fileName)
				Notice.SentText(fmt.Sprintf("收到一张图片，已保存为 %s", fileName))
			}
		}
	} else {
		autoLog.Sugar.Warnf("收到 status=2 消息，但 Img 字段为空")
	}
}
