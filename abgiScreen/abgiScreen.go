package abgiScreen

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/nfnt/resize"
	"github.com/vcaesar/screenshot"
	"image"
	"image/jpeg"
	"log"
	"net/http"
	"sync"
	"time"
)

// 全局变量
var (
	clients    = make(map[*websocket.Conn]bool)
	clientsMux = sync.RWMutex{}
	captureCh  = make(chan []byte, 60) // 捕获->编码管道
)

// 高 FPS 屏幕捕获
func CaptureScreen() {
	// 直接捕获 1280x720 区域
	bounds := imageRect1280x720()
	fmt.Printf("捕获区域: %dx%d\n", bounds.Dx(), bounds.Dy())

	ticker := time.NewTicker(time.Second / 60) // 目标 60 FPS
	defer ticker.Stop()

	for range ticker.C {
		clientsMux.RLock()
		if len(clients) == 0 {
			clientsMux.RUnlock()
			time.Sleep(time.Second / 10)
			continue
		}
		clientsMux.RUnlock()

		img, err := screenshot.CaptureRect(bounds)
		if err != nil {
			log.Println("屏幕捕获失败:", err)
			continue
		}

		go func(img image.Image) {
			resized := resize.Resize(1280, 720, img, resize.Bilinear)
			buf := encodeJPEG(resized, 60) // 质量 60
			select {
			case captureCh <- buf:
			default:
			}
		}(img)
	}
}

// 返回 1280x720 捕获区域（左上角）
func imageRect1280x720() (rect image.Rectangle) {
	return image.Rect(0, 0, 1920, 1080)
}

// JPEG 编码
func encodeJPEG(img image.Image, quality int) []byte {
	buf := &bytes.Buffer{}
	jpeg.Encode(buf, img, &jpeg.Options{Quality: quality})
	return buf.Bytes()
}

// WebSocket 客户端管理
func HandleWebSocket(c *gin.Context) {
	conn, err := (&websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}).Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket Upgrade失败:", err)
		return
	}
	defer conn.Close()

	clientsMux.Lock()
	clients[conn] = true
	clientsMux.Unlock()
	fmt.Printf("客户端已连接，当前数: %d\n", len(clients))

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}

	clientsMux.Lock()
	delete(clients, conn)
	clientsMux.Unlock()
	fmt.Printf("客户端已断开，当前数: %d\n", len(clients))
}

//// 广播线程
//func BroadcastFrames() {
//	frameCount := 0
//	lastTime := time.Now()
//
//	for frame := range captureCh {
//		clientsMux.RLock()
//		for conn := range clients {
//			if err := conn.WriteMessage(websocket.BinaryMessage, frame); err != nil {
//				clientsMux.RUnlock()
//				clientsMux.Lock()
//				delete(clients, conn)
//				clientsMux.Unlock()
//				clientsMux.RLock()
//			}
//		}
//		clientsMux.RUnlock()
//
//		// FPS 统计
//		frameCount++
//		if time.Since(lastTime) >= 10*time.Second {
//			fps := float64(frameCount) / time.Since(lastTime).Seconds()
//			fmt.Printf("实际帧率: %.1f FPS, 客户端数: %d\n", fps, len(clients))
//			frameCount = 0
//			lastTime = time.Now()
//		}
//	}
//}

func BroadcastFrames() {
	frameCount := 0
	lastTime := time.Now()

	// 控制目标 FPS：25~30
	targetFPS := 28.0
	targetInterval := time.Duration(1e9 / targetFPS) // 纳秒间隔
	lastPush := time.Now()

	for frame := range captureCh {
		now := time.Now()
		if now.Sub(lastPush) < targetInterval {
			// 没到发送间隔，跳过这帧
			continue
		}
		lastPush = now

		// 广播给客户端
		clientsMux.RLock()
		for conn := range clients {
			if err := conn.WriteMessage(websocket.BinaryMessage, frame); err != nil {
				clientsMux.RUnlock()
				clientsMux.Lock()
				delete(clients, conn)
				clientsMux.Unlock()
				clientsMux.RLock()
			}
		}
		clientsMux.RUnlock()

		// FPS统计
		frameCount++
		if time.Since(lastTime) >= 10*time.Second {
			fps := float64(frameCount) / time.Since(lastTime).Seconds()
			fmt.Printf("实际推送帧率: %.1f FPS, 客户端数: %d\n", fps, len(clients))
			frameCount = 0
			lastTime = time.Now()
		}
	}
}
