package abgiSSE

import (
	"auto-bgi/config"
	"embed"
	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
)

//go:embed abgiFont
var abgiFont embed.FS

//func NameToImage(name string) {
//	// 读取 TTF 字体文件
//	fontBytes, err := abgiFont.ReadFile("abgiFont/HYW.ttf")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// 解析字体集合
//	collection, err := opentype.ParseCollection(fontBytes)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// 选择第 0 个字体
//	tt, err := collection.Font(0)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// 创建字体 Face，固定大小
//	fontSize := 23.0
//	face, err := opentype.NewFace(tt, &opentype.FaceOptions{
//		Size:    fontSize,
//		DPI:     72,
//		Hinting: font.HintingFull,
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// 测量文字宽度
//	dummyImg := image.NewRGBA(image.Rect(0, 0, 1, 1))
//	d := &font.Drawer{
//		Dst:  dummyImg,
//		Src:  image.NewUniform(color.RGBA{56, 61, 85, 255}), // 蓝色文字
//		Face: face,
//	}
//	textWidth := d.MeasureString(name).Round()
//
//	// 图片宽度自适应文字 + 边距，固定高度 29
//	paddingX := 8
//	width := textWidth + 4*paddingX
//	height := 50
//
//	img := image.NewRGBA(image.Rect(0, 0, width, height))
//	// 背景颜色
//	//draw.Draw(img, img.Bounds(), &image.Uniform{color.RGBA{241, 241, 240, 255}}, image.Point{}, draw.Src)
//
//	// 设置绘制器
//	d.Dst = img
//	// 垂直居中
//	y := (height+int(fontSize))/2 - 4 // 微调
//	d.Dot = fixed.P(paddingX, y)
//	d.DrawString(name)
//
//	// 保存 PNG
//	outFile, err := os.Create(config.Cfg.BetterGIAddress + "/User/JsScript/ArtifactsGroupPurchasing/targets/" + name + ".png")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer outFile.Close()
//
//	if err := png.Encode(outFile, img); err != nil {
//		log.Fatal(err)
//	}
//
//	log.Println("已生成 PNG:", name+".png")
//}

func NameToImage(name string) {
	fontBytes, err := abgiFont.ReadFile("abgiFont/HYW.ttf")
	if err != nil {
		log.Fatal(err)
	}

	collection, err := opentype.ParseCollection(fontBytes)
	if err != nil {
		log.Fatal(err)
	}

	tt, err := collection.Font(0)
	if err != nil {
		log.Fatal(err)
	}

	fontSize := 24.0 // 稍微大一点，更贴近截图比例
	face, err := opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    fontSize,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

	dummyImg := image.NewRGBA(image.Rect(241, 241, 239, 1))
	d := &font.Drawer{
		Dst:  dummyImg,
		Src:  image.NewUniform(color.RGBA{80, 87, 105, 255}),
		Face: face,
	}
	textWidth := d.MeasureString(name).Round()

	paddingX := 3
	width := textWidth + 1*paddingX
	height := 28

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	bg := color.RGBA{245, 246, 247, 255}
	draw.Draw(img, img.Bounds(), &image.Uniform{bg}, image.Point{}, draw.Src)

	d.Dst = img
	y := (height+int(fontSize))/2 - 4
	d.Dot = fixed.P(paddingX, y)
	d.DrawString(name)

	outFile, err := os.Create(config.Cfg.BetterGIAddress + "/User/JsScript/ArtifactsGroupPurchasing/targets/" + name + ".png")
	if err != nil {
		log.Fatal(err)
	}
	defer outFile.Close()

	if err := png.Encode(outFile, img); err != nil {
		log.Fatal(err)
	}

	log.Println("已生成 PNG:", name+".png")
	NameToImageMark(name)
}

func NameToImageMark(name string) {
	fontBytes, err := abgiFont.ReadFile("abgiFont/HYW.ttf")
	if err != nil {
		log.Fatal(err)
	}

	collection, err := opentype.ParseCollection(fontBytes)
	if err != nil {
		log.Fatal(err)
	}

	tt, err := collection.Font(0)
	if err != nil {
		log.Fatal(err)
	}

	fontSize := 24.0 // 稍微大一点，更贴近截图比例
	face, err := opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    fontSize,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

	dummyImg := image.NewRGBA(image.Rect(241, 241, 239, 1))
	d := &font.Drawer{
		Dst:  dummyImg,
		Src:  image.NewUniform(color.RGBA{100, 119, 171, 255}),
		Face: face,
	}
	textWidth := d.MeasureString(name).Round()

	paddingX := 3
	width := textWidth + 1*paddingX
	height := 28

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	bg := color.RGBA{245, 246, 247, 255}
	draw.Draw(img, img.Bounds(), &image.Uniform{bg}, image.Point{}, draw.Src)

	d.Dst = img
	y := (height+int(fontSize))/2 - 4
	d.Dot = fixed.P(paddingX, y)
	d.DrawString(name)

	outFile, err := os.Create(config.Cfg.BetterGIAddress + "/User/JsScript/ArtifactsGroupPurchasing/targets/" + name + "备注.png")
	if err != nil {
		log.Fatal(err)
	}
	defer outFile.Close()

	if err := png.Encode(outFile, img); err != nil {
		log.Fatal(err)
	}

	log.Println("已生成备注 PNG:", name+"备注.png")
}
