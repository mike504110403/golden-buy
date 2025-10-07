# Platform Service

平台閘道服務 - 連接前端和後端微服務的核心層

## 功能特性

- **gRPC 客戶端**: 連接 Price Service 獲取歷史 K 線資料
- **Redis 訂閱**: 訂閱 Price Service 的即時價格推送
- **價格策略**: 每秒處理 3 筆價格，選擇最佳或最差價格
- **未來功能**: HTTP API、WebSocket 推送、用戶管理

## 核心功能

### 1. gRPC 客戶端

連接 Price Service，提供以下功能：
- 獲取單個/多個商品當前價格
- 獲取歷史 K 線資料（用於圖表）
- 訂閱價格流（未使用，改用 Redis Pub/Sub）

### 2. Redis 訂閱器

訂閱 `price:updates` 頻道，接收 Price Service 推送的價格更新：
- **每秒 3 筆價格**: Price Service 每秒推送 3 次價格（每 333ms 一次）
- **價格緩衝**: 將同一秒內的 3 筆價格存入緩衝區
- **策略選擇**: 每秒結束時，從緩衝區選擇：
  - `best`: 最低價格（對用戶最有利的買入價）
  - `worst`: 最高價格（對用戶最不利的買入價）
- **推送頻率**: 每秒推送 1 次處理後的價格（未來用於 WebSocket）

### 3. 價格策略說明

```
Price Service 推送（每秒 3 筆）:
  0ms    → $1850.23
  333ms  → $1850.45  ← 最高價（worst）
  666ms  → $1850.12  ← 最低價（best）

Platform Service 處理（每秒 1 筆）:
  根據策略選擇並推送給前端
  - PRICE_STRATEGY=best  → $1850.12
  - PRICE_STRATEGY=worst → $1850.45
```

這個設計對應訂單系統需求：Demo 階段可以配置給用戶「這一秒內最佳價格」或「最差價格」。

## 技術規格

- gRPC 端口: 連接 Price Service (50051)
- HTTP 端口: 8080（未來 API 服務器）
- Redis 訂閱: `price:updates` 頻道
- 價格緩衝: 每秒收集 3 筆，推送 1 筆

## 環境變數

| 變數名 | 預設值 | 說明 |
|--------|--------|------|
| `PRICE_SERVICE_ADDR` | localhost:50051 | Price Service gRPC 地址 |
| `GRPC_TIMEOUT` | 10s | gRPC 請求超時時間 |
| `REDIS_ADDR` | localhost:6379 | Redis 連線位址 |
| `REDIS_PASSWORD` | "" | Redis 密碼 |
| `REDIS_DB` | 0 | Redis 資料庫編號 |
| `PRICE_STRATEGY` | best | 價格策略：best 或 worst |
| `HTTP_PORT` | 8080 | HTTP API 端口（未來使用） |
| `LOG_LEVEL` | info | 日誌級別 |

## 快速開始

### 1. 前置條件

確保以下服務已啟動：
```bash
# 啟動基礎設施（Redis、InfluxDB）
cd ../infrastructure
docker-compose up -d

# 啟動 Price Service
cd ../price
docker-compose up -d
# 或本地運行: go run .
```

### 2. 本地開發

```bash
# 編譯並執行
go build -o platform . && ./platform

# 或直接執行
go run .
```

### 3. Docker 部署

```bash
# 構建 Docker 映像
docker-compose build

# 啟動 Platform Service
docker-compose up -d

# 查看日誌
docker-compose logs -f platform-service
```

## 測試結果 ✅

**Platform Service 測試成功！**

### 已驗證功能

1. ✅ **gRPC 客戶端連接** - 成功連接 Price Service
2. ✅ **Redis 訂閱** - 成功訂閱 `price:updates` 頻道
3. ✅ **K 線查詢** - 成功獲取歷史 OHLC 數據
   ```
   ✅ [GOLD] Retrieved 10 klines
      Latest kline: Open=1838.62, High=2037.09, Low=1767.47, Close=1899.58
   ✅ [SILVER] Retrieved 10 klines
   ✅ [PLATINUM] Retrieved 10 klines
   ✅ [PALLADIUM] Retrieved 10 klines
   ```
4. ✅ **價格緩衝** - 每秒收集 3 筆價格
5. ✅ **價格策略** - 成功選擇 best/worst 價格

### 自動測試內容

Platform Service 啟動後會自動：
1. 連接 Price Service（gRPC）
2. 訂閱 Redis 價格更新
3. 測試獲取 K 線資料（4 種貴金屬，1 分鐘間隔，最近 10 筆）
4. 每 10 秒顯示最新處理後的價格

### 預期輸出

```
🎯 Golden Buy - Platform Service
========================================
Price Service: localhost:50051
Redis: localhost:6379
Price Strategy: best
HTTP Port: 8080
========================================
✅ Connected to Price Service at localhost:50051
✅ Connected to Redis at localhost:6379
🚀 Starting Platform Service...
✅ Subscribed to Redis channel: price:updates
📊 Price strategy: best
✅ Platform Service started

📝 [GOLD] New second buffer: 1234567890
📝 [SILVER] New second buffer: 1234567890
💰 [GOLD] Selected best price: 1850.23 (from 3 prices)
💰 [SILVER] Selected best price: 24.12 (from 3 prices)
📊 Latest price updated: GOLD = 1850.23 (change: 0.05%)

🧪 Testing K-line data retrieval...
📈 Retrieved 10 klines for GOLD (1m interval)
✅ [GOLD] Retrieved 10 klines
   Latest kline: Open=1850.00, High=1851.50, Low=1849.80, Close=1850.23
```

## 專案結構

```
platform/
├── Dockerfile              # Docker 映像定義
├── docker-compose.yml      # Docker Compose 配置
├── go.mod                 # Go 模組定義
├── main.go                # 服務入口
├── proto/                 # gRPC 定義（從 Price Service 複製）
│   ├── price.proto
│   ├── price.pb.go
│   └── price_grpc.pb.go
└── internal/              # 內部包
    ├── config/            # 配置管理
    ├── model/             # 資料模型（含價格緩衝邏輯）
    ├── grpc/              # gRPC 客戶端
    ├── redis/             # Redis 訂閱器（含價格策略）
    └── service/           # 業務邏輯（整合 gRPC 和 Redis）
```

## 開發指令

```bash
# 生成 protobuf 程式碼
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/price.proto

# 編譯應用程式
go build -o platform .

# 執行應用程式
./platform

# 清理編譯檔案
rm -f platform

# Docker 相關指令
docker-compose build    # 構建映像
docker-compose up -d    # 啟動服務
docker-compose down     # 停止服務
docker-compose logs -f  # 查看日誌
```

## 設計決策

### 為什麼用 Redis Pub/Sub 而不是 gRPC Streaming？

1. **解耦**: Price Service 不需要知道有多少個 Platform Service 實例
2. **可擴展**: 可以輕鬆水平擴展 Platform Service
3. **統一推送**: 所有訂閱者都能收到相同的價格更新
4. **容錯**: 如果 Platform Service 重啟，不影響 Price Service

### 為什麼每秒只推送 1 次？

1. **減少前端壓力**: 前端不需要處理每秒 3 次更新
2. **更好的 UX**: 避免價格跳動太快，用戶看不清
3. **符合業務需求**: 訂單系統需要「這一秒內的最佳/最差價」
4. **節省頻寬**: WebSocket 推送頻率降低到原來的 1/3

## 未來開發

- [ ] HTTP API 服務器
  - `GET /api/prices/current` - 獲取當前價格
  - `GET /api/prices/history` - 獲取 K 線資料
  - `GET /api/user/info` - 用戶資訊（Demo）
- [ ] WebSocket 服務器
  - `WS /ws/prices` - 即時價格推送
- [ ] 用戶管理整合（簡化版）
- [ ] 健康檢查端點
- [ ] Prometheus 監控指標
- [ ] 測試覆蓋

## 故障排除

### 常見問題

1. **無法連接 Price Service**
   - 確認 Price Service 已啟動
   - 檢查 `PRICE_SERVICE_ADDR` 是否正確
   - 使用 `grpcurl` 測試 Price Service

2. **無法連接 Redis**
   - 確認 Redis 服務已啟動
   - 檢查 `REDIS_ADDR` 是否正確
   - 使用 `redis-cli` 測試連接

3. **沒有收到價格更新**
   - 確認 Price Service 正在運行並推送價格
   - 檢查 Redis Pub/Sub：`redis-cli SUBSCRIBE price:updates`
   - 查看日誌確認訂閱狀態

### 日誌查看

```bash
# Docker 日誌
docker-compose logs -f platform-service

# 檢查是否訂閱成功
docker-compose logs platform-service | grep "Subscribed"

# 檢查價格處理
docker-compose logs platform-service | grep "Selected"
```

