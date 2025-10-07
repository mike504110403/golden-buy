# Golden Buy - 貴金屬交易平台

> 微服務架構學習專案：gRPC、時序資料庫、即時價格推送

## 技術棧

**後端**: Golang + gRPC  
**前端**: Vue 3 + TypeScript  
**資料庫**: PostgreSQL (關聯) + InfluxDB (時序) + Redis (快取/Pub-Sub)  
**圖表**: TradingView Lightweight Charts

## 系統架構

```
前端 (Vue3 + TS)
        ↓ HTTP/WebSocket
    API Gateway
        ↓ gRPC
┌───────┼───────┼───────┐
Price   Order   User    ...
Service Service Service
```

### 微服務職責

- **Price Service** ✅: 價格模擬 → InfluxDB 存儲 → Redis Pub/Sub → gRPC API
- **Platform Gateway** 🔄: gRPC 客戶端 + Redis 訂閱（已完成）→ HTTP API + WebSocket（待開發）
- **Order Service** 📋: 訂單創建/撮合/查詢（未開始）

## 核心決策

✅ 價格來源：模擬器（幾何布朗運動，每秒更新）  
✅ 商品：GOLD / SILVER / PLATINUM / PALLADIUM  
✅ 訂單撮合：Demo 階段全局統一（未來可按用戶配置最佳/最差價）  
✅ K 線圖：TradingView Lightweight Charts  
✅ 開發策略：**由大到小，單服務獨立開發**

## 專案結構

```
golden-buy/
├── infrastructure/
│   └── docker-compose.yml        # Redis + InfluxDB + PostgreSQL + Grafana
├── price/                        # 價格服務 (✅ 已完成)
│   ├── go.mod                    # Go module 定義
│   ├── main.go                   # 服務入口
│   ├── proto/                    # gRPC 定義和生成檔案
│   │   ├── price.proto
│   │   ├── price.pb.go
│   │   └── price_grpc.pb.go
│   └── internal/                 # 內部包
│       ├── config/               # 配置管理
│       ├── model/                # 資料模型
│       ├── simulator/            # 價格模擬器
│       ├── pubsub/               # Redis 發布
│       ├── repository/           # InfluxDB 存儲
│       ├── service/              # 業務邏輯
│       └── grpc/                 # gRPC 服務器
├── platform/                     # Platform Gateway (🔄 規劃中)
│   ├── go.mod                    # Go module 定義
│   ├── main.go                   # 服務入口
│   ├── proto/                    # gRPC 客戶端定義
│   └── internal/                 # 內部包
│       ├── config/               # 配置管理
│       ├── grpc/                 # gRPC 客戶端
│       ├── http/                 # HTTP API 服務器
│       ├── websocket/            # WebSocket 服務器
│       ├── redis/                # Redis 訂閱
│       └── service/              # 業務邏輯
├── order/                        # 訂單服務 (未來開發)
├── web/                          # 前端 (未來開發)
├── .vscode/
│   └── launch.json               # 調試配置
└── README.md
```

## Price Service 架構

### 資料流
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

### gRPC 接口
- `GetCurrentPrice` - 獲取當前價格
- `GetCurrentPrices` - 批量獲取當前價格
- `SubscribePrices` - 訂閱價格流 (Server Streaming)
- `GetKlines` - 獲取歷史 K 線資料

### 支援商品
- GOLD (黃金) - 初始價格: $1,850
- SILVER (白銀) - 初始價格: $24  
- PLATINUM (鉑金) - 初始價格: $950
- PALLADIUM (鈀金) - 初始價格: $1,280

### 技術規格
- 價格更新頻率: 每秒 3 次 (間隔 333ms)
- 波動率: 0.5% - 1%
- 快取 TTL: 5 分鐘
- 每秒價格記錄: Redis List，保留 3 筆，10 秒後過期
- gRPC 端口: 50051

## 快速開始

```bash
# 1. 啟動基礎設施（Redis、InfluxDB、PostgreSQL、Grafana）
cd infrastructure
docker-compose up -d

# 2. 啟動 Price Service
cd ../price
go run .
# 或使用 Docker: docker-compose up -d

# 3. 啟動 Platform Service
cd ../platform
go run .

# 4. 測試
# - 觀察終端輸出，應該看到：
#   ✅ K 線查詢成功
#   💰 每秒選擇最佳/最差價格
#   📊 價格更新推送
```

### 測試 gRPC（可選）

```bash
# 測試 Price Service
grpcurl -plaintext -d '{"symbol":"GOLD"}' localhost:50051 price.PriceService/GetCurrentPrice

# 測試 K 線查詢
grpcurl -plaintext -d '{"symbol":"GOLD","interval":"1m","limit":10}' localhost:50051 price.PriceService/GetKlines
```

## Platform Gateway 規劃

### 功能設計

#### 1. 前端通訊方式
**建議使用 HTTP API + WebSocket 組合**：
- **HTTP API**: 歷史資料查詢、用戶操作、一次性請求
- **WebSocket**: 即時價格推送、即時通知、雙向通訊

#### 2. 核心功能
- **gRPC 客戶端**: 連接 Price Service 獲取歷史資料
- **Redis 訂閱**: 接收即時價格更新
- **WebSocket 服務器**: 推送即時價格到前端
- **HTTP API 服務器**: 提供 RESTful 接口
- **用戶整合**: 簡化用戶管理 (Demo 階段)

#### 3. API 設計
```
GET  /api/prices/current     # 獲取當前價格
GET  /api/prices/history     # 獲取歷史 K 線資料
WS   /ws/prices              # WebSocket 價格推送
GET  /api/user/info          # 用戶資訊 (Demo)  
```

## 開發進度

- [x] 專案架構設計
- [x] **Price Service** (✅ 已完成)
  - [x] 專案結構與配置
  - [x] Proto 定義
  - [x] 價格模擬器（幾何布朗運動）
  - [x] Redis Pub/Sub（每秒推送 3 次）
  - [x] InfluxDB 整合（存儲價格和 K 線）
  - [x] gRPC 服務（查詢、訂閱、K 線）
  - [x] Docker 容器化
  - [x] K 線查詢修復（Flux 語法 OHLC 聚合）
- [x] **Platform Gateway - Phase 1** (✅ 已完成並測試)
  - [x] 專案結構與配置
  - [x] Proto 定義（gRPC 客戶端）
  - [x] gRPC 客戶端連接 Price Service
    - [x] GetCurrentPrice - 單個商品價格
    - [x] GetCurrentPrices - 批量查詢
    - [x] GetKlines - 歷史 K 線資料 ✅ 測試成功
  - [x] Redis 訂閱器整合
    - [x] 訂閱 `price:updates` 頻道
    - [x] 價格緩衝（每秒收集 3 筆）
    - [x] 策略選擇（best/worst）
    - [x] 每秒推送 1 筆處理後的價格
  - [x] 數據模型（Price、Kline、PriceBuffer）
  - [x] 主服務邏輯整合
  - [x] Docker 容器化
  - [x] 測試驗證
- [ ] **Platform Gateway - Phase 2** (📋 待開發)
  - [ ] HTTP API 服務器
    - [ ] GET /api/prices/current
    - [ ] GET /api/prices/history
    - [ ] GET /api/user/info (Demo)
  - [ ] WebSocket 服務器
    - [ ] WS /ws/prices - 即時推送
  - [ ] 用戶管理整合（簡化版）
- [ ] Order Service
- [ ] 前端應用

---

**當前狀態**: Platform Gateway Phase 1 完成並測試通過
- ✅ gRPC 連接正常
- ✅ Redis 訂閱運作正常
- ✅ K 線查詢返回正確的 OHLC 數據
- ✅ 價格策略（best/worst）運作正常
- 📋 下一步：開發 HTTP API + WebSocket 推送
