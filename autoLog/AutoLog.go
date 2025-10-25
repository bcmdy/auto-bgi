package autoLog

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Logger *zap.Logger
	Sugar  *zap.SugaredLogger
)

func Init() {
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		panic(err)
	}

	// 按日期生成日志文件，如 logs/2025-08-30.log
	writer, err := rotatelogs.New(
		filepath.Join(logDir, "%Y-%m-%d.log"),
		rotatelogs.WithLinkName(filepath.Join(logDir, "current.log")), // 创建软链指向最新日志
		rotatelogs.WithMaxAge(15*24*time.Hour),                        // 保留15天
		rotatelogs.WithRotationTime(24*time.Hour),                     // 每24小时切割一次
		rotatelogs.WithClock(rotatelogs.Local),                        // ✅ 使用本地时间
		rotatelogs.WithHandler(rotatelogs.HandlerFunc(func(e rotatelogs.Event) {
			println("日志切割了，新文件:", e.Type())
		})),
	)
	if err != nil {
		panic(err)
	}

	// 控制台日志编码器（带颜色）
	consoleEncoder := zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		TimeKey:       "time",
		LevelKey:      "level",
		NameKey:       "logger",
		MessageKey:    "msg",
		StacktraceKey: "stacktrace",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.CapitalColorLevelEncoder,
		EncodeTime:    zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05"),
		EncodeCaller:  zapcore.ShortCallerEncoder,
	})

	// 文件日志编码器（JSON，无颜色）
	fileEncoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		TimeKey:       "time",
		LevelKey:      "level",
		NameKey:       "logger",
		MessageKey:    "msg",
		StacktraceKey: "stacktrace",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.CapitalLevelEncoder,
		EncodeTime:    zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05"),
		EncodeCaller:  zapcore.ShortCallerEncoder,
	})

	// 输出到控制台（Debug及以上级别）和日志文件（Info及以上级别）
	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapcore.DebugLevel),
		zapcore.NewCore(fileEncoder, zapcore.AddSync(writer), zapcore.InfoLevel),
	)

	Logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.DPanicLevel))
	Sugar = Logger.Sugar()
}

func Sync() {
	_ = Logger.Sync()
}

// 根据日期查询日志,默认是今天的
// QueryLogs 查询日志，date为空时默认今天
func QueryLogs(date string) ([]string, error) {
	logDir := "logs"

	// 如果没有指定日期，默认今天
	var logFile string
	if date == "" {
		logFile = time.Now().Format("2006-01-02") + ".log"
	} else {
		// 简单校验日期格式
		if _, err := time.Parse("2006-01-02", date); err != nil {
			return nil, fmt.Errorf("日期格式错误，应为 YYYY-MM-DD")
		}
		logFile = date + ".log"
	}

	fullPath := filepath.Join(logDir, logFile)

	// 打开日志文件
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("打开日志文件失败: %w", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("读取日志失败: %w", err)
	}

	return lines, nil
}
