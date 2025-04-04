#!/bin/bash

# test-api.sh
echo "開始 API 測試..."

# 健康檢查
echo "1. 測試健康檢查端點"
curl -s http://localhost:8080/health | jq .

# 創建產品
echo -e "\n2. 創建新產品"
CREATE_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/products \
  -H "Content-Type: application/json" \
  -d '{
    "sku_code": "TEST001",
    "sku_name": "測試產品",
    "sku_amount": 100,
    "expiration": "2025-12-31"
  }')
echo $CREATE_RESPONSE | jq .

# 解析產品 ID
PRODUCT_ID=$(echo $CREATE_RESPONSE | jq -r '.id')
echo "創建的產品 ID: $PRODUCT_ID"

# 獲取產品列表
echo -e "\n3. 獲取產品列表"
curl -s http://localhost:8080/api/v1/products | jq .

# 獲取單個產品
echo -e "\n4. 獲取單個產品"
curl -s http://localhost:8080/api/v1/products/$PRODUCT_ID | jq .

# 更新產品
echo -e "\n5. 更新產品"
curl -s -X PUT http://localhost:8080/api/v1/products/$PRODUCT_ID \
  -H "Content-Type: application/json" \
  -d '{
    "sku_code": "TEST001",
    "sku_name": "已更新測試產品",
    "sku_amount": 200,
    "expiration": "2026-12-31"
  }' | jq .

# 再次獲取產品查看更新結果
echo -e "\n6. 查看更新後的產品"
curl -s http://localhost:8080/api/v1/products/$PRODUCT_ID | jq .

# 刪除產品
echo -e "\n7. 刪除產品"
curl -s -X DELETE http://localhost:8080/api/v1/products/$PRODUCT_ID | jq .

# 確認產品已刪除
echo -e "\n8. 確認產品已刪除"
curl -s http://localhost:8080/api/v1/products/$PRODUCT_ID

echo -e "\n測試完成!"