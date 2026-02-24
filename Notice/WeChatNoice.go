package Notice

import (
	"auto-bgi/autoLog"
	"auto-bgi/config"
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// 向企业微信发送通知（文本）
func sendWeChatNotification(content string) {

	// 通知内容
	message := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			//"content": "BetterGI 已经关闭:\n" + Config.Content + "/test",
			"content": content,
		},
	}
	jsonData, err := json.Marshal(message)
	if err != nil {
		autoLog.Sugar.Error("Error marshaling JSON:", err)
		return
	}

	client := &http.Client{}

	req, err := http.NewRequest("POST", config.Cfg.Notice.Wechat, bytes.NewBuffer(jsonData))
	if err != nil {

		autoLog.Sugar.Error("sendWeChatNotification Error creating request:", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		autoLog.Sugar.Error("sendWeChatNotification Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		autoLog.Sugar.Error("企业微信机器人配置错误:", resp.Status)

	} else {
		autoLog.Sugar.Info("企业微信机器人配置成功:", resp.Status)
	}
}

// 向企业微信发送通知（图片）
func sendWeChatImage(path string) error {

	//获取本地文件
	imageData, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Printf("Error reading image file: %v\n", err)
		return err
	}
	// 计算 Base64 编码
	base64Data := base64.StdEncoding.EncodeToString(imageData)

	// 计算 MD5 哈希
	md5Hash := md5.Sum(imageData)
	md5String := hex.EncodeToString(md5Hash[:])

	// 通知内容
	message := map[string]interface{}{
		"msgtype": "image",
		"image": map[string]string{
			"base64": base64Data,
			"md5":    md5String,
		},
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		autoLog.Sugar.Error("Error marshaling JSON:", err)
		return err
	}

	client := &http.Client{}

	req, err := http.NewRequest("POST", config.Cfg.Notice.Wechat, bytes.NewBuffer(jsonData))
	if err != nil {

		autoLog.Sugar.Error("sendWeChatImage Error creating request:", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {

		autoLog.Sugar.Error("Error sending request:", err)
		return err
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("微信通知：http error: %s", resp.Status)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("微信通知：unmarshal response error: %v, body=%s", err, string(respBody))
	}
	if code, ok := result["errcode"].(float64); ok && code != 0 {
		return fmt.Errorf("微信通知：wechat error: %v, errmsg: %v", result["errcode"], result["errmsg"])
	}

	return nil
}
