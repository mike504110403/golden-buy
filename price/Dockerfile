# 使用官方 Go 映像作為構建階段
FROM golang:1.23-alpine AS builder

# 設置工作目錄
WORKDIR /app

# 安裝必要的工具
RUN apk add --no-cache git protobuf protobuf-dev

# 安裝 Go 依賴
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# 複製 go.mod 和 go.sum
COPY go.mod go.sum ./

# 下載依賴
RUN go mod download

# 複製源碼
COPY . .

# 生成 protobuf 程式碼
RUN protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/price.proto

# 構建應用程式
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o price .

# 使用輕量級的 Alpine 映像作為運行階段
FROM alpine:3.18

# 安裝 ca-certificates 用於 HTTPS 請求
RUN apk --no-cache add ca-certificates

# 創建非 root 用戶
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 設置工作目錄
WORKDIR /app

# 從構建階段複製二進制文件
COPY --from=builder /app/price .

# 更改文件所有權
RUN chown appuser:appgroup /app/price

# 切換到非 root 用戶
USER appuser

# 暴露端口
EXPOSE 50051

# 設置環境變量
ENV GRPC_PORT=50051
ENV INFLUXDB_URL=http://influxdb:8086
ENV INFLUXDB_TOKEN=my-super-secret-auth-token
ENV INFLUXDB_ORG=golden-buy
ENV INFLUXDB_BUCKET=golden_buy
ENV REDIS_ADDR=redis:6379
ENV LOG_LEVEL=info

# 健康檢查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD nc -z localhost 50051 || exit 1

# 啟動應用程式
CMD ["./price"]
