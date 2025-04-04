package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"main/internal/models"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

// 錯誤定義
var (
	ErrProductNotFound = errors.New("產品未找到")
)

// ProductRepository 定義產品儲存庫接口
type ProductRepository interface {
	GetAll() ([]models.Product, error)
	GetByID(id int64) (models.Product, error)
	Create(input models.Product) (models.Product, error)
	UpdateNonBlank(id int64, input models.Product) (models.Product, error)
	Delete(id int64) error
}

type PostgresProductRepository struct {
	db *sqlx.DB
}

func NewProductRepository(db *sqlx.DB) ProductRepository {
	return &PostgresProductRepository{db: db}
}

// GetAll 獲取所有產品
func (r *PostgresProductRepository) GetAll() ([]models.Product, error) {
	var products []models.Product

	err := r.db.Select(&products, `
		SELECT *
		FROM products
		ORDER BY id
	`)

	if err != nil {
		return nil, err
	}

	return products, nil
}

func (r *PostgresProductRepository) GetByID(id int64) (models.Product, error) {
	var product models.Product

	err := r.db.Get(&product, `
		SELECT *
		FROM products
		WHERE id = $1
	`, id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Product{}, ErrProductNotFound
		}
		return models.Product{}, err
	}

	return product, nil
}

// Create 創建新產品
func (r *PostgresProductRepository) Create(input models.Product) (models.Product, error) {
	var product models.Product

	err := r.db.QueryRowx(`
		INSERT INTO products (sku_code, sku_name, sku_amount, expiration)
		VALUES ($1, $2, $3, $4)
		RETURNING id, sku_code, sku_name, sku_amount, expiration
	`, input.SkuCode, input.SkuName, input.SkuAmount, input.Expiration).StructScan(&product)

	if err != nil {
		return models.Product{}, err
	}

	return product, nil
}

// Update 更新產品
func (r *PostgresProductRepository) UpdateNonBlank(id int64, input models.Product) (models.Product, error) {
	// 準備 SQL 查詢部分
	sets := []string{}
	args := []interface{}{}
	argIndex := 1

	// 檢查每個欄位，只添加非空欄位
	if input.SkuCode != "" {
		sets = append(sets, fmt.Sprintf("sku_code = $%d", argIndex))
		args = append(args, input.SkuCode)
		argIndex++
	}

	if input.SkuName != "" {
		sets = append(sets, fmt.Sprintf("sku_name = $%d", argIndex))
		args = append(args, input.SkuName)
		argIndex++
	}

	if input.Expiration != "" {
		sets = append(sets, fmt.Sprintf("expiration = $%d", argIndex))
		args = append(args, input.Expiration)
		argIndex++
	}

	sets = append(sets, fmt.Sprintf("sku_amount = $%d", argIndex))
	args = append(args, input.SkuAmount)
	argIndex++

	sets = append(sets, fmt.Sprintf("update_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	// 構建完整的 SQL 查詢
	query := fmt.Sprintf(`
        UPDATE products
        SET %s
        WHERE id = $%d
        RETURNING id, update_at, sku_code, sku_name, sku_amount, expiration
    `, strings.Join(sets, ", "), argIndex)

	// 添加 ID 到參數列表
	args = append(args, id)

	// 執行查詢
	var product models.Product
	err := r.db.QueryRowx(query, args...).StructScan(&product)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Product{}, ErrProductNotFound
		}
		return models.Product{}, err
	}

	return product, nil
}

// Delete 刪除產品
func (r *PostgresProductRepository) Delete(id int64) error {
	result, err := r.db.Exec(`DELETE FROM products WHERE id = $1`, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrProductNotFound
	}

	return nil
}
