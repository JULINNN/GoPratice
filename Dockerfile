# 建置階段
FROM golang:1.24-alpine AS builder

# 設置工作目錄
WORKDIR /app

# 安裝必要的系統依賴
RUN apk add --no-cache git

# 複製 go.mod 和 go.sum 檔案並下載依賴
COPY go.mod go.sum ./
RUN go mod download

# 複製源代碼
COPY . .

# 編譯應用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# 運行階段
FROM alpine:3.18

# 安裝 CA 證書和時區資料
RUN apk --no-cache add ca-certificates tzdata

# 建立非特權用戶
RUN adduser -D -g '' appuser

# 設置工作目錄
WORKDIR /app

# 從建置階段複製編譯好的二進制檔案
COPY --from=builder /app/main .

# 複製配置檔案
COPY --from=builder /app/configs /app/configs

# 建立日誌目錄並設置權限
RUN mkdir -p /app/logs && chown -R appuser:appuser /app

# 切換到非特權用戶
USER appuser

# 暴露應用端口
EXPOSE 8080

# 設置健康檢查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://localhost:8080/health || exit 1

# 啟動應用
CMD ["./main"]