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
- **Platform Gateway** ✅: gRPC 客戶端 + Redis 訂閱 + HTTP API + WebSocket（已完成）
- **Web Frontend** ✅: Vue 3 + TypeScript + TradingView Charts（已完成）
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
├── docker-compose.yml            # 🎯 統一的服務編排（一鍵啟動）
├── infrastructure/
│   └── docker-compose.yml        # 基礎設施配置（已整合到根目錄）
├── price/                        # ✅ 價格服務（已完成）
│   ├── Dockerfile                # Docker 構建配置
│   ├── go.mod                    # Go module 定義
│   ├── main.go                   # 服務入口
│   ├── proto/                    # gRPC 定義和生成檔案
│   │   ├── price.proto
│   │   ├── price.pb.go
│   │   └── price_grpc.pb.go
│   └── internal/                 # 內部包
│       ├── config/               # 配置管理
│       ├── model/                # 資料模型
│       ├── simulator/            # 價格模擬器（幾何布朗運動）
│       ├── pubsub/               # Redis 發布
│       ├── repository/           # InfluxDB 存儲
│       ├── service/              # 業務邏輯
│       └── grpc/                 # gRPC 服務器
├── platform/                     # ✅ Platform Gateway（已完成）
│   ├── Dockerfile                # Docker 構建配置
│   ├── go.mod                    # Go module 定義
│   ├── main.go                   # 服務入口
│   ├── proto/                    # gRPC 客戶端定義
│   └── internal/                 # 內部包
│       ├── config/               # 配置管理
│       ├── grpc/                 # gRPC 客戶端
│       ├── http/                 # HTTP API 服務器（Gin）
│       ├── websocket/            # WebSocket 服務器
│       ├── redis/                # Redis 訂閱
│       ├── user/                 # 用戶管理（Demo）
│       └── service/              # 業務邏輯
├── web/                          # ✅ 前端（已完成）
│   ├── Dockerfile                # Docker 構建配置
│   ├── package.json              # npm 依賴定義
│   ├── vite.config.ts            # Vite 配置
│   ├── tailwind.config.js        # Tailwind CSS 配置
│   └── src/                      # 源代碼
│       ├── api/                  # API 服務封裝
│       ├── components/           # Vue 組件
│       ├── stores/               # Pinia 狀態管理
│       ├── router/               # Vue Router 路由
│       ├── views/                # 頁面視圖
│       ├── utils/                # 工具函數
│       └── types/                # TypeScript 類型定義
├── order/                        # 📋 訂單服務（待開發）
├── QUICKSTART.md                 # 快速啟動指南
└── CHANGELOG.md                  # 開發日誌
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

### 一鍵啟動所有服務

```bash
# 在專案根目錄執行
docker-compose up -d
```

這個命令會自動啟動：
- ✅ Redis (Port 6379)
- ✅ InfluxDB (Port 8086) 
- ✅ PostgreSQL (Port 5432)
- ✅ Grafana (Port 3000)
- ✅ Price Service (Port 50051)
- ✅ Platform Service (Port 8080)
- ✅ Web Frontend (Port 5173)

### 常用命令

```bash
# 查看服務狀態
docker-compose ps

# 查看日誌
docker-compose logs -f

# 查看特定服務日誌
docker-compose logs -f platform-service
docker-compose logs -f price-service

# 停止所有服務
docker-compose down

# 停止並清理數據
docker-compose down -v

# 重啟服務
docker-compose restart
```

### 訪問服務

```bash
# 前端應用
http://localhost:5173

# Platform API
curl http://localhost:8080/health
curl http://localhost:8080/api/prices/current
curl http://localhost:8080/api/prices/current?symbol=GOLD
curl "http://localhost:8080/api/prices/history?symbol=GOLD&interval=1m&limit=10"
curl http://localhost:8080/api/user/info

# 支援的 K 線時間間隔
# 1m (1分鐘), 5m (5分鐘), 15m (15分鐘), 30m (30分鐘)
# 1h (1小時), 4h (4小時), 1d (1天)
```

詳細說明請參考 [QUICKSTART.md](./QUICKSTART.md)

### 測試 gRPC（可選）

```bash
# 需要安裝 grpcurl
# macOS: brew install grpcurl

# 測試 Price Service
grpcurl -plaintext -d '{"symbol":"GOLD"}' localhost:50051 price.PriceService/GetCurrentPrice

# 測試 K 線查詢
grpcurl -plaintext -d '{"symbol":"GOLD","interval":"1m","limit":10}' localhost:50051 price.PriceService/GetKlines
```

## 技術亮點

### 1. 統一服務編排
- 單一 `docker-compose.yml` 管理所有服務
- 自動依賴管理和健康檢查
- 一鍵啟動/停止所有服務

### 2. 微服務架構
- **價格服務**: 獨立的價格模擬和資料存儲
- **平台閘道**: 統一的 API 入口和 WebSocket 推送
- **服務間通訊**: gRPC（高效能）+ Redis Pub/Sub（解耦）

### 3. 即時價格推送
- 每秒 3 次價格生成（Price Service）
- 每秒 1 次精選價格推送（Platform Service）
- 支援 best/worst 價格策略

### 4. 現代化技術棧
- **後端**: Golang + Gin + gRPC
- **資料庫**: InfluxDB（時序）+ PostgreSQL（關聯）+ Redis（快取）
- **容器化**: Docker + Docker Compose
- **監控**: Grafana + InfluxDB

### 5. 良好的開發體驗
- 清晰的專案結構
- 完整的健康檢查
- 詳細的日誌輸出
- 易於擴展的架構

## 系統整合完成 ✅

所有後端服務已完成並通過測試：

### Price Service
- ✅ 價格模擬器（幾何布朗運動）
- ✅ InfluxDB 時序資料存儲
- ✅ Redis Pub/Sub 即時推送
- ✅ gRPC 服務完整實現
- ✅ Docker 容器化
- ✅ 健康檢查通過

### Platform Gateway  
- ✅ gRPC 客戶端（連接 Price Service）
- ✅ Redis 訂閱器（接收價格更新）
- ✅ HTTP API 服務器（Gin 框架）
  - GET `/health` - 健康檢查
  - GET `/api/prices/current` - 獲取當前價格
  - GET `/api/prices/history` - 獲取 K 線資料
  - GET `/api/user/info` - 用戶資訊（Demo）
- ✅ WebSocket 服務器
  - WS `/ws/prices` - 即時價格推送
  - 訂閱/取消訂閱機制
  - 心跳檢測（Ping/Pong）
- ✅ 用戶管理系統（Demo 版本）
- ✅ CORS 支援
- ✅ Docker 容器化
- ✅ 健康檢查通過

---

## Platform Gateway 設計文檔（參考）

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
- [x] **Platform Gateway - Phase 2** (✅ 已完成)
  - [x] HTTP API 服務器（使用 Gin 框架）
    - [x] GET /health - 健康檢查
    - [x] GET /api/prices/current - 獲取當前價格
    - [x] GET /api/prices/history - 獲取 K 線資料
    - [x] GET /api/user/info - 用戶資訊（Demo）
  - [x] WebSocket 服務器
    - [x] WS /ws/prices - 即時價格推送
    - [x] 訂閱/取消訂閱機制
    - [x] 心跳檢測（Ping/Pong）
  - [x] 用戶管理整合（簡化版）
  - [x] CORS 支援
  - [x] 測試頁面（test_websocket.html）
- [x] **前端應用 - Phase 1** (✅ 已完成)
  - [x] 專案初始化（Vue 3 + TypeScript + Vite + pnpm）
  - [x] 技術棧整合
    - [x] Tailwind CSS 3.4（基礎樣式）
    - [x] Element Plus 2.11（UI 組件）
    - [x] Pinia 3.0（狀態管理）
    - [x] Vue Router 4（路由）
    - [x] Axios（HTTP 客戶端）
  - [x] WebSocket 服務封裝（自動重連、心跳檢測）
  - [x] HTTP API 服務封裝
  - [x] 價格卡片組件（即時更新、動畫效果、秒數計數）
  - [x] 基礎佈局（導航欄、頁面結構）
  - [x] 工具函數（格式化、常量）
  - [x] TypeScript 類型定義
- [x] **前端應用 - Phase 2** (✅ 已完成)
  - [x] TradingView Lightweight Charts 整合（v4.1.3）
  - [x] K 線圖表組件（支援 7 種時間間隔）
  - [x] 數據驗證和修復機制
    - [x] 去重處理（按時間戳）
    - [x] OHLC 數據修復（處理零值和邏輯錯誤）
    - [x] 時間序列驗證（過濾異常時間戳）
  - [x] 圖表自動刷新（根據時間間隔智能調整）
  - [x] 錯誤處理和優雅降級（無數據時顯示空圖表）
  - [x] Docker 容器化（Node.js 22 Alpine）
- [ ] **前端應用 - Phase 3** (📋 待開發)
  - [ ] 訂單功能（需要 Order Service）
  - [ ] 移動端響應式優化
  - [ ] 深色模式
- [ ] **Order Service** (📋 待開發)
  - [ ] 訂單創建 API
  - [ ] 訂單撮合邏輯
  - [ ] 訂單查詢 API
  - [ ] PostgreSQL 整合

---

**當前狀態**: 全棧系統完整運行，功能齊全 🎉
- ✅ 統一的 docker-compose.yml（一鍵啟動所有服務）
- ✅ Price Service 完整實現並運行正常
  - ✅ 支援 7 種 K 線時間間隔（1m, 5m, 15m, 30m, 1h, 4h, 1d）
  - ✅ InfluxDB Flux 查詢優化（OHLC 聚合）
- ✅ Platform Gateway 完整實現（HTTP API + WebSocket）
  - ✅ gRPC 通訊正常（Platform ↔ Price Service）
  - ✅ Redis Pub/Sub 運作正常（價格推送）
  - ✅ 價格策略（best/worst）運作正常
  - ✅ HTTP API 服務器（Gin 框架，端口 8080）
  - ✅ WebSocket 即時推送（每秒更新）
  - ✅ API 錯誤處理和數據填充機制
- ✅ Web Frontend 完整實現
  - ✅ 現代化技術棧（Vue 3 + TypeScript + Vite）
  - ✅ UI 框架（Tailwind CSS + Element Plus）
  - ✅ 即時價格卡片（WebSocket 推送、秒數自動計數）
  - ✅ TradingView K 線圖表（7 種時間間隔）
  - ✅ 專業數據驗證（去重、修復、時間序列驗證）
  - ✅ 自動刷新機制（根據時間間隔智能調整）
  - ✅ Docker 容器化部署（端口 3000）
- ✅ 所有服務健康檢查通過
- ✅ CORS 支援完整
- 📋 下一步：開發 Order Service（訂單創建、撮合、查詢）

## 服務端點

### HTTP API

```bash
# 健康檢查
curl http://localhost:8080/health

# 獲取所有商品當前價格
curl http://localhost:8080/api/prices/current

# 獲取特定商品價格
curl http://localhost:8080/api/prices/current?symbol=GOLD

# 獲取 K 線資料
curl "http://localhost:8080/api/prices/history?symbol=GOLD&interval=1m&limit=10"

# 獲取用戶資訊
curl http://localhost:8080/api/user/info
```

### WebSocket

```javascript
const ws = new WebSocket('ws://localhost:8080/ws/prices');

ws.onopen = () => {
  ws.send(JSON.stringify({
    type: 'subscribe',
    symbols: ['GOLD', 'SILVER', 'PLATINUM', 'PALLADIUM']
  }));
};

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  console.log(message);
};
```

### 監控面板

- **Grafana**: http://localhost:3000 (admin / admin123)
- **InfluxDB**: http://localhost:8086 (admin / admin123456)

## 技術棧詳細

### 後端技術
- **Language**: Go 1.24
- **Web Framework**: Gin v1.10
- **RPC**: gRPC v1.76
- **Databases**: 
  - InfluxDB 2.7（時序數據）
  - PostgreSQL 16（關聯數據）
  - Redis 7.2（快取/Pub-Sub）
- **WebSocket**: Gorilla WebSocket v1.5

### 前端技術
- **Framework**: Vue 3.5
- **Language**: TypeScript 5.9
- **Build Tool**: Vite 7.1
- **Package Manager**: pnpm 10.18
- **UI Library**: Element Plus 2.11
- **CSS Framework**: Tailwind CSS 3.4
- **State Management**: Pinia 3.0
- **Router**: Vue Router 4.5
- **HTTP Client**: Axios 1.12
- **Charts**: TradingView Lightweight Charts 4.1
- **Date Utils**: dayjs 1.11

### DevOps
- **Containerization**: Docker + Docker Compose
- **Monitoring**: Grafana 10.2
- **Node.js**: v22 (Alpine)
- **Go Runtime**: Alpine-based

## 專案亮點

### 1. 完整的時序數據處理
- **數據生成**: 每秒 3 次價格模擬（幾何布朗運動）
- **數據聚合**: InfluxDB Flux 查詢支援 7 種時間間隔
- **數據推送**: Redis Pub/Sub + WebSocket 即時推送
- **數據驗證**: 前端多層數據驗證和修復機制

### 2. 專業的金融圖表
- **TradingView 整合**: 業界標準的圖表庫
- **多時間間隔**: 1分鐘到1天，共 7 種選擇
- **智能刷新**: 根據時間間隔自動調整刷新頻率
- **數據修復**: 自動處理缺失、重複、異常數據

### 3. 微服務架構
- **服務解耦**: Price Service / Platform Gateway / Web Frontend
- **gRPC 通訊**: 高效能的服務間通訊
- **Redis Pub/Sub**: 異步消息推送
- **統一編排**: Docker Compose 一鍵部署

### 4. 現代化開發體驗
- **類型安全**: TypeScript 全棧類型定義
- **狀態管理**: Pinia 響應式狀態
- **實時更新**: WebSocket 自動重連和心跳檢測
- **錯誤處理**: 優雅降級和用戶友好的錯誤提示

## 下一步計劃

- [ ] Order Service 開發
  - [ ] 訂單創建 API
  - [ ] 訂單撮合引擎
  - [ ] PostgreSQL 持久化
  - [ ] 訂單歷史查詢
- [ ] 前端優化
  - [ ] 移動端響應式設計
  - [ ] 深色模式支援
  - [ ] 性能優化（懶加載、代碼分割）
- [ ] 監控增強
  - [ ] Prometheus 整合
  - [ ] 自定義 Grafana 儀表板
  - [ ] 告警機制

---

**開發時間**: 2025年10月7日  
**狀態**: ✅ 核心功能完成，可投入演示和進一步開發
