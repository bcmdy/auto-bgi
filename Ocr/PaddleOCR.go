package Ocr

import (
	"auto-bgi/Notice"
	"auto-bgi/autoLog"
	"auto-bgi/control"
	"fmt"
	"image"
	"image/color"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	Paddle "github.com/getcharzp/go-ocr"
	"github.com/vcaesar/screenshot"
)

var (
	regionX = 806
	regionY = 995
	regionW = 300
	regionH = 40
)

var hpPattern = regexp.MustCompile(`^\d+/\d+$`)

func PaddleOCR() {

	config := Paddle.Config{
		OnnxRuntimeLibPath: "./lib/onnxruntime.dll",
		DetModelPath:       "./paddle_weights/det.onnx",
		RecModelPath:       "./paddle_weights/rec.onnx",
		DictPath:           "./paddle_weights/dict.txt",
	}

	engine, err := Paddle.NewPaddleOcrEngine(config)
	if err != nil {
		autoLog.Sugar.Errorf("OCR 初始化失败: %v", err)
		return
	}
	defer engine.Destroy()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	autoLog.Sugar.Infof("血条识别启动")

	for range ticker.C {

		text, err := captureAndOCR(engine)
		if err != nil {
			log.Println("识别失败:", err)
			autoLog.Sugar.Errorf("识别失败:%s", err)
			continue
		}

		if hpPattern.MatchString(text) {
			//log.Println("血量:", text)
			split := strings.Split(text, "/")
			if len(split) == 2 {
				a, _ := strconv.ParseInt(split[0], 10, 64)
				b, _ := strconv.ParseInt(split[1], 10, 64)
				//除于
				res := fmt.Sprintf("%.2f", float64(a)/float64(b)*100)

				if float64(a)/float64(b)*100 < 30 {
					autoLog.Sugar.Infof("血条：%s", text)

					autoLog.Sugar.Infof("血量百分比:%s，血量低于:%s，自动回血z", res, "30%")
					control.PressKey("z")
					Notice.SentText(fmt.Sprintf("血量百分比:%s，血量低于:%s，自动回血z", res, "30%"))
				}
			}
		} else {
			//log.Println("异常识别:", text)
		}
	}
}

func captureAndOCR(engine Paddle.Engine) (string, error) {

	img, err := screenshot.Capture(regionX, regionY, regionW, regionH)
	if err != nil {
		return "", err
	}

	// 图像预处理
	processed := preprocess(img)

	results, err := engine.RunOCR(processed)
	if err != nil {
		return "", err
	}

	var builder strings.Builder
	for _, r := range results {
		builder.WriteString(r.Text)
	}

	text := strings.TrimSpace(builder.String())
	text = strings.ReplaceAll(text, "\n", "")
	text = fixHPFormat(text)

	return text, nil
}

func fixHPFormat(text string) string {

	text = filterNumber(text)

	// 如果已经正确
	if strings.Count(text, "/") == 1 {
		return text
	}

	// 如果没有斜杠，自动插入
	if strings.Count(text, "/") == 0 {

		// 找一个合理分割点
		if len(text) >= 6 {
			mid := len(text) / 2
			return text[:mid] + "/" + text[mid:]
		}
	}

	return text
}

func filterNumber(input string) string {
	var b strings.Builder
	for _, r := range input {
		if (r >= '0' && r <= '9') || r == '/' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func preprocess(src image.Image) image.Image {

	bounds := src.Bounds()
	grayImg := image.NewGray(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {

			r, _, _, _ := src.At(x, y).RGBA()

			// 只用红色通道增强
			gray := uint8(r >> 8)

			grayImg.Set(x, y, color.Gray{Y: gray})
		}
	}

	return grayImg
}
