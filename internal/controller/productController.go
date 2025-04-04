package controller

import (
	"errors"
	model "main/internal/models"
	"main/internal/repository"
	"main/internal/service"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ErrorResponse struct {
	ErrorCode    string `json:"error_code"`
	ErrorMessage string `json:"error_message"`
	RequestID    string `json:"request_id,omitempty"`
}

type ProductController struct {
	service service.ProductService
	logger  *zap.Logger
}

func NewProducController(service service.ProductService, logger *zap.Logger) *ProductController {
	return &ProductController{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes 註冊路由
func (h *ProductController) RegisterRoutes(router *gin.Engine) {

	router.GET("/health", h.HealthCheck)

	api := router.Group("/api/v1")
	{
		products := api.Group("/products")
		{
			products.GET("", h.GetProducts)
			products.GET("/:id", h.GetProduct)
			products.POST("", h.CreateProduct)
			products.PUT("/:id", h.UpdateProduct)
			products.DELETE("/:id", h.DeleteProduct)
		}
	}
}

// HealthCheck 健康檢查端點
func (h *ProductController) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"version":   "1.0.0",
		"timestamp": time.Now().Unix(),
	})
}

// GetProducts 獲取所有產品
func (h *ProductController) GetProducts(c *gin.Context) {
	products, err := h.service.GetProducts()
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "PRODUCT_FETCH_ERROR", "獲取產品列表失敗", c.GetHeader("X-Request-ID"))
		return
	}

	c.JSON(http.StatusOK, products)
}

// GetProduct 獲取單個產品
func (h *ProductController) GetProduct(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondWithError(c, http.StatusBadRequest, "INVALID_PRODUCT_ID", "無效的產品ID", requestID)
		return
	}

	product, err := h.service.GetProduct(id)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			respondWithError(c, http.StatusNotFound, "PRODUCT_NOT_FOUND", "產品未找到", requestID)
			return
		}

		respondWithError(c, http.StatusInternalServerError, "PRODUCT_FETCH_ERROR", "獲取產品失敗", requestID)
		return
	}

	c.JSON(http.StatusOK, product)
}

// CreateProduct 創建新產品
func (h *ProductController) CreateProduct(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")

	var input model.Product
	if err := c.ShouldBindJSON(&input); err != nil {
		respondWithError(c, http.StatusBadRequest, "INVALID_REQUEST_DATA", "無效的請求數據", requestID)
		return
	}

	// 基本驗證
	if err := validateProduct(input); err != nil {
		respondWithError(c, http.StatusBadRequest, "PRODUCT_VALIDATION_ERROR", err.Error(), requestID)
		return
	}

	product, err := h.service.CreateProduct(input)
	if err != nil {
		respondWithError(c, http.StatusInternalServerError, "PRODUCT_CREATE_ERROR", "創建產品失敗", requestID)
		return
	}

	c.JSON(http.StatusCreated, product)
}

func validateProduct(product model.Product) error {
	if product.SkuCode == "" {
		return errors.New("產品名稱不能為空")
	}

	if product.SkuAmount < 0 {
		return errors.New("產品庫存不能為負數")
	}

	return nil
}

func (h *ProductController) UpdateProduct(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondWithError(c, http.StatusBadRequest, "INVALID_PRODUCT_ID", "無效的產品ID", requestID)
		return
	}

	var input model.Product
	if err := c.ShouldBindJSON(&input); err != nil {
		respondWithError(c, http.StatusBadRequest, "INVALID_REQUEST_DATA", "無效的請求數據", requestID)
		return
	}

	// 基本驗證
	if err := validateProduct(input); err != nil {
		respondWithError(c, http.StatusBadRequest, "PRODUCT_VALIDATION_ERROR", err.Error(), requestID)
		return
	}

	product, err := h.service.UpdateProduct(id, input)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrProductNotFound):
			respondWithError(c, http.StatusNotFound, "PRODUCT_NOT_FOUND", "產品未找到", requestID)
			return
		default:
			respondWithError(c, http.StatusInternalServerError, "PRODUCT_UPDATE_ERROR", "更新產品失敗", requestID)
			return
		}
	}

	c.JSON(http.StatusOK, product)
}

func (h *ProductController) DeleteProduct(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondWithError(c, http.StatusBadRequest, "INVALID_PRODUCT_ID", "無效的產品ID", requestID)
		return
	}

	err = h.service.DeleteProduct(id)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			respondWithError(c, http.StatusNotFound, "PRODUCT_NOT_FOUND", "產品未找到", requestID)
			return
		}

		respondWithError(c, http.StatusInternalServerError, "PRODUCT_DELETE_ERROR", "刪除產品失敗", requestID)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "產品已成功刪除",
	})
}

func respondWithError(c *gin.Context, statusCode int, errorCode string, message string, requestID string) {
	c.JSON(statusCode, ErrorResponse{
		ErrorCode:    errorCode,
		ErrorMessage: message,
		RequestID:    requestID,
	})
}
