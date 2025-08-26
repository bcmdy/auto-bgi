package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

var (
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
	DebugLogger *log.Logger
)

func init() {
	// 创建日志文件
	logFile, err := os.OpenFile("mihoyo.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("无法创建日志文件:", err)
	}

	// 初始化日志记录器
	InfoLogger = log.New(logFile, "[INFO] ", log.Ldate|log.Ltime)
	ErrorLogger = log.New(logFile, "[ERROR] ", log.Ldate|log.Ltime)
	DebugLogger = log.New(logFile, "[DEBUG] ", log.Ldate|log.Ltime)

	// 同时输出到控制台和文件
	InfoLogger.SetOutput(os.Stdout)
	ErrorLogger.SetOutput(os.Stderr)
}

// Info 记录信息日志
func Info(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	InfoLogger.Println(message)
}

// Error 记录错误日志
func Error(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	ErrorLogger.Println(message)
}

// Debug 记录调试日志
func Debug(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	DebugLogger.Println(message)
}

// LogWithTime 记录带时间戳的日志
func LogWithTime(level, message string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	log.Printf("[%s] %s: %s", level, timestamp, message)
} 