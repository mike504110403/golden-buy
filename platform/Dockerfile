# 多階段構建
FROM golang:1.23-alpine AS builder

# 安裝必要工具
RUN apk add --no-cache git

# 設置工作目錄
WORKDIR /app

# 複製 go mod 文件
COPY go.mod go.sum ./

# 下載依賴
RUN go mod download

# 複製源代碼
COPY . .

# 構建應用程式
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o platform .

# 運行階段
FROM alpine:latest

# 安裝 ca-certificates (用於 HTTPS)
RUN apk --no-cache add ca-certificates tzdata

# 設置時區
ENV TZ=Asia/Taipei

# 創建非 root 用戶
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

# 從構建階段複製二進制文件
COPY --from=builder /app/platform .

# 切換到非 root 用戶
USER appuser

# 暴露端口（未來 HTTP API 使用）
EXPOSE 8080

# 健康檢查（未來實現 HTTP 健康檢查端點後啟用）
# HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
#   CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 運行應用程式
CMD ["./platform"]

