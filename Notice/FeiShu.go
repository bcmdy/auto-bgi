package Notice

import (
	"auto-bgi/config"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

// 飞书上传图片接口（机器人专用）
const FeishuImageUploadURL = "https://open.feishu.cn/open-apis/im/v1/images"

// SendFeishuTextMessage 发送纯文本消息
func sendFeishuTextMessage(content string) error {
	payload := map[string]interface{}{
		"msg_type": "text",
		"content": map[string]string{
			"text": content,
		},
	}

	data, _ := json.Marshal(payload)
	resp, err := http.Post(config.Cfg.Notice.FeiShu.FeiShuWebhookURL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("发送失败: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println("飞书响应:", string(body))
	return nil
}

// ---------------------- 图片消息 ----------------------

// SendFeishuImageMessage 发送图片消息（传入本地图片路径）
// 注意：机器人 webhook 发送图片必须先上传图片以获取 image_key。
func sendFeiShuImageMessage(imagePath string) error {

	appAccessToken := getTenantAccessToken(config.Cfg.Notice.FeiShu.AppID, config.Cfg.Notice.FeiShu.AppSecret)

	imageKey, err := uploadFeishuImage(imagePath, appAccessToken)
	if err != nil {
		return err
	}

	payload := map[string]interface{}{
		"msg_type": "image",
		"content": map[string]string{
			"image_key": imageKey,
		},
	}

	data, _ := json.Marshal(payload)
	resp, err := http.Post(config.Cfg.Notice.FeiShu.FeiShuWebhookURL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("发送图片失败: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println("飞书响应:", string(body))
	return nil
}

// uploadFeishuImage 上传图片到飞书并返回 image_key
func uploadFeishuImage(imagePath, appAccessToken string) (string, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return "", fmt.Errorf("无法打开图片: %v", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, _ := writer.CreateFormFile("image", imagePath)
	io.Copy(part, file)

	writer.WriteField("image_type", "message")
	writer.Close()

	req, _ := http.NewRequest("POST", FeishuImageUploadURL, body)
	req.Header.Set("Authorization", "Bearer "+appAccessToken)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("上传图片失败: %v", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	fmt.Println("上传响应:", string(respBody))

	var result struct {
		Code int `json:"code"`
		Data struct {
			ImageKey string `json:"image_key"`
		} `json:"data"`
		Msg string `json:"msg"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}
	if result.Code != 0 {
		return "", fmt.Errorf("上传失败: %v", result.Msg)
	}
	return result.Data.ImageKey, nil
}

func getTenantAccessToken(appID, appSecret string) string {
	url := "https://open.feishu.cn/open-apis/auth/v3/tenant_access_token/internal"
	payload := map[string]string{
		"app_id":     appID,
		"app_secret": appSecret,
	}
	data, _ := json.Marshal(payload)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Code              int    `json:"code"`
		Msg               string `json:"msg"`
		TenantAccessToken string `json:"tenant_access_token"`
		Expire            int    `json:"expire"`
	}
	json.Unmarshal(body, &result)
	if result.Code != 0 {
		panic(fmt.Sprintf("获取token失败: %v", result.Msg))
	}
	return result.TenantAccessToken
}
