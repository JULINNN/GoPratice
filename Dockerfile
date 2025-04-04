# 建置階段
FROM golang:1.24-alpine AS builder

# 設置工作目錄
WORKDIR /app

# 安裝必要的系統依賴
RUN apk add --no-cache git bash wget

# 複製 go.mod 和 go.sum 檔案並下載依賴
COPY go.mod go.sum ./
RUN go mod download

# 複製源代碼
COPY . .

# 編譯應用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# 運行階段
FROM alpine:3.18

# 安裝 CA 證書、時區資料和必要的工具
RUN apk --no-cache add ca-certificates tzdata bash wget curl jq

# 建立應用目錄
RUN mkdir -p /app/logs

# 從建置階段複製編譯好的二進制檔案
COPY --from=builder /app/main /app/main

# 複製配置檔案
COPY --from=builder /app/configs /app/configs

# 複製測試腳本和入口點腳本
COPY test-api.sh /app/test-api.sh
COPY entrypoint.sh /app/entrypoint.sh

# 設置權限
RUN chmod +x /app/main /app/test-api.sh /app/entrypoint.sh

# 建立非特權用戶
RUN adduser -D -g '' appuser

# 設置目錄和檔案的所有權
RUN chown -R appuser:appuser /app

# 設置工作目錄
WORKDIR /app

# 切換到非特權用戶
USER appuser

# 暴露應用端口
EXPOSE 8080

# 設置健康檢查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://localhost:8080/health || exit 1

# 使用入口點腳本
ENTRYPOINT ["/app/entrypoint.sh"]