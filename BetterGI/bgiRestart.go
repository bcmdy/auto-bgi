package BetterGI

import (
	"auto-bgi/Notice"
	"auto-bgi/autoLog"
	"auto-bgi/control"
	"fmt"
	"github.com/vcaesar/screenshot"
	"image"
	"time"
)

var (
	// 血条区域坐标（适配1920*1080分辨率）
	regionX = 806 // 血条起始横坐标
	regionY = 995 // 血条起始纵坐标（靠近屏幕底部）
	regionW = 300 // 血条区域宽度
	regionH = 40  // 血条区域高度
)

var CheckRedBloodSum = 0

// 红血检测
func CheckRedBlood() {
	rect := image.Rect(
		regionX,
		regionY,
		regionX+regionW,
		regionY+regionH,
	)

	autoLog.Sugar.Infof("开始检测红色血条...")

	for {

		if hasRedBlood(rect) {
			autoLog.Sugar.Infof("检测到红血条。")
			// TODO: 检测到红色血条，执行相关操作
			control.PressKey("z")
			CheckRedBloodSum++
			Notice.SentText(fmt.Sprintf("检测到红血条，自动按Z键，第%d次", CheckRedBloodSum))
		}
		time.Sleep(1000 * time.Millisecond)
	}
}

// 检测区域内是否存在足够多红色像素
func hasRedBlood(rect image.Rectangle) bool {
	img, err := screenshot.CaptureRect(rect)
	if err != nil {
		return false
	}

	//file, _ := os.Create("./logs/" + time.Now().Format("2006-01-02-15-04-05") + ".png")
	//defer file.Close()
	//png.Encode(file, img)

	redCount := 0
	total := 0

	step := 2 // 抽样步长

	for y := 0; y < img.Bounds().Dy(); y += step {
		for x := 0; x < img.Bounds().Dx(); x += step {
			r, g, b, _ := img.At(x, y).RGBA()
			r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(b>>8)

			if isRed(r8, g8, b8) {
				redCount++
			}
			total++
		}
	}

	ratio := float64(redCount) / float64(total)

	// 调试用（你可以先开着）
	//fmt.Printf("red ratio: %.2f\n", ratio)

	// 红色占比阈值（血条非常合适）
	//fmt.Println(ratio >= 0.05)
	return ratio > 0.01
}

// 红色判定规则（为血条调优）
func isRed(r, g, b uint8) bool {
	//fmt.Println(r, g, b)

	return r == 255 && g == 90 && b == 90

	//return r >= 201 &&
	//	g >= 60 && g <= 120 &&
	//	b >= 60 && b <= 120

}
