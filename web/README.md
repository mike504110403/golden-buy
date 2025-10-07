# Golden Buy - 前端應用

基於 Vue 3 + TypeScript + Vite + Tailwind CSS + Element Plus 的現代化前端應用。

## 技術棧

- **框架**: Vue 3 (Composition API)
- **語言**: TypeScript
- **構建工具**: Vite
- **包管理**: pnpm
- **狀態管理**: Pinia
- **路由**: Vue Router 4
- **UI 框架**: 
  - Tailwind CSS (基礎樣式、佈局)
  - Element Plus (複雜組件)
- **圖表**: TradingView Lightweight Charts
- **HTTP 客戶端**: Axios
- **日期處理**: dayjs

## 快速開始

```bash
# 安裝依賴
pnpm install

# 開發模式
pnpm dev

# 構建生產版本
pnpm build

# 預覽生產版本
pnpm preview
```

## 專案結構

```
web/
├── src/
│   ├── api/                 # API 服務
│   │   ├── index.ts        # HTTP API
│   │   └── websocket.ts    # WebSocket 服務
│   ├── assets/             # 靜態資源
│   ├── components/         # 組件
│   │   ├── common/         # 通用組件
│   │   └── charts/         # 圖表組件
│   ├── composables/        # 組合式函數
│   ├── router/             # 路由配置
│   ├── stores/             # Pinia 狀態管理
│   │   ├── price.ts        # 價格狀態
│   │   └── user.ts         # 用戶狀態
│   ├── types/              # TypeScript 類型定義
│   ├── utils/              # 工具函數
│   │   ├── constants.ts    # 常量
│   │   └── format.ts       # 格式化函數
│   ├── views/              # 頁面視圖
│   │   ├── dashboard/      # 儀表板
│   │   └── orders/         # 訂單管理
│   ├── App.vue             # 根組件
│   ├── main.ts             # 應用入口
│   └── style.css           # 全局樣式
├── .env.development        # 開發環境配置
├── .env.production         # 生產環境配置
├── index.html              # HTML 模板
├── tailwind.config.js      # Tailwind CSS 配置
├── postcss.config.js       # PostCSS 配置
├── tsconfig.json           # TypeScript 配置
└── vite.config.ts          # Vite 配置
```

## 主要功能

### 已實現
- ✅ 專案架構搭建
- ✅ Tailwind CSS + Element Plus 整合
- ✅ API 服務封裝
- ✅ WebSocket 自動重連機制
- ✅ Pinia 狀態管理
- ✅ 價格卡片組件
- ✅ 即時價格更新
- ✅ 格式化工具函數

### 待實現
- [ ] TradingView K 線圖表
- [ ] 訂單創建功能
- [ ] 訂單列表和管理
- [ ] 響應式設計優化
- [ ] 深色模式
- [ ] 國際化（i18n）

## 環境變數

開發環境 (`.env.development`):
```
VITE_API_BASE_URL=http://localhost:8080
VITE_WS_URL=ws://localhost:8080/ws/prices
```

生產環境 (`.env.production`):
```
VITE_API_BASE_URL=https://api.golden-buy.com
VITE_WS_URL=wss://api.golden-buy.com/ws/prices
```

## API 端點

### HTTP API
- GET `/health` - 健康檢查
- GET `/api/prices/current` - 獲取當前價格
- GET `/api/prices/history` - 獲取 K 線資料
- GET `/api/user/info` - 獲取用戶資訊

### WebSocket
- WS `/ws/prices` - 即時價格推送

## 開發指南

### 添加新頁面
1. 在 `src/views/` 創建頁面組件
2. 在 `src/router/index.ts` 添加路由

### 添加新組件
1. 在 `src/components/` 創建組件
2. 使用 Tailwind CSS 和 Element Plus 混合樣式

### 調用 API
```typescript
import { priceApi } from '@/api'

const prices = await priceApi.getCurrentPrices()
```

### 使用 WebSocket
```typescript
import { wsService } from '@/api/websocket'

await wsService.connect()
wsService.subscribe(['GOLD', 'SILVER'])
wsService.onPriceUpdate((data) => {
  console.log(data)
})
```

### 使用 Pinia Store
```typescript
import { usePriceStore } from '@/stores/price'

const priceStore = usePriceStore()
await priceStore.fetchCurrentPrices()
```

## 樣式規範

### Tailwind CSS
用於基礎樣式和佈局：
- 間距、顏色、字體
- Flexbox / Grid 佈局
- 響應式設計
- 動畫效果

### Element Plus
用於複雜組件：
- 表格、表單
- 彈窗、提示
- 下拉選單
- 日期選擇器

## 注意事項

1. 確保後端服務（Platform Service）已啟動
2. WebSocket 會自動重連，最多嘗試 5 次
3. 所有 API 調用都有錯誤處理
4. 價格更新會觸發動畫效果

## 後續開發

- [ ] 整合 TradingView Lightweight Charts
- [ ] 實現訂單功能
- [ ] 添加更多圖表類型
- [ ] 優化移動端體驗
- [ ] 添加單元測試
