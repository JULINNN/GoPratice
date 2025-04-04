-- 創建產品表
CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    sku_code VARCHAR(50) NOT NULL,
    sku_name VARCHAR(100) NOT NULL,
    sku_amount INT NOT NULL DEFAULT 0,
    expiration VARCHAR(50),
    create_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    update_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 添加索引
CREATE INDEX IF NOT EXISTS idx_products_sku_code ON products(sku_code);

-- 添加一些測試數據
INSERT INTO products (sku_code, sku_name, sku_amount, expiration)
VALUES 
    ('SKU001', '測試產品1', 100, '2025-12-31'),
    ('SKU002', '測試產品2', 50, '2025-06-30'),
    ('SKU003', '測試產品3', 200, '2026-01-15');