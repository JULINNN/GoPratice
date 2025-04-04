# GoPratice

Go REST API 產品管理系統，使用Gin框架與PostgreSQL資料庫。

## 專案架構

- PostgreSQL資料庫
- Docker容器化
- Zap日誌
- 配置管理
- 單元與整合測試

## 依賴

- Gin框架
- PostgreSQL
- sqlx
- Zap + lumberjack
- testify、sqlmock、dockertest
- Docker & Docker Compose

## 目錄結構

```
.
├── cmd/server/           # 應用入口
├── configs/              # 配置文件
├── internal/             # 核心代碼
│   ├── config/           # 配置管理
│   ├── controller/       # API控制器
│   ├── logger/           # 日誌功能
│   ├── models/           # 資料模型
│   ├── repository/       # 資料存取
│   └── service/          # 業務邏輯
├── pkg/database/         # 資料庫工具
├── tests/                # 測試文件
├── Dockerfile            # Docker構建
├── docker-compose.yml    # Docker組合
└── go.mod                # Go模組
```

## API端點

| 方法    | 端點                | 描述           | 狀態碼 |
|--------|---------------------|---------------|--------|
| GET    | /health             | 健康檢查       | 200 OK |
| GET    | /api/v1/products    | 獲取所有產品    | 200 OK |
| GET    | /api/v1/products/:id | 獲取單個產品   | 200 OK / 404 Not Found |
| POST   | /api/v1/products    | 創建產品       | 201 Created / 400 Bad Request |
| PUT    | /api/v1/products/:id | 更新產品       | 200 OK / 404 Not Found |
| DELETE | /api/v1/products/:id | 刪除產品       | 200 OK / 404 Not Found |

## 產品PoJo

```json
{
  "id": 1,
  "sku_code": "SKU001",
  "sku_name": "產品名稱",
  "sku_amount": 100,
  "expiration": "2025-12-31",
  "create_at": "2024-04-04 12:34:56",
  "update_at": "2024-04-04 12:34:56"
}
```

## 錯誤回應格式

```json
{
  "error_code": "PRODUCT_NOT_FOUND",
  "error_message": "產品未找到",
  "request_id": "a4b3c2-d1e0-f1g2-h3i4"
}
```

## 使用方法

### 本地運行

```bash
# 安裝依賴
go mod download

# 運行應用
go run cmd/server/main.go
```

### Docker運行

```bash
# 啟動容器
docker-compose up -d
```

- 應用: http://localhost:8080

## 測試

```bash
# 所有測試
go test ./tests/...

# 整合測試
go test ./tests/... -tags=integration

# API測試
./test-api.sh
```

## 環境變數

| 變數        | 描述         | 默認值            |
|------------|--------------|------------------|
| SERVER_PORT | 服務器端口    | 8080             |
| GIN_MODE    | Gin模式      | debug            |
| DB_HOST     | 資料庫主機    | localhost        |
| DB_PORT     | 資料庫端口    | 5432             |
| DB_USER     | 資料庫用戶    | postgres         |
| DB_PASSWORD | 資料庫密碼    | postgres         |
| DB_NAME     | 資料庫名稱    | product_db       |
| LOG_LEVEL   | 日誌級別      | info             |