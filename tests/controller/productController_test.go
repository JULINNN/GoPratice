package tests

import (
	"bytes"
	"encoding/json"
	"main/internal/controller"
	"main/internal/models"
	"main/internal/repository"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// 建立模擬產品服務
type MockProductService struct {
	mock.Mock
}

func (m *MockProductService) GetProducts() ([]models.Product, error) {
	args := m.Called()
	return args.Get(0).([]models.Product), args.Error(1)
}

func (m *MockProductService) GetProduct(id int64) (models.Product, error) {
	args := m.Called(id)
	return args.Get(0).(models.Product), args.Error(1)
}

func (m *MockProductService) CreateProduct(product models.Product) (models.Product, error) {
	args := m.Called(product)
	return args.Get(0).(models.Product), args.Error(1)
}

func (m *MockProductService) UpdateProduct(id int64, product models.Product) (models.Product, error) {
	args := m.Called(id, product)
	return args.Get(0).(models.Product), args.Error(1)
}

func (m *MockProductService) DeleteProduct(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

// 設置測試環境
func setupTestRouter(mockService *MockProductService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 使用測試用的 zap logger
	logger, _ := zap.NewDevelopment()

	// 創建控制器並註冊路由
	controller := controller.NewProducController(mockService, logger)
	controller.RegisterRoutes(router)

	return router
}

// 測試健康檢查端點
func TestHealthCheck(t *testing.T) {
	// 設置模擬服務和路由
	mockService := new(MockProductService)
	router := setupTestRouter(mockService)

	// 創建請求
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	resp := httptest.NewRecorder()

	// 執行請求
	router.ServeHTTP(resp, req)

	// 驗證結果
	assert.Equal(t, http.StatusOK, resp.Code)

	var response map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &response)

	assert.Nil(t, err)
	assert.Equal(t, "ok", response["status"])
	assert.NotNil(t, response["version"])
	assert.NotNil(t, response["timestamp"])
}

// 測試獲取所有產品
func TestGetProducts(t *testing.T) {
	// 設置模擬服務和路由
	mockService := new(MockProductService)
	router := setupTestRouter(mockService)

	// 模擬產品數據
	products := []models.Product{
		{SkuCode: "SKU001", SkuName: "產品 1", SkuAmount: 10},
		{SkuCode: "SKU002", SkuName: "產品 2", SkuAmount: 20},
	}

	// 設置模擬服務預期行為
	mockService.On("GetProducts").Return(products, nil)

	// 創建請求
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/products", nil)
	resp := httptest.NewRecorder()

	// 執行請求
	router.ServeHTTP(resp, req)

	// 驗證結果
	assert.Equal(t, http.StatusOK, resp.Code)

	var responseProducts []models.Product
	err := json.Unmarshal(resp.Body.Bytes(), &responseProducts)

	assert.Nil(t, err)
	assert.Len(t, responseProducts, 2)
	assert.Equal(t, "SKU001", responseProducts[0].SkuCode)
	assert.Equal(t, "SKU002", responseProducts[1].SkuCode)

	// 驗證模擬服務方法被調用
	mockService.AssertExpectations(t)
}

// 測試獲取單個產品
func TestGetProduct(t *testing.T) {
	// 設置模擬服務和路由
	mockService := new(MockProductService)
	router := setupTestRouter(mockService)

	// 模擬產品數據
	product := models.Product{
		SkuCode:   "SKU001",
		SkuName:   "產品 1",
		SkuAmount: 10,
	}

	// 設置模擬服務預期行為
	mockService.On("GetProduct", int64(1)).Return(product, nil)

	// 創建請求
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/products/1", nil)
	resp := httptest.NewRecorder()

	// 執行請求
	router.ServeHTTP(resp, req)

	// 驗證結果
	assert.Equal(t, http.StatusOK, resp.Code)

	var responseProduct models.Product
	err := json.Unmarshal(resp.Body.Bytes(), &responseProduct)

	assert.Nil(t, err)
	assert.Equal(t, "SKU001", responseProduct.SkuCode)
	assert.Equal(t, "產品 1", responseProduct.SkuName)
	assert.Equal(t, 10, responseProduct.SkuAmount)

	// 驗證模擬服務方法被調用
	mockService.AssertExpectations(t)
}

// 測試獲取不存在的產品
func TestGetProductNotFound(t *testing.T) {
	// 設置模擬服務和路由
	mockService := new(MockProductService)
	router := setupTestRouter(mockService)

	// 設置模擬服務預期行為 - 返回未找到錯誤
	mockService.On("GetProduct", int64(999)).Return(models.Product{}, repository.ErrProductNotFound)

	// 創建請求
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/products/999", nil)
	resp := httptest.NewRecorder()

	// 執行請求
	router.ServeHTTP(resp, req)

	// 驗證結果
	assert.Equal(t, http.StatusNotFound, resp.Code)

	var response controller.ErrorResponse
	err := json.Unmarshal(resp.Body.Bytes(), &response)

	assert.Nil(t, err)
	assert.Equal(t, "PRODUCT_NOT_FOUND", response.ErrorCode)

	// 驗證模擬服務方法被調用
	mockService.AssertExpectations(t)
}

// 測試創建產品
func TestCreateProduct(t *testing.T) {
	// 設置模擬服務和路由
	mockService := new(MockProductService)
	router := setupTestRouter(mockService)

	// 創建產品請求數據
	productInput := models.Product{
		SkuCode:   "SKU003",
		SkuName:   "新產品",
		SkuAmount: 15,
	}

	// 創建預期返回的產品（包含ID）
	productOutput := models.Product{
		SkuCode:   "SKU003",
		SkuName:   "新產品",
		SkuAmount: 15,
	}

	// 設置模擬服務預期行為
	mockService.On("CreateProduct", mock.Anything).Return(productOutput, nil)

	// 創建請求
	jsonBody, _ := json.Marshal(productInput)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/products", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	// 執行請求
	router.ServeHTTP(resp, req)

	// 驗證結果
	assert.Equal(t, http.StatusCreated, resp.Code)

	var responseProduct models.Product
	err := json.Unmarshal(resp.Body.Bytes(), &responseProduct)

	assert.Nil(t, err)
	assert.Equal(t, "SKU003", responseProduct.SkuCode)
	assert.Equal(t, "新產品", responseProduct.SkuName)
	assert.Equal(t, 15, responseProduct.SkuAmount)

	// 驗證模擬服務方法被調用
	mockService.AssertExpectations(t)
}

// 測試創建無效產品
func TestCreateInvalidProduct(t *testing.T) {
	// 設置模擬服務和路由
	mockService := new(MockProductService)
	router := setupTestRouter(mockService)

	// 創建無效產品數據（缺少必要字段）
	invalidProduct := models.Product{
		// 沒有 SkuCode，這是必需的
		SkuName:   "無效產品",
		SkuAmount: 15,
	}

	// 創建請求
	jsonBody, _ := json.Marshal(invalidProduct)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/products", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	// 執行請求
	router.ServeHTTP(resp, req)

	// 驗證結果
	assert.Equal(t, http.StatusBadRequest, resp.Code)

	var response controller.ErrorResponse
	err := json.Unmarshal(resp.Body.Bytes(), &response)

	assert.Nil(t, err)
	assert.Equal(t, "PRODUCT_VALIDATION_ERROR", response.ErrorCode)
}

// 測試更新產品
func TestUpdateProduct(t *testing.T) {
	// 設置模擬服務和路由
	mockService := new(MockProductService)
	router := setupTestRouter(mockService)

	// 更新產品請求數據
	productInput := models.Product{
		SkuCode:   "SKU001",
		SkuName:   "更新產品名稱",
		SkuAmount: 25,
	}

	// 創建預期返回的產品
	productOutput := models.Product{
		SkuCode:   "SKU001",
		SkuName:   "更新產品名稱",
		SkuAmount: 25,
	}

	// 設置模擬服務預期行為
	mockService.On("UpdateProduct", int64(1), mock.Anything).Return(productOutput, nil)

	// 創建請求
	jsonBody, _ := json.Marshal(productInput)
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/products/1", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	// 執行請求
	router.ServeHTTP(resp, req)

	// 驗證結果
	assert.Equal(t, http.StatusOK, resp.Code)

	var responseProduct models.Product
	err := json.Unmarshal(resp.Body.Bytes(), &responseProduct)

	assert.Nil(t, err)
	assert.Equal(t, "SKU001", responseProduct.SkuCode)
	assert.Equal(t, "更新產品名稱", responseProduct.SkuName)
	assert.Equal(t, 25, responseProduct.SkuAmount)

	// 驗證模擬服務方法被調用
	mockService.AssertExpectations(t)
}

// 測試刪除產品
func TestDeleteProduct(t *testing.T) {
	// 設置模擬服務和路由
	mockService := new(MockProductService)
	router := setupTestRouter(mockService)

	// 設置模擬服務預期行為
	mockService.On("DeleteProduct", int64(1)).Return(nil)

	// 創建請求
	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/products/1", nil)
	resp := httptest.NewRecorder()

	// 執行請求
	router.ServeHTTP(resp, req)

	// 驗證結果
	assert.Equal(t, http.StatusOK, resp.Code)

	var response map[string]string
	err := json.Unmarshal(resp.Body.Bytes(), &response)

	assert.Nil(t, err)
	assert.Equal(t, "產品已成功刪除", response["message"])

	// 驗證模擬服務方法被調用
	mockService.AssertExpectations(t)
}

// 測試刪除不存在的產品
func TestDeleteProductNotFound(t *testing.T) {
	// 設置模擬服務和路由
	mockService := new(MockProductService)
	router := setupTestRouter(mockService)

	// 設置模擬服務預期行為 - 返回未找到錯誤
	mockService.On("DeleteProduct", int64(999)).Return(repository.ErrProductNotFound)

	// 創建請求
	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/products/999", nil)
	resp := httptest.NewRecorder()

	// 執行請求
	router.ServeHTTP(resp, req)

	// 驗證結果
	assert.Equal(t, http.StatusNotFound, resp.Code)

	var response controller.ErrorResponse
	err := json.Unmarshal(resp.Body.Bytes(), &response)

	assert.Nil(t, err)
	assert.Equal(t, "PRODUCT_NOT_FOUND", response.ErrorCode)

	// 驗證模擬服務方法被調用
	mockService.AssertExpectations(t)
}
