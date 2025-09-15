package abgiSSE

import (
	"auto-bgi/config"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"
)

// 解密
func Decrypt(encryptedText, key string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return "", errors.New("key length must be 16, 24, or 32 bytes")
	}

	cipherText, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", err
	}

	if len(cipherText) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}

	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(cipherText, cipherText)

	return string(cipherText), nil
}

func NameToImage(name string) {
	// 读取 TTF 字体文件
	fontBytes, err := os.ReadFile("font/HYW.ttf")
	if err != nil {
		log.Fatal(err)
	}

	// 解析字体集合
	collection, err := opentype.ParseCollection(fontBytes)
	if err != nil {
		log.Fatal(err)
	}

	// 选择第 0 个字体
	tt, err := collection.Font(0)
	if err != nil {
		log.Fatal(err)
	}

	// 创建字体 Face，固定大小
	fontSize := 28.0
	face, err := opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    fontSize,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

	// 测量文字宽度
	dummyImg := image.NewRGBA(image.Rect(0, 0, 1, 1))
	d := &font.Drawer{
		Dst:  dummyImg,
		Src:  image.NewUniform(color.RGBA{131, 171, 249, 255}), // 蓝色文字
		Face: face,
	}
	textWidth := d.MeasureString(name).Round()

	// 图片宽度自适应文字 + 边距，固定高度 29
	paddingX := 4
	width := textWidth + 2*paddingX
	height := 29

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	// 背景颜色
	draw.Draw(img, img.Bounds(), &image.Uniform{color.RGBA{131, 121, 85, 255}}, image.Point{}, draw.Src)

	// 设置绘制器
	d.Dst = img
	// 垂直居中
	y := (height+int(fontSize))/2 - 4 // 微调
	d.Dot = fixed.P(paddingX, y)
	d.DrawString(name)

	// 保存 PNG
	outFile, err := os.Create(config.Cfg.BetterGIAddress + "/User/JsScript/ArtifactsGroupPurchasing/targets/" + name + ".png")
	if err != nil {
		log.Fatal(err)
	}
	defer outFile.Close()

	if err := png.Encode(outFile, img); err != nil {
		log.Fatal(err)
	}

	log.Println("已生成 PNG:", name+".png")
}
