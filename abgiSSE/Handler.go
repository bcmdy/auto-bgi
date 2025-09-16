package abgiSSE

import (
	"auto-bgi/Notice"
	"auto-bgi/autoLog"
	"encoding/base64"
	"fmt"
	"os"
	"time"
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

// artifactsGroupPurchasing 联机启动
func artifactsGroupPurchasing(info Information) {
	//转成map
	xiaoxi := ""
	var dd []map[string]interface{}
	var names []string
	for _, v := range info.AA {
		dd = append(dd, map[string]interface{}{
			"ID":   v.ID,
			"UID":  v.UID,
			"Name": v.Name,
		})
		xiaoxi += fmt.Sprintf("序号%d -- %s\n", v.ID, v.Name)
		names = append(names, v.Name)
	}
	//生成图片
	if len(names) > 0 {
		//生成图片
		for _, name := range names {
			NameToImage(name)
		}
	}
	time.Sleep(500 * time.Millisecond)
	scriptGroupConfig.StartDogFoodOnline(RunDebug, dd)
	fmt.Println(xiaoxi)
	Notice.SentText(xiaoxi)
}
