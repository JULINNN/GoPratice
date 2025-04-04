package logger

import (
	"bytes"
	"fmt"
	"log"
	"main/internal/config"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// InitLogger 根據配置初始化 zap 日誌記錄器
func InitLogger(logConfig *config.LoggerConfig) (*zap.Logger, error) {
	// 創建基本 zap 配置
	zapConfig := zap.NewProductionConfig()

	// 設置日誌級別
	level := getZapLogLevel(logConfig.Level)
	zapConfig.Level = zap.NewAtomicLevelAt(level)

	// 設置日誌格式
	if logConfig.Format == "console" {
		zapConfig.Encoding = "console"
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		zapConfig.Encoding = "json"
	}

	// 設置輸出路徑
	if logConfig.OutputPaths != "" {
		zapConfig.OutputPaths = stringToSlice(logConfig.OutputPaths)
	}
	if logConfig.ErrorOutputs != "" {
		zapConfig.ErrorOutputPaths = stringToSlice(logConfig.ErrorOutputs)
	}

	// 自定義時間格式
	zapConfig.EncoderConfig.TimeKey = "timestamp"
	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// 如果啟用日誌輪轉，使用自定義 Sink
	if logConfig.EnableRotate {
		return initLoggerWithRotation(logConfig, zapConfig)
	}

	// 確保日誌目錄存在
	ensureLogDirExists(zapConfig.OutputPaths)
	ensureLogDirExists(zapConfig.ErrorOutputPaths)

	return zapConfig.Build()
}

// LoggerMiddleware 創建一個用於記錄API執行時間和錯誤的中間件
func LoggerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 開始時間
		start := time.Now()

		// 獲取請求ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
			c.Header("X-Request-ID", requestID)
		}

		// 創建自定義的響應寫入器來捕獲狀態碼
		blw := &bodyLogWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		c.Writer = blw

		// 處理請求
		c.Next()

		// 計算執行時間
		duration := time.Since(start)

		// 獲取狀態碼
		statusCode := c.Writer.Status()

		// 只記錄成功請求的執行時間
		if statusCode < 400 {
			// 記錄API執行時間
			logger.Info("API執行完成",
				zap.String("request_id", requestID),
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.String("client_ip", c.ClientIP()),
				zap.Int("status", statusCode),
				zap.Duration("duration", duration),
				zap.String("duration_ms", fmt.Sprintf("%.2fms", float64(duration.Microseconds())/1000.0)),
			)
		} else {
			// 只在失敗時記錄錯誤日誌
			// 擷取響應體中的錯誤信息
			var responseBody string
			if len(blw.body.Bytes()) > 0 {
				responseBody = blw.body.String()
			}

			logger.Error("API執行失敗",
				zap.String("request_id", requestID),
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.String("client_ip", c.ClientIP()),
				zap.Int("status", statusCode),
				zap.Duration("duration", duration),
				zap.String("duration_ms", fmt.Sprintf("%.2fms", float64(duration.Microseconds())/1000.0)),
				zap.String("error_response", responseBody),
			)
		}
	}
}

// bodyLogWriter 是一個自定義的響應寫入器，用於捕獲響應體
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write 重寫Write方法，同時將響應體寫入緩衝區
func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// WriteString 重寫WriteString方法
func (w *bodyLogWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

// 以下是私有輔助函數

// initLoggerWithRotation 初始化具有文件輪轉功能的日誌記錄器
func initLoggerWithRotation(logConfig *config.LoggerConfig, zapConfig zap.Config) (*zap.Logger, error) {
	// 獲取輸出路徑
	outputPaths := stringToSlice(logConfig.OutputPaths)
	errorOutputPaths := stringToSlice(logConfig.ErrorOutputs)

	// 確保日誌目錄存在
	ensureLogDirExists(outputPaths)
	ensureLogDirExists(errorOutputPaths)

	// 創建自定義編碼器
	encoderConfig := zapConfig.EncoderConfig
	var encoder zapcore.Encoder
	if logConfig.Format == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// 收集要使用的核心
	cores := []zapcore.Core{}

	// 定義日誌級別
	level := getZapLogLevel(logConfig.Level)
	highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel && lvl >= level
	})
	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl < zapcore.ErrorLevel && lvl >= level
	})

	// 處理標準輸出/錯誤
	if contains(outputPaths, "stdout") {
		cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), lowPriority))
	}
	if contains(errorOutputPaths, "stderr") {
		cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(os.Stderr), highPriority))
	}

	// 處理文件輸出
	for _, path := range outputPaths {
		if path != "stdout" && path != "stderr" {
			writer := getLogWriter(path, logConfig.MaxSize, logConfig.MaxBackups,
				logConfig.MaxAge, logConfig.Compress)
			cores = append(cores, zapcore.NewCore(encoder, writer, lowPriority))
		}
	}

	for _, path := range errorOutputPaths {
		if path != "stdout" && path != "stderr" {
			writer := getLogWriter(path, logConfig.MaxSize, logConfig.MaxBackups,
				logConfig.MaxAge, logConfig.Compress)
			cores = append(cores, zapcore.NewCore(encoder, writer, highPriority))
		}
	}

	// 創建日誌記錄器
	core := zapcore.NewTee(cores...)
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1), zap.AddStacktrace(zapcore.ErrorLevel))

	return logger, nil
}

// getLogWriter 配置 lumberjack 的日誌輪轉
func getLogWriter(filename string, maxSize, maxBackups, maxAge int, compress bool) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		MaxAge:     maxAge,
		Compress:   compress,
	}
	return zapcore.AddSync(lumberJackLogger)
}

// stringToSlice 將字串轉換為字符串切片，用逗號分隔
func stringToSlice(s string) []string {
	if s == "" {
		return []string{}
	}
	parts := strings.Split(s, ",")
	var result []string
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// contains 檢查切片是否包含字符串
func contains(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// ensureLogDirExists 確保日誌目錄存在
func ensureLogDirExists(paths []string) {
	for _, path := range paths {
		if path != "stdout" && path != "stderr" {
			dir := filepath.Dir(path)
			if err := os.MkdirAll(dir, 0755); err != nil {
				// 使用標準庫日誌記錄錯誤，避免循環依賴
				log.Printf("無法創建日誌目錄 %s: %v", dir, err)
			}
		}
	}
}

// getZapLogLevel 將配置字串轉換為 zap 日誌級別
func getZapLogLevel(levelStr string) zapcore.Level {
	switch levelStr {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel // 默認使用 Info 級別
	}
}
