package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"main/internal/config"
	"main/internal/controller"
	"main/internal/logger"
	"main/internal/repository"
	"main/internal/service"
	"main/pkg/database"
)

func SetupApplication() (*gin.Engine, *config.AppConfig, error) {

	// 加載配置
	appConfig, err := config.LoadConfig()
	if err != nil {
		panic("無法加載配置: " + err.Error())
	}

	// 設置 Gin 模式
	gin.SetMode(appConfig.Server.Mode)

	// 初始化日誌
	loggerConfig := config.GetLoggerConfig(appConfig)
	appLogger, err := logger.InitLogger(loggerConfig)
	if err != nil {
		panic("無法初始化日誌: " + err.Error())
	}
	defer appLogger.Sync()

	appLogger.Info("應用程序啟動中",
		zap.String("mode", appConfig.Server.Mode),
		zap.Int("port", appConfig.Server.Port))

	db, err := database.NewPostgresDB(&appConfig.Database)
	if err != nil {
		return nil, nil, err
	}

	productRepository := repository.NewProductRepository(db)

	productService := service.NewProductService(productRepository)

	productController := controller.NewProducController(productService, appLogger)

	// 設置 Gin
	router := gin.New() // 使用 New() 而不是 Default()，因為我們將使用自定義日誌中間件

	// 添加恢復中間件，避免請求處理中的 panic
	router.Use(gin.Recovery())

	// 添加自定義的日誌中間件
	router.Use(logger.LoggerMiddleware(appLogger))

	// 註冊路由
	productController.RegisterRoutes(router)

	// 啟動服務器
	appLogger.Info("服務器啟動", zap.String("address", ":"+strconv.Itoa(appConfig.Server.Port)))
	if err := router.Run(":" + strconv.Itoa(appConfig.Server.Port)); err != nil {
		appLogger.Fatal("服務器啟動失敗", zap.Error(err))
	}

	return router, appConfig, nil
}

// NewServer 建立並啟動伺服器
func NewServer() error {

	router, cfg, err := SetupApplication()
	gin.SetMode(cfg.Server.Mode)
	if err != nil {
		return err
	}
	return router.Run(fmt.Sprintf(":%d", cfg.Server.Port))
}

// main 函数
func main() {
	if err := NewServer(); err != nil {
		log.Printf("伺服器啟動錯誤：%v", err)
		os.Exit(1)
	}
}
