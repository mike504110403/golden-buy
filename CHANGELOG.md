# 開發日誌

## 2025-10-07 - Platform Gateway Phase 2 完成

### 新增功能

#### HTTP API 服務器（使用 Gin 框架）
- ✅ 建立 HTTP 服務器（端口 8080）
- ✅ API 路由設計
  - GET /health - 健康檢查
  - GET /api/prices/current - 獲取當前價格（支援單個或全部商品）
  - GET /api/prices/history - 獲取歷史 K 線資料
  - GET /api/user/info - 獲取用戶資訊（Demo）
- ✅ 統一回應格式（success, data, error）
- ✅ CORS 中間件支援
- ✅ 日誌中間件（Gin 內建）

#### WebSocket 服務器
- ✅ WebSocket Hub 連接管理器
  - 客戶端註冊/註銷
  - 訂閱管理（symbol -> clients）
  - 廣播機制
- ✅ WebSocket 客戶端處理
  - 讀寫分離（readPump/writePump）
  - 心跳檢測（Ping/Pong）
  - 訊息處理（subscribe/unsubscribe）
- ✅ 價格推送整合
  - 從 Redis 接收價格更新
  - 即時推送到訂閱的 WebSocket 客戶端
  - 每秒 1 次推送頻率

#### 用戶管理系統（Demo 版本）
- ✅ 簡化的用戶管理器
- ✅ Demo 用戶資料
- ✅ 餘額管理功能

#### 測試工具
- ✅ 建立 WebSocket 測試頁面（test_websocket.html）
  - 美觀的 UI 設計
  - 即時價格顯示
  - WebSocket 連接控制
  - HTTP API 測試功能
  - 連接日誌記錄

### 技術細節

#### Gin 框架整合
```go
// 路由分組
api := engine.Group("/api")
{
    prices := api.Group("/prices")
    {
        prices.GET("/current", handler.HandleGetCurrentPrice)
        prices.GET("/history", handler.HandleGetHistory)
    }
}
```

#### WebSocket 訊息格式
```json
// 訂閱
{"type": "subscribe", "symbols": ["GOLD", "SILVER"]}

// 價格更新
{
  "type": "price_update",
  "symbol": "GOLD",
  "data": {
    "symbol": "GOLD",
    "price": 1850.23,
    "change_percent": 0.15,
    "timestamp": 1234567890000
  }
}
```

#### 價格推送流程
```
Redis Subscriber (每秒 1 次處理後的價格)
  → PlatformService.handlePriceUpdate()
    → WebSocket Hub.BroadcastPrice()
      → 推送到所有訂閱該商品的客戶端
```

### 檔案清單

**新增檔案**:
```
platform/internal/
├── http/
│   ├── handler.go (使用 Gin 重寫)
│   └── router.go (使用 Gin 重寫)
├── websocket/
│   ├── hub.go (連接管理器)
│   ├── client.go (客戶端處理)
│   └── handler.go (WebSocket 處理器)
└── user/
    ├── user.go (用戶管理)
    └── errors.go (錯誤定義)

platform/
└── test_websocket.html (WebSocket 測試頁面)
```

**修改檔案**:
```
platform/internal/service/service.go (整合 WebSocket Hub)
platform/main.go (啟動 HTTP 服務器)
platform/go.mod (添加 Gin 和 gorilla/websocket 依賴)
platform/README.md (更新文檔)
README.md (更新開發進度)
```

### 依賴更新
- 添加 `github.com/gin-gonic/gin` v1.11.0
- 添加 `github.com/gorilla/websocket` v1.5.3

### 測試驗證
- ✅ HTTP API 編譯通過
- ✅ WebSocket 服務器編譯通過
- ✅ 服務整合編譯通過
- ⏳ 待運行測試（需要先啟動 infrastructure 和 price service）

### 下一步計劃

**Order Service**:
1. 訂單創建 API
2. 訂單撮合邏輯
3. 訂單查詢 API
4. PostgreSQL 整合
5. 訂單狀態管理

---

## 2025-10-07 - Platform Gateway Phase 1 完成

### 新增功能

#### Platform Service
- ✅ 建立 Platform Service 專案結構
- ✅ 配置管理系統（環境變數支援）
- ✅ gRPC 客戶端實現
  - 連接 Price Service
  - GetCurrentPrice - 單個商品查詢
  - GetCurrentPrices - 批量查詢
  - GetKlines - K 線歷史資料查詢
  - Ping - 健康檢查
- ✅ Redis 訂閱器實現
  - 訂閱 `price:updates` 頻道
  - 價格緩衝機制（每秒收集 3 筆）
  - 價格策略選擇（best/worst）
  - 每秒處理並推送 1 筆精選價格
- ✅ 數據模型
  - Price - 價格資料結構
  - Kline - K 線資料結構
  - PriceBuffer - 價格緩衝區（含最佳/最差價格邏輯）
- ✅ 主服務邏輯整合
- ✅ Docker 容器化配置

#### Price Service 修復
- 🔧 修復 K 線查詢 Flux 語法錯誤
  - 移除無效的 `aggregate` 函數
  - 改用 `first`, `max`, `min`, `last` 分別計算 OHLC
  - 使用 `union` 和 `pivot` 合併結果
  - 新增 `getFloat64Value` 輔助函數安全提取數值

### 測試結果

#### 成功驗證
- ✅ gRPC 客戶端成功連接 Price Service (localhost:50051)
- ✅ Redis 訂閱器成功訂閱價格更新
- ✅ K 線查詢返回正確的 OHLC 數據
  - GOLD: 10 筆 K 線，正確的開高低收
  - SILVER: 10 筆 K 線，正確的開高低收
  - PLATINUM: 10 筆 K 線，正確的開高低收
  - PALLADIUM: 10 筆 K 線，正確的開高低收
- ✅ 價格緩衝機制運作正常（每秒 3 筆 → 1 筆）
- ✅ 價格策略選擇功能正常（best/worst）

### 技術細節

#### K 線查詢 Flux 優化
```flux
# 修復前（錯誤）
|> aggregateWindow(every: 1m, fn: aggregate, createEmpty: false)

# 修復後（正確）
open = data |> aggregateWindow(every: 1m, fn: first, createEmpty: false)
high = data |> aggregateWindow(every: 1m, fn: max, createEmpty: false)
low = data |> aggregateWindow(every: 1m, fn: min, createEmpty: false)
close = data |> aggregateWindow(every: 1m, fn: last, createEmpty: false)
union(tables: [open, high, low, close]) |> pivot(...)
```

#### 價格處理流程
```
Price Service (每 333ms) 
  → Redis Pub/Sub
    → Platform Subscriber (緩衝)
      → 每秒處理 (選擇 best/worst)
        → 未來推送到 WebSocket
```

### 檔案清單

**新增檔案**:
```
platform/
├── go.mod, go.sum
├── main.go
├── Dockerfile
├── docker-compose.yml
├── .dockerignore
├── .gitignore
├── README.md
├── proto/
│   ├── price.proto
│   ├── price.pb.go
│   └── price_grpc.pb.go
└── internal/
    ├── config/config.go
    ├── model/price.go
    ├── grpc/client.go
    ├── redis/subscriber.go
    └── service/service.go
```

**修改檔案**:
```
price/internal/repository/influxdb.go (K 線查詢修復)
README.md (更新開發進度)
```

### 下一步計劃

**Platform Gateway - Phase 2**:
1. HTTP API 服務器
   - GET /api/prices/current
   - GET /api/prices/history
   - GET /api/user/info (Demo)
2. WebSocket 服務器
   - WS /ws/prices - 即時價格推送
3. 用戶管理整合（簡化版）

---

## 之前的開發

### Price Service (已完成)
- 價格模擬器（幾何布朗運動）
- InfluxDB 時序資料存儲
- Redis Pub/Sub 即時推送
- gRPC 服務完整實現
- Docker 容器化部署

