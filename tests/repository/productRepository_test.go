package repository

import (
	"main/internal/models"
	"main/internal/repository"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 創建測試環境
func setupMockDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	// 創建 sqlmock
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	// 將普通 sql.DB 轉為 sqlx.DB
	db := sqlx.NewDb(mockDB, "sqlmock")

	return db, mock
}

// 測試獲取所有產品
func TestGetAll(t *testing.T) {
	// 設置模擬數據庫
	db, mock := setupMockDB(t)
	defer db.Close()

	// 創建儲存庫
	repo := repository.NewProductRepository(db)

	// 模擬數據庫返回的行
	rows := sqlmock.NewRows([]string{"id", "sku_code", "sku_name", "sku_amount", "expiration", "create_at", "update_at"}).
		AddRow(1, "SKU001", "產品 1", 10, "2023-12-31", time.Now().Format("2006-01-02 15:04:05"), time.Now().Format("2006-01-02 15:04:05")).
		AddRow(2, "SKU002", "產品 2", 20, "2024-12-31", time.Now().Format("2006-01-02 15:04:05"), time.Now().Format("2006-01-02 15:04:05"))

	// 設置 SQL 查詢預期
	mock.ExpectQuery("SELECT (.+) FROM products").WillReturnRows(rows)

	// 調用儲存庫方法
	products, err := repo.GetAll()

	// 驗證結果
	assert.NoError(t, err)
	assert.Len(t, products, 2)
	assert.Equal(t, "SKU001", products[0].SkuCode)
	assert.Equal(t, "SKU002", products[1].SkuCode)

	// 確保所有預期都被滿足
	assert.NoError(t, mock.ExpectationsWereMet())
}

// 測試獲取單個產品
func TestGetByID(t *testing.T) {
	// 設置模擬數據庫
	db, mock := setupMockDB(t)
	defer db.Close()

	// 創建儲存庫
	repo := repository.NewProductRepository(db)

	// 模擬數據庫返回的行
	row := sqlmock.NewRows([]string{"id", "sku_code", "sku_name", "sku_amount", "expiration", "create_at", "update_at"}).
		AddRow(1, "SKU001", "產品 1", 10, "2023-12-31", time.Now().Format("2006-01-02 15:04:05"), time.Now().Format("2006-01-02 15:04:05"))

	// 設置 SQL 查詢預期
	mock.ExpectQuery("SELECT (.+) FROM products WHERE id = \\$1").
		WithArgs(1).
		WillReturnRows(row)

	// 調用儲存庫方法
	product, err := repo.GetByID(1)

	// 驗證結果
	assert.NoError(t, err)
	assert.Equal(t, "SKU001", product.SkuCode)
	assert.Equal(t, "產品 1", product.SkuName)
	assert.Equal(t, 10, product.SkuAmount)

	// 確保所有預期都被滿足
	assert.NoError(t, mock.ExpectationsWereMet())
}

// 測試獲取不存在的產品
func TestGetByIDNotFound(t *testing.T) {
	// 設置模擬數據庫
	db, mock := setupMockDB(t)
	defer db.Close()

	// 創建儲存庫
	repo := repository.NewProductRepository(db)

	// 設置 SQL 查詢預期 - 返回空結果
	mock.ExpectQuery("SELECT (.+) FROM products WHERE id = \\$1").
		WithArgs(999).
		WillReturnRows(sqlmock.NewRows([]string{"id", "sku_code", "sku_name", "sku_amount", "expiration", "create_at", "update_at"}))

	// 調用儲存庫方法
	_, err := repo.GetByID(999)

	// 驗證結果
	assert.Error(t, err)
	assert.Equal(t, repository.ErrProductNotFound, err)

	// 確保所有預期都被滿足
	assert.NoError(t, mock.ExpectationsWereMet())
}

// 測試創建產品
func TestCreate(t *testing.T) {
	// 設置模擬數據庫
	db, mock := setupMockDB(t)
	defer db.Close()

	// 創建儲存庫
	repo := repository.NewProductRepository(db)

	// 創建產品輸入
	productInput := models.Product{
		SkuCode:    "SKU003",
		SkuName:    "新產品",
		SkuAmount:  15,
		Expiration: "2025-01-01",
	}

	// 模擬數據庫返回的行
	rows := sqlmock.NewRows([]string{"id", "sku_code", "sku_name", "sku_amount", "expiration"}).
		AddRow(3, "SKU003", "新產品", 15, "2025-01-01")

	// 設置 SQL 插入預期
	mock.ExpectQuery("INSERT INTO products").
		WithArgs(productInput.SkuCode, productInput.SkuName, productInput.SkuAmount, productInput.Expiration).
		WillReturnRows(rows)

	// 調用儲存庫方法
	product, err := repo.Create(productInput)

	// 驗證結果
	assert.NoError(t, err)
	assert.Equal(t, "SKU003", product.SkuCode)
	assert.Equal(t, "新產品", product.SkuName)
	assert.Equal(t, 15, product.SkuAmount)

	// 確保所有預期都被滿足
	assert.NoError(t, mock.ExpectationsWereMet())
}

// 測試更新產品
func TestUpdateNonBlank(t *testing.T) {
	// 設置模擬數據庫
	db, mock := setupMockDB(t)
	defer db.Close()

	// 創建儲存庫
	repo := repository.NewProductRepository(db)

	// 更新產品輸入
	productInput := models.Product{
		SkuCode:    "SKU001",
		SkuName:    "更新產品名稱",
		SkuAmount:  25,
		Expiration: "2024-06-30",
	}

	// 模擬數據庫返回的行
	rows := sqlmock.NewRows([]string{"id", "update_at", "sku_code", "sku_name", "sku_amount", "expiration"}).
		AddRow(1, time.Now(), "SKU001", "更新產品名稱", 25, "2024-06-30")

	// 設置 SQL 更新預期 - 使用更精確的匹配
	mock.ExpectQuery(`UPDATE products SET sku_code = \$1, sku_name = \$2, expiration = \$3, sku_amount = \$4, update_at = \$5 WHERE id = \$6 RETURNING id, update_at, sku_code, sku_name, sku_amount, expiration`).
		WithArgs(productInput.SkuCode, productInput.SkuName, productInput.Expiration, productInput.SkuAmount, sqlmock.AnyArg(), int64(1)).
		WillReturnRows(rows)

	// 調用儲存庫方法
	product, err := repo.UpdateNonBlank(1, productInput)

	// 驗證結果
	assert.NoError(t, err)
	assert.Equal(t, "SKU001", product.SkuCode)
	assert.Equal(t, "更新產品名稱", product.SkuName)
	assert.Equal(t, 25, product.SkuAmount)

	// 確保所有預期都被滿足
	assert.NoError(t, mock.ExpectationsWereMet())
}

// 測試刪除產品
func TestDelete(t *testing.T) {
	// 設置模擬數據庫
	db, mock := setupMockDB(t)
	defer db.Close()

	// 創建儲存庫
	repo := repository.NewProductRepository(db)

	// 設置 SQL 刪除預期
	mock.ExpectExec("DELETE FROM products WHERE id = \\$1").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// 調用儲存庫方法
	err := repo.Delete(1)

	// 驗證結果
	assert.NoError(t, err)

	// 確保所有預期都被滿足
	assert.NoError(t, mock.ExpectationsWereMet())
}

// 測試刪除不存在的產品
func TestDeleteNotFound(t *testing.T) {
	// 設置模擬數據庫
	db, mock := setupMockDB(t)
	defer db.Close()

	// 創建儲存庫
	repo := repository.NewProductRepository(db)

	// 設置 SQL 刪除預期 - 返回沒有影響的行
	mock.ExpectExec("DELETE FROM products WHERE id = \\$1").
		WithArgs(999).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// 調用儲存庫方法
	err := repo.Delete(999)

	// 驗證結果
	assert.Error(t, err)
	assert.Equal(t, repository.ErrProductNotFound, err)

	// 確保所有預期都被滿足
	assert.NoError(t, mock.ExpectationsWereMet())
}
