package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"main/internal/config"
	"main/internal/controller"
	"main/internal/models"
	"main/internal/repository"
	"main/internal/service"
	"main/pkg/database"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type IntegrationTestSuite struct {
	suite.Suite
	db         *sqlx.DB
	router     *gin.Engine
	controller *controller.ProductController
	pool       *dockertest.Pool
	resource   *dockertest.Resource
}

// 設置測試套件 - 啟動 Docker PostgreSQL
func (s *IntegrationTestSuite) SetupSuite() {
	// 使用測試模式
	gin.SetMode(gin.TestMode)

	// 使用 dockertest 啟動臨時 PostgreSQL 容器
	var err error
	s.pool, err = dockertest.NewPool("")
	if err != nil {
		log.Fatalf("無法連接到 Docker: %s", err)
	}

	// 啟動 PostgreSQL 容器
	s.resource, err = s.pool.Run("postgres", "13", []string{
		"POSTGRES_PASSWORD=postgres",
		"POSTGRES_USER=postgres",
		"POSTGRES_DB=test_db",
	})
	if err != nil {
		log.Fatalf("無法啟動 PostgreSQL 容器: %s", err)
	}

	// 獲取容器連接端口
	portStr := s.resource.GetPort("5432/tcp")
	port, err := strconv.Atoi(portStr)

	// 創建數據庫配置
	dbConfig := config.DatabaseConfig{
		Host:     "localhost",
		Port:     port,
		User:     "postgres",
		Password: "postgres",
		DBName:   "test_db",
		SSLMode:  "disable",
	}

	// 嘗試連接到數據庫
	if err := s.pool.Retry(func() error {
		var err error
		s.db, err = database.NewPostgresDB(&dbConfig)
		if err != nil {
			return err
		}
		return s.db.Ping()
	}); err != nil {
		log.Fatalf("無法連接到數據庫: %s", err)
	}

	// 創建測試表
	s.createTestTables()

	// 設置應用依賴
	logger, _ := zap.NewDevelopment()
	productRepo := repository.NewProductRepository(s.db)
	productService := service.NewProductService(productRepo)
	s.controller = controller.NewProducController(productService, logger)

	// 設置路由
	s.router = gin.New()
	s.controller.RegisterRoutes(s.router)
}

// 創建測試表
func (s *IntegrationTestSuite) createTestTables() {
	// 創建產品表
	schema := `
    CREATE TABLE IF NOT EXISTS products (
        id SERIAL PRIMARY KEY,
        sku_code VARCHAR(50) NOT NULL,
        sku_name VARCHAR(100) NOT NULL,
        sku_amount INT NOT NULL DEFAULT 0,
        expiration VARCHAR(50),
        create_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        update_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );
    `
	_, err := s.db.Exec(schema)
	if err != nil {
		log.Fatalf("無法創建測試表: %s", err)
	}
}

// 測試每個方法前清理表數據
func (s *IntegrationTestSuite) SetupTest() {
	_, err := s.db.Exec("TRUNCATE TABLE products RESTART IDENTITY")
	if err != nil {
		log.Fatalf("無法清理測試表: %s", err)
	}
}

// 拆解測試套件 - 關閉 Docker 容器
func (s *IntegrationTestSuite) TearDownSuite() {
	if s.db != nil {
		s.db.Close()
	}

	if s.resource != nil {
		if err := s.pool.Purge(s.resource); err != nil {
			log.Fatalf("無法清理 Docker 資源: %s", err)
		}
	}
}

// 測試健康檢查端點
func (s *IntegrationTestSuite) TestHealthEndpoint() {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert.Equal(s.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "ok", response["status"])
}

// 測試創建產品端點
func (s *IntegrationTestSuite) TestCreateProduct() {
	// 創建產品請求
	productInput := models.Product{
		SkuCode:    "TEST001",
		SkuName:    "測試產品",
		SkuAmount:  100,
		Expiration: "2025-12-31",
	}

	jsonBody, _ := json.Marshal(productInput)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/products", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	// 驗證響應
	assert.Equal(s.T(), http.StatusCreated, w.Code)

	var response models.Product
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "TEST001", response.SkuCode)
	assert.Equal(s.T(), "測試產品", response.SkuName)

	assert.Equal(s.T(), 100, response.SkuAmount)
}

// 測試獲取所有產品端點
func (s *IntegrationTestSuite) TestGetProducts() {
	// 先添加一些測試數據
	s.insertTestProducts(3)

	// 發送請求
	req := httptest.NewRequest(http.MethodGet, "/api/v1/products", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	// 驗證響應
	assert.Equal(s.T(), http.StatusOK, w.Code)

	var products []models.Product
	err := json.Unmarshal(w.Body.Bytes(), &products)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), products, 3)
}

// 向測試資料庫插入測試產品
func (s *IntegrationTestSuite) insertTestProducts(count int) {
	for i := 0; i < count; i++ {
		_, err := s.db.Exec(`
			INSERT INTO products (sku_code, sku_name, sku_amount, expiration)
			VALUES ($1, $2, $3, $4)
		`, fmt.Sprintf("TEST%03d", i), fmt.Sprintf("測試產品 %d", i), 100+i, "2025-12-31")

		if err != nil {
			s.T().Fatalf("無法插入測試產品: %s", err)
		}
	}
}

// 測試獲取單個產品端點
func (s *IntegrationTestSuite) TestGetProduct() {
	// 添加測試數據
	s.insertTestProducts(1)

	// 發送請求
	req := httptest.NewRequest(http.MethodGet, "/api/v1/products/1", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	// 驗證響應
	assert.Equal(s.T(), http.StatusOK, w.Code)

	var product models.Product
	err := json.Unmarshal(w.Body.Bytes(), &product)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "TEST000", product.SkuCode)
}

// 測試獲取不存在的產品
func (s *IntegrationTestSuite) TestGetProductNotFound() {
	// 發送請求 - 尋找不存在的產品
	req := httptest.NewRequest(http.MethodGet, "/api/v1/products/999", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	// 驗證響應
	assert.Equal(s.T(), http.StatusNotFound, w.Code)

	var response controller.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "PRODUCT_NOT_FOUND", response.ErrorCode)
}

// 測試更新產品端點
func (s *IntegrationTestSuite) TestUpdateProduct() {
	// 添加測試數據
	s.insertTestProducts(1)

	// 更新請求
	updateInput := models.Product{
		SkuCode:    "TEST000",
		SkuName:    "已更新的測試產品",
		SkuAmount:  50,
		Expiration: "2026-06-30",
	}

	jsonBody, _ := json.Marshal(updateInput)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/products/1", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	// 驗證響應
	assert.Equal(s.T(), http.StatusOK, w.Code)

	var response models.Product
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "TEST000", response.SkuCode)
	assert.Equal(s.T(), "已更新的測試產品", response.SkuName)
	assert.Equal(s.T(), 50, response.SkuAmount)
}

// 測試刪除產品端點
func (s *IntegrationTestSuite) TestDeleteProduct() {
	// 添加測試數據
	s.insertTestProducts(1)

	// 發送刪除請求
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/products/1", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	// 驗證響應
	assert.Equal(s.T(), http.StatusOK, w.Code)

	// 驗證產品已被刪除
	var count int
	err := s.db.Get(&count, "SELECT COUNT(*) FROM products WHERE id = 1")
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), 0, count)
}

// 測試刪除不存在的產品
func (s *IntegrationTestSuite) TestDeleteProductNotFound() {
	// 發送刪除請求 - 嘗試刪除不存在的產品
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/products/999", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	// 驗證響應
	assert.Equal(s.T(), http.StatusNotFound, w.Code)

	var response controller.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "PRODUCT_NOT_FOUND", response.ErrorCode)
}

// 運行整合測試套件
func TestIntegrationSuite(t *testing.T) {
	// 跳過整合測試如果沒有 Docker 環境
	if os.Getenv("SKIP_INTEGRATION") != "" {
		t.Skip("跳過整合測試 (設置了 SKIP_INTEGRATION 環境變數)")
	}

	suite.Run(t, new(IntegrationTestSuite))
}
