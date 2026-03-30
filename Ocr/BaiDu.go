package Ocr

import (
	"auto-bgi/autoLog"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	OAuthTokenURL = "https://aip.baidubce.com/oauth/2.0/token"
	OcrURL        = "https://aip.baidubce.com/rest/2.0/ocr/v1/general"
)

// AccessTokenResponse: 从百度 oauth 返回的结构（只取关键信息）
type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	Error       string `json:"error"`
	ErrorDesc   string `json:"error_description"`
}

// OCR 返回中的 words_result 单项
type WordItem struct {
	Words string `json:"words"`
}

// OCRResponse: 只解析我们关心的字段
type OCRResponse struct {
	LogID       int64      `json:"log_id"`
	WordsResult []WordItem `json:"words_result"`
	ErrorMsg    string     `json:"error_msg"` // 某些错误会放这里
	ErrorCode   int        `json:"error_code"`
}

func getAccessToken(apiKey, secretKey string) (string, error) {
	if apiKey == "" || secretKey == "" {
		return "", errors.New("apiKey or secretKey is empty")
	}

	// 构造 token 请求 URL (GET)
	u, _ := url.Parse(OAuthTokenURL)
	q := u.Query()
	q.Set("grant_type", "client_credentials")
	q.Set("client_id", apiKey)
	q.Set("client_secret", secretKey)
	u.RawQuery = q.Encode()

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(u.String())
	if err != nil {
		return "", fmt.Errorf("请求 access_token 失败: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取 access_token 响应失败: %w", err)
	}

	var tokenResp AccessTokenResponse
	if err := json.Unmarshal(bodyBytes, &tokenResp); err != nil {
		return "", fmt.Errorf("解析 access_token 响应失败: %w, body: %s", err, string(bodyBytes))
	}

	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("未取得 access_token: %s %s", tokenResp.Error, tokenResp.ErrorDesc)
	}

	return tokenResp.AccessToken, nil
}

// readImageBase64 从本地文件读取并返回 base64 编码字符串（去掉换行）
func readImageBase64(path string) (string, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("读取图片失败: %w", err)
	}
	enc := base64.StdEncoding.EncodeToString(b)
	// 百度服务要求上传的 base64 不包含换行
	enc = strings.ReplaceAll(enc, "\n", "")
	enc = strings.ReplaceAll(enc, "\r", "")
	return enc, nil
}

func ocrAccurateBasic(accessToken, imageBase64 string) (*OCRResponse, error) {
	if accessToken == "" {
		return nil, errors.New("accessToken is empty")
	}
	// 请求地址带 access_token 参数
	reqURL := OcrURL + "?access_token=" + url.QueryEscape(accessToken)

	// POST body: application/x-www-form-urlencoded, param: image=<base64>
	form := url.Values{}
	form.Set("image", imageBase64)
	// 可选参数示例（按需打开）
	// form.Set("detect_direction", "true")

	reqBody := strings.NewReader(form.Encode())
	req, err := http.NewRequest("POST", reqURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("构造 OCR 请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("调用 OCR 接口失败: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取 OCR 响应失败: %w", err)
	}

	// 有时候返回不是 200 但 body 提供错误信息，仍然解析
	var ocrResp OCRResponse
	if err := json.Unmarshal(respBytes, &ocrResp); err != nil {
		// 返回原始响应，便于调试
		return nil, fmt.Errorf("解析 OCR 响应失败: %w, body: %s", err, string(respBytes))
	}

	// 检查 error_code / error_msg
	if ocrResp.ErrorCode != 0 || ocrResp.ErrorMsg != "" {
		return &ocrResp, fmt.Errorf("OCR 接口返回错误: code=%d msg=%s", ocrResp.ErrorCode, ocrResp.ErrorMsg)
	}

	return &ocrResp, nil
}

func ocr(path, apiKey, secretKey string) string {
	// 1. 获取 access_token
	fmt.Println("获取 access_token ...")
	token, err := getAccessToken(apiKey, secretKey)
	if err != nil {
		fmt.Printf("获取 token 失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("access_token 获取成功")

	// 2. 读取图片 => base64
	fmt.Println("读取并编码图片 ...")
	imageBase64, err := readImageBase64(path)
	if err != nil {
		fmt.Printf("图片编码失败: %v\n", err)
		os.Exit(1)
	}

	// 3. 调用 OCR 接口
	fmt.Println("调用 OCR 接口 ...")
	ocrResp, err := ocrAccurateBasic(token, imageBase64)
	if err != nil {
		// 仍然可能拿到部分解析结果，先打印错误与响应摘要
		fmt.Printf("OCR 请求出错: %v\n", err)
		if ocrResp != nil && len(ocrResp.WordsResult) > 0 {
			fmt.Println("但部分结果如下：")
			for i, w := range ocrResp.WordsResult {

				fmt.Printf("%d: %s\n", i+1, w.Words)

			}
		}
		os.Exit(1)
	}

	// 4. 打印识别结果
	autoLog.Sugar.Info("识别结果:")
	var buf bytes.Buffer
	for _, w := range ocrResp.WordsResult {
		buf.WriteString(w.Words)
	}
	autoLog.Sugar.Info(buf.String())
	return buf.String()
}

func CropImage(srcPath, dstPath string, x, y, w, h int) error {
	// 打开源图片
	file, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("打开图片失败: %v", err)
	}
	defer file.Close()

	// 自动识别图片格式
	img, format, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("解码图片失败: %v", err)
	}
	fmt.Println("图片格式:", format)

	// 图片的矩形区域
	rect := image.Rect(x, y, x+w, y+h)
	// 创建新图像
	cropped := image.NewRGBA(rect)
	draw.Draw(cropped, rect, img, image.Point{X: x, Y: y}, draw.Src)

	// 创建输出文件
	out, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("创建输出文件失败: %v", err)
	}
	defer out.Close()

	// 根据格式保存
	switch format {
	case "jpeg":
		err = jpeg.Encode(out, cropped, nil)
	case "png":
		err = png.Encode(out, cropped)
	default:
		return fmt.Errorf("不支持的图片格式: %s", format)
	}
	if err != nil {
		return fmt.Errorf("保存裁剪图片失败: %v", err)
	}

	fmt.Println("裁剪成功:", dstPath)
	return nil
}

func BaiDuOcr(apiKey, secretKey string) (string, error) {
	err := CropImage("./img/abgi/baiDu.jpg", "./img/abgi/baiDuCrop.jpg", 1534, 889, 110, 60)
	if err != nil {
		autoLog.Sugar.Errorf("裁剪图片失败: %v", err)
		return "截图识别", err
	}
	ocr := ocr("./img/abgi/baiDuCrop.jpg", apiKey, secretKey)
	return ocr, nil
}
