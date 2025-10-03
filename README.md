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
- **Platform Gateway** 🔄: HTTP API + WebSocket + User 整合 + gRPC 客戶端
- **Order Service** 🔄: 訂單創建/撮合/查詢

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
# 1. 啟動基礎設施
cd infrastructure
docker-compose up -d

# 2. 開發 Price Service
cd ../price
go mod download
# 按 F5 啟動調試

# 3. 測試 gRPC
grpcurl -plaintext -d '{"symbol":"GOLD"}' localhost:50051 price.PriceService/GetCurrentPrice
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
  - [x] 價格模擬器
  - [x] Redis Pub/Sub
  - [x] InfluxDB 整合
  - [x] gRPC 服務
  - [x] Docker 容器化
- [ ] **Platform Gateway** (🔄 規劃中)
  - [ ] gRPC 客戶端連接 Price Service
  - [ ] HTTP API 服務器
  - [ ] WebSocket 價格推送
  - [ ] Redis 訂閱整合
  - [ ] 用戶管理整合
- [ ] Order Service
- [ ] 前端應用

---

**當前任務**: 開發 Platform Gateway (整合 User + API Gateway)
