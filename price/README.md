# Price Service

貴金屬價格服務 - 負責價格模擬、存儲和推送

## 功能特性

- **價格模擬**: 使用幾何布朗運動生成 4 種貴金屬的即時價格
- **即時推送**: 通過 Redis Pub/Sub 廣播價格更新
- **資料存儲**: 寫入 InfluxDB 時序資料庫
- **gRPC 服務**: 提供價格查詢和訂閱接口
- **每秒記錄**: Redis 記錄每種金屬每秒的 3 筆價格

## 支援商品

- GOLD (黃金) - 初始價格: $1,850
- SILVER (白銀) - 初始價格: $24  
- PLATINUM (鉑金) - 初始價格: $950
- PALLADIUM (鈀金) - 初始價格: $1,280

## 技術規格

- 價格更新頻率: 每秒 3 次 (間隔 333ms)
- 波動率: 0.5% - 1%
- gRPC 端口: 50051
- Redis 快取 TTL: 60 秒
- 每秒價格記錄 TTL: 10 分鐘

## 快速開始

### 1. 啟動周邊服務

```bash
# 啟動 InfluxDB 和 Redis
cd ../infrastructure
docker-compose up -d
```

### 2. 本地開發

```bash
# 生成 protobuf 程式碼
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/price.proto

# 編譯並執行
go build -o price . && ./price

# 或直接執行
go run .
```

### 3. Docker 部署

```bash
# 構建 Docker 映像
docker-compose build

# 啟動 Price Service
docker-compose up -d

# 查看服務狀態
docker-compose ps

# 查看日誌
docker-compose logs -f price-service
```

## gRPC 接口

- `GetCurrentPrice` - 獲取當前價格
- `GetCurrentPrices` - 批量獲取當前價格
- `SubscribePrices` - 訂閱價格流 (Server Streaming)
- `GetKlines` - 獲取歷史 K 線資料

## 資料流

```
1. 價格生成流 (每秒 3 次)：
   Simulator → Service → InfluxDB (存儲)
                    → Redis Pub/Sub (廣播最新價格)
                    → Redis 即時價格 (覆蓋更新)
                    → Redis List (每秒價格記錄)

2. 查詢流程：
   Client → gRPC → Service → Simulator/Cache/InfluxDB

3. 即時推送：
   Client → gRPC Streaming → Service → Simulator

4. Redis 存儲結構：
   即時價格: price:{SYMBOL} (4 筆固定 key)
   每秒記錄: price:second:{SYMBOL}:{UNIX_MILLIS} (List of 3 prices)
   TTL: 10 minutes
```

## 環境變數

| 變數名 | 預設值 | 說明 |
|--------|--------|------|
| `GRPC_PORT` | 50051 | gRPC 服務端口 |
| `INFLUXDB_URL` | http://localhost:8086 | InfluxDB 連線位址 |
| `INFLUXDB_TOKEN` | my-super-secret-auth-token | InfluxDB 認證令牌 |
| `INFLUXDB_ORG` | golden-buy | InfluxDB 組織名稱 |
| `INFLUXDB_BUCKET` | golden_buy | InfluxDB 儲存桶名稱 |
| `REDIS_ADDR` | localhost:6379 | Redis 連線位址 |
| `LOG_LEVEL` | info | 日誌級別 |

## 測試

### gRPC 接口測試

```bash
# 測試獲取當前價格
grpcurl -plaintext -d '{"symbol":"GOLD"}' localhost:50051 price.PriceService/GetCurrentPrice

# 測試訂閱價格流
grpcurl -plaintext -d '{"symbols":["GOLD","SILVER"]}' localhost:50051 price.PriceService/SubscribePrices

# 測試獲取 K 線資料
grpcurl -plaintext -d '{"symbol":"GOLD","interval":"1m","limit":10}' localhost:50051 price.PriceService/GetKlines
```

### Redis 資料驗證

```bash
# 查看即時價格
docker exec golden-buy-redis redis-cli GET price:GOLD

# 查看每秒價格記錄
docker exec golden-buy-redis redis-cli LRANGE "price:second:GOLD:$(date -u +%s)000" 0 -1

# 訂閱價格更新
docker exec golden-buy-redis redis-cli SUBSCRIBE price:updates
```

## 開發指令

```bash
# 生成 protobuf 程式碼
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/price.proto

# 編譯應用程式
go build -o price .

# 執行應用程式
./price

# 執行測試
go test -v ./...

# 清理編譯檔案
rm -f price

# Docker 相關指令
docker-compose build    # 構建映像
docker-compose up -d    # 啟動服務
docker-compose down     # 停止服務
docker-compose down -v --rmi all  # 清理資源
```

## 專案結構

```
price/
├── Dockerfile              # Docker 映像定義
├── docker-compose.yml      # Docker Compose 配置
├── go.mod                 # Go 模組定義
├── main.go                # 服務入口
├── proto/                 # gRPC 定義
│   ├── price.proto
│   ├── price.pb.go
│   └── price_grpc.pb.go
└── internal/              # 內部包
    ├── config/            # 配置管理
    ├── model/             # 資料模型
    ├── simulator/         # 價格模擬器
    ├── pubsub/            # Redis 發布
    ├── repository/        # InfluxDB 存儲
    ├── service/           # 業務邏輯
    └── grpc/              # gRPC 服務器
```

## 監控

### Grafana 查詢範例

```flux
# 基本價格查詢
from(bucket: "golden_buy")
  |> range(start: -1h)
  |> filter(fn: (r) => r["_measurement"] == "prices")
  |> filter(fn: (r) => r["_field"] == "price")
  |> filter(fn: (r) => r["symbol"] == "GOLD")

# K 線聚合查詢 (1 分鐘)
from(bucket: "golden_buy")
  |> range(start: -1h)
  |> filter(fn: (r) => r["_measurement"] == "prices")
  |> filter(fn: (r) => r["_field"] == "price")
  |> filter(fn: (r) => r["symbol"] == "GOLD")
  |> aggregateWindow(every: 1m, fn: aggregate, createEmpty: false)
```

## 故障排除

### 常見問題

1. **服務無法啟動**
   - 確認 InfluxDB 和 Redis 服務已啟動
   - 檢查端口是否被佔用

2. **gRPC 連接失敗**
   - 確認服務運行在正確端口 (50051)
   - 檢查防火牆設定

3. **InfluxDB 連接失敗**
   - 確認 InfluxDB 服務狀態
   - 檢查認證令牌是否正確

4. **Redis 連接失敗**
   - 確認 Redis 服務狀態
   - 檢查連線位址和端口

### 日誌查看

```bash
# Docker 日誌
make logs

# 或直接使用 docker-compose
docker-compose logs -f price-service
```