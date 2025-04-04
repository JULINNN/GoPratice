package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
)

// AppConfig 應用程序配置結構
type AppConfig struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	Logger   LoggerConfig   `json:"logger"`
}

// ServerConfig 服務器配置
type ServerConfig struct {
	Port int    `json:"port"`
	Mode string `json:"mode"`
}

// DatabaseConfig 數據庫配置
type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
	SSLMode  string `json:"sslmode"`
}

// LoggerConfig 日誌配置
type LoggerConfig struct {
	Level        string `json:"level"`         // 日誌級別: debug, info, warn, error, fatal
	Format       string `json:"format"`        // 格式: json, console
	OutputPaths  string `json:"output_paths"`  // 輸出路徑，多個路徑用逗號分隔
	ErrorOutputs string `json:"error_outputs"` // 錯誤日誌輸出路徑
	EnableRotate bool   `json:"enable_rotate"` // 是否啟用日誌輪轉
	MaxSize      int    `json:"max_size"`      // 每個日誌文件的最大尺寸，單位MB
	MaxBackups   int    `json:"max_backups"`   // 保留舊文件的最大數量
	MaxAge       int    `json:"max_age"`       // 保留舊文件的最大天數
	Compress     bool   `json:"compress"`      // 是否壓縮舊文件
}

// DSN 獲取數據庫連接字符串
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

// LoadConfig 加載配置，優先使用環境變數，然後使用配置文件
func LoadConfig() (*AppConfig, error) {
	// 獲取默認配置
	config := DefaultConfig()

	// 嘗試從配置文件加載 (如果存在)
	configFile := os.Getenv("CONFIG_FILE")
	if configFile == "" {
		// 嘗試默認路徑
		possiblePaths := []string{"../../configs/config.json", "./configs/config.json", "./config.json"}
		for _, path := range possiblePaths {
			if _, err := os.Stat(path); err == nil {
				configFile = path
				break
			}
		}
	}

	// 如果找到配置文件，加載它
	if configFile != "" {
		log.Printf("發現配置文件: %s", configFile)
		if err := loadConfigFromFile(config, configFile); err != nil {
			log.Printf("從文件加載配置時出錯: %v, 使用默認值並繼續", err)
		}
	} else {
		log.Println("未找到配置文件，使用默認配置和環境變數")
	}

	// 使用環境變數覆蓋配置
	overrideWithEnv(config)

	// 記錄最終配置（排除敏感信息）
	logConfig(config)

	return config, nil
}

// DefaultConfig 返回默認配置
func DefaultConfig() *AppConfig {
	return &AppConfig{
		Server: ServerConfig{
			Port: 8080,
			Mode: "debug",
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "postgres",
			DBName:   "product_db",
			SSLMode:  "disable",
		},
		Logger: LoggerConfig{
			Level:        "info",
			Format:       "json",
			OutputPaths:  "stdout,./logs/app.log",
			ErrorOutputs: "stderr,./logs/error.log",
			EnableRotate: true,
			MaxSize:      10,
			MaxBackups:   5,
			MaxAge:       30,
			Compress:     true,
		},
	}
}

// loadConfigFromFile 從配置文件加載配置
func loadConfigFromFile(config *AppConfig, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("無法打開配置文件: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return fmt.Errorf("解析配置文件失敗: %w", err)
	}

	return nil
}

// overrideWithEnv 使用環境變數覆蓋配置
func overrideWithEnv(config *AppConfig) {
	// 服務器配置
	if port := getEnvAsInt("SERVER_PORT", 0); port != 0 {
		config.Server.Port = port
	}
	if mode := os.Getenv("GIN_MODE"); mode != "" {
		config.Server.Mode = mode
	}

	// 數據庫配置
	if host := os.Getenv("DB_HOST"); host != "" {
		config.Database.Host = host
	}
	if port := getEnvAsInt("DB_PORT", 0); port != 0 {
		config.Database.Port = port
	}
	if user := os.Getenv("DB_USER"); user != "" {
		config.Database.User = user
	}
	if password := os.Getenv("DB_PASSWORD"); password != "" {
		config.Database.Password = password
	}
	if dbName := os.Getenv("DB_NAME"); dbName != "" {
		config.Database.DBName = dbName
	}
	if sslMode := os.Getenv("DB_SSL_MODE"); sslMode != "" {
		config.Database.SSLMode = sslMode
	}

	// 日誌配置
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		config.Logger.Level = level
	}
	if format := os.Getenv("LOG_FORMAT"); format != "" {
		config.Logger.Format = format
	}
	if outputs := os.Getenv("LOG_OUTPUT_PATHS"); outputs != "" {
		config.Logger.OutputPaths = outputs
	}
	if errorOutputs := os.Getenv("LOG_ERROR_OUTPUTS"); errorOutputs != "" {
		config.Logger.ErrorOutputs = errorOutputs
	}
	if enableRotate := getEnvAsBool("LOG_ENABLE_ROTATE", config.Logger.EnableRotate); enableRotate != config.Logger.EnableRotate {
		config.Logger.EnableRotate = enableRotate
	}
	if maxSize := getEnvAsInt("LOG_MAX_SIZE", 0); maxSize > 0 {
		config.Logger.MaxSize = maxSize
	}
	if maxBackups := getEnvAsInt("LOG_MAX_BACKUPS", 0); maxBackups > 0 {
		config.Logger.MaxBackups = maxBackups
	}
	if maxAge := getEnvAsInt("LOG_MAX_AGE", 0); maxAge > 0 {
		config.Logger.MaxAge = maxAge
	}
	if compress := getEnvAsBool("LOG_COMPRESS", config.Logger.Compress); compress != config.Logger.Compress {
		config.Logger.Compress = compress
	}
}

// logConfig 記錄配置信息（排除敏感信息）
func logConfig(config *AppConfig) {
	log.Printf("服務器配置: 端口=%d, 模式=%s",
		config.Server.Port, config.Server.Mode)

	log.Printf("數據庫配置: 主機=%s, 端口=%d, 用戶=%s, 數據庫=%s, SSL模式=%s",
		config.Database.Host, config.Database.Port,
		config.Database.User, config.Database.DBName,
		config.Database.SSLMode)

	log.Printf("日誌配置: 級別=%s, 格式=%s, 輸出路徑=%s, 錯誤輸出=%s, 輪轉=%v",
		config.Logger.Level, config.Logger.Format,
		config.Logger.OutputPaths, config.Logger.ErrorOutputs,
		config.Logger.EnableRotate)
}

// 從環境變數獲取整數值
func getEnvAsInt(key string, defaultVal int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
		log.Printf("警告: 環境變數 %s 的值 '%s' 無法轉換為整數", key, value)
	}
	return defaultVal
}

// 從環境變數獲取布爾值
func getEnvAsBool(key string, defaultVal bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
		log.Printf("警告: 環境變數 %s 的值 '%s' 無法轉換為布爾值", key, value)
	}
	return defaultVal
}

// GetDatabaseConfig 從應用配置中提取數據庫配置
func GetDatabaseConfig(config *AppConfig) *DatabaseConfig {
	return &config.Database
}

// GetLoggerConfig 從應用配置中提取日誌配置
func GetLoggerConfig(config *AppConfig) *LoggerConfig {
	return &config.Logger
}
