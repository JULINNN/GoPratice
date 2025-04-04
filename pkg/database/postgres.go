// pkg/database/db.go
package database

import (
	"fmt"
	"log"
	"main/internal/config"
	"time"

	"github.com/jmoiron/sqlx"
)

// NewPostgresDB 創建 PostgreSQL 資料庫連接，包含重試邏輯
func NewPostgresDB(cfg *config.DatabaseConfig) (*sqlx.DB, error) {
	// 在 Docker 環境中，資料庫可能需要時間啟動，添加重試邏輯
	var db *sqlx.DB
	var err error

	maxRetries := 5
	retryInterval := time.Second * 3

	for i := 0; i < maxRetries; i++ {
		db, err = sqlx.Connect("postgres", cfg.DSN())
		if err == nil {
			break
		}

		log.Printf("嘗試連接資料庫失敗 (%d/%d): %v", i+1, maxRetries, err)
		if i < maxRetries-1 {
			log.Printf("等待 %v 後重試...", retryInterval)
			time.Sleep(retryInterval)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("連接資料庫失敗 (重試 %d 次後): %w", maxRetries, err)
	}

	// 驗證連接
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("資料庫連接測試失敗: %w", err)
	}

	// 配置連接池
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Minute * 5)

	log.Println("成功連接到 PostgreSQL 資料庫")

	return db, nil
}
