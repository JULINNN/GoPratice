package tests

import (
	"errors"
	"main/internal/models"
	"main/internal/repository"
	"main/internal/service"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// 模擬儲存庫
type MockProductRepository struct {
	mock.Mock
}

func (m *MockProductRepository) GetAll() ([]models.Product, error) {
	args := m.Called()
	return args.Get(0).([]models.Product), args.Error(1)
}

func (m *MockProductRepository) GetByID(id int64) (models.Product, error) {
	args := m.Called(id)
	return args.Get(0).(models.Product), args.Error(1)
}

func (m *MockProductRepository) Create(product models.Product) (models.Product, error) {
	args := m.Called(product)
	return args.Get(0).(models.Product), args.Error(1)
}

func (m *MockProductRepository) UpdateNonBlank(id int64, product models.Product) (models.Product, error) {
	args := m.Called(id, product)
	return args.Get(0).(models.Product), args.Error(1)
}

func (m *MockProductRepository) Delete(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

// 測試獲取所有產品
func TestGetProducts(t *testing.T) {
	// 創建模擬儲存庫
	mockRepo := new(MockProductRepository)

	// 創建產品服務
	service := service.NewProductService(mockRepo)

	// 模擬產品數據
	expectedProducts := []models.Product{
		{SkuCode: "SKU001", SkuName: "產品 1", SkuAmount: 10},
		{SkuCode: "SKU002", SkuName: "產品 2", SkuAmount: 20},
	}

	// 設置模擬儲存庫預期行為
	mockRepo.On("GetAll").Return(expectedProducts, nil)

	// 調用服務方法
	products, err := service.GetProducts()

	// 驗證結果
	assert.Nil(t, err)
	assert.Equal(t, expectedProducts, products)

	// 驗證模擬儲存庫方法被調用
	mockRepo.AssertExpectations(t)
}

// 測試獲取所有產品 - 發生錯誤
func TestGetProductsError(t *testing.T) {
	// 創建模擬儲存庫
	mockRepo := new(MockProductRepository)

	// 創建產品服務
	service := service.NewProductService(mockRepo)

	// 設置模擬儲存庫預期行為 - 返回錯誤
	expectedError := errors.New("資料庫連接錯誤")
	mockRepo.On("GetAll").Return([]models.Product{}, expectedError)

	// 調用服務方法
	products, err := service.GetProducts()

	// 驗證結果
	assert.Equal(t, expectedError, err)
	assert.Empty(t, products)

	// 驗證模擬儲存庫方法被調用
	mockRepo.AssertExpectations(t)
}

// 測試獲取單個產品
func TestGetProduct(t *testing.T) {
	// 創建模擬儲存庫
	mockRepo := new(MockProductRepository)

	// 創建產品服務
	service := service.NewProductService(mockRepo)

	// 模擬產品數據
	expectedProduct := models.Product{
		SkuCode:   "SKU001",
		SkuName:   "產品 1",
		SkuAmount: 10,
	}

	// 設置模擬儲存庫預期行為
	mockRepo.On("GetByID", int64(1)).Return(expectedProduct, nil)

	// 調用服務方法
	product, err := service.GetProduct(1)

	// 驗證結果
	assert.Nil(t, err)
	assert.Equal(t, expectedProduct, product)

	// 驗證模擬儲存庫方法被調用
	mockRepo.AssertExpectations(t)
}

// 測試獲取不存在的產品
func TestGetProductNotFound(t *testing.T) {
	// 創建模擬儲存庫
	mockRepo := new(MockProductRepository)

	// 創建產品服務
	service := service.NewProductService(mockRepo)

	// 設置模擬儲存庫預期行為 - 返回未找到錯誤
	mockRepo.On("GetByID", int64(999)).Return(models.Product{}, repository.ErrProductNotFound)

	// 調用服務方法
	product, err := service.GetProduct(999)

	// 驗證結果
	assert.Equal(t, repository.ErrProductNotFound, err)
	assert.Empty(t, product)

	// 驗證模擬儲存庫方法被調用
	mockRepo.AssertExpectations(t)
}

// 測試創建產品
func TestCreateProduct(t *testing.T) {
	// 創建模擬儲存庫
	mockRepo := new(MockProductRepository)

	// 創建產品服務
	service := service.NewProductService(mockRepo)

	// 創建產品輸入和預期輸出
	productInput := models.Product{
		SkuCode:   "SKU003",
		SkuName:   "新產品",
		SkuAmount: 15,
	}

	expectedProduct := models.Product{
		SkuCode:   "SKU003",
		SkuName:   "新產品",
		SkuAmount: 15,
	}

	// 設置模擬儲存庫預期行為
	mockRepo.On("Create", productInput).Return(expectedProduct, nil)

	// 調用服務方法
	product, err := service.CreateProduct(productInput)

	// 驗證結果
	assert.Nil(t, err)
	assert.Equal(t, expectedProduct, product)

	// 驗證模擬儲存庫方法被調用
	mockRepo.AssertExpectations(t)
}

// 測試創建產品 - 發生錯誤
func TestCreateProductError(t *testing.T) {
	// 創建模擬儲存庫
	mockRepo := new(MockProductRepository)

	// 創建產品服務
	service := service.NewProductService(mockRepo)

	// 創建產品輸入
	productInput := models.Product{
		SkuCode:   "SKU003",
		SkuName:   "新產品",
		SkuAmount: 15,
	}

	// 設置模擬儲存庫預期行為 - 返回錯誤
	expectedError := errors.New("資料庫錯誤")
	mockRepo.On("Create", productInput).Return(models.Product{}, expectedError)

	// 調用服務方法
	product, err := service.CreateProduct(productInput)

	// 驗證結果
	assert.Equal(t, expectedError, err)
	assert.Empty(t, product)

	// 驗證模擬儲存庫方法被調用
	mockRepo.AssertExpectations(t)
}

// 測試更新產品
func TestUpdateProduct(t *testing.T) {
	// 創建模擬儲存庫
	mockRepo := new(MockProductRepository)

	// 創建產品服務
	service := service.NewProductService(mockRepo)

	// 更新前的產品
	existingProduct := models.Product{
		SkuCode:   "SKU001",
		SkuName:   "原產品",
		SkuAmount: 10,
	}

	// 更新後的產品
	updatedProduct := models.Product{
		SkuCode:   "SKU001",
		SkuName:   "更新產品名稱",
		SkuAmount: 25,
	}

	// 更新輸入
	updateInput := models.Product{
		SkuCode:   "SKU001",
		SkuName:   "更新產品名稱",
		SkuAmount: 25,
	}

	// 設置模擬儲存庫預期行為
	mockRepo.On("GetByID", int64(1)).Return(existingProduct, nil)
	mockRepo.On("UpdateNonBlank", int64(1), updateInput).Return(updatedProduct, nil)

	// 調用服務方法
	product, err := service.UpdateProduct(1, updateInput)

	// 驗證結果
	assert.Nil(t, err)
	assert.Equal(t, updatedProduct, product)

	// 驗證模擬儲存庫方法被調用
	mockRepo.AssertExpectations(t)
}

// 測試更新不存在的產品
func TestUpdateProductNotFound(t *testing.T) {
	// 創建模擬儲存庫
	mockRepo := new(MockProductRepository)

	// 創建產品服務
	service := service.NewProductService(mockRepo)

	// 更新輸入
	updateInput := models.Product{
		SkuCode:   "SKU999",
		SkuName:   "不存在的產品",
		SkuAmount: 5,
	}

	// 設置模擬儲存庫預期行為 - 返回未找到錯誤
	mockRepo.On("GetByID", int64(999)).Return(models.Product{}, repository.ErrProductNotFound)

	// 調用服務方法
	product, err := service.UpdateProduct(999, updateInput)

	// 驗證結果
	assert.Equal(t, repository.ErrProductNotFound, err)
	assert.Empty(t, product)

	// 驗證模擬儲存庫方法被調用
	mockRepo.AssertExpectations(t)

	// 確保 UpdateNonBlank 沒有被調用
	mockRepo.AssertNotCalled(t, "UpdateNonBlank")
}

// 測試刪除產品
func TestDeleteProduct(t *testing.T) {
	// 創建模擬儲存庫
	mockRepo := new(MockProductRepository)

	// 創建產品服務
	service := service.NewProductService(mockRepo)

	// 模擬產品數據
	existingProduct := models.Product{
		SkuCode:   "SKU001",
		SkuName:   "產品 1",
		SkuAmount: 10,
	}

	// 設置模擬儲存庫預期行為
	mockRepo.On("GetByID", int64(1)).Return(existingProduct, nil)
	mockRepo.On("Delete", int64(1)).Return(nil)

	// 調用服務方法
	err := service.DeleteProduct(1)

	// 驗證結果
	assert.Nil(t, err)

	// 驗證模擬儲存庫方法被調用
	mockRepo.AssertExpectations(t)
}

// 測試刪除不存在的產品
func TestDeleteProductNotFound(t *testing.T) {
	// 創建模擬儲存庫
	mockRepo := new(MockProductRepository)

	// 創建產品服務
	service := service.NewProductService(mockRepo)

	// 設置模擬儲存庫預期行為 - 返回未找到錯誤
	mockRepo.On("GetByID", int64(999)).Return(models.Product{}, repository.ErrProductNotFound)

	// 調用服務方法
	err := service.DeleteProduct(999)

	// 驗證結果
	assert.Equal(t, repository.ErrProductNotFound, err)

	// 驗證模擬儲存庫方法被調用
	mockRepo.AssertExpectations(t)

	// 確保 Delete 沒有被調用
	mockRepo.AssertNotCalled(t, "Delete")
}
