package service

import (
	model "main/internal/models"
	"main/internal/repository"
)

// ProductService 定義產品服務接口
type ProductService interface {
	GetProducts() ([]model.Product, error)
	GetProduct(id int64) (model.Product, error)
	CreateProduct(input model.Product) (model.Product, error)
	UpdateProduct(id int64, input model.Product) (model.Product, error)
	DeleteProduct(id int64) error
}

// DefaultProductService 實現默認產品服務
type DefaultProductService struct {
	repo repository.ProductRepository
}

// NewProductService 創建新的產品服務
func NewProductService(repo repository.ProductRepository) ProductService {
	return &DefaultProductService{
		repo: repo,
	}
}

// GetProducts 獲取所有產品
func (s *DefaultProductService) GetProducts() ([]model.Product, error) {
	return s.repo.GetAll()
}

// GetProduct 獲取特定產品
func (s *DefaultProductService) GetProduct(id int64) (model.Product, error) {
	return s.repo.GetByID(id)
}

// CreateProduct 創建新產品
func (s *DefaultProductService) CreateProduct(input model.Product) (model.Product, error) {
	// 這裡可以添加業務邏輯，如庫存檢查、價格驗證等
	return s.repo.Create(input)
}

// UpdateProduct 更新產品
func (s *DefaultProductService) UpdateProduct(id int64, input model.Product) (model.Product, error) {
	// 先檢查產品是否存在
	_, err := s.repo.GetByID(id)
	if err != nil {
		return model.Product{}, err
	}

	return s.repo.UpdateNonBlank(id, input)
}

// DeleteProduct 刪除產品
func (s *DefaultProductService) DeleteProduct(id int64) error {
	// 先檢查產品是否存在
	_, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	return s.repo.Delete(id)
}
