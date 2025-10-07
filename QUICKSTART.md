# Golden Buy - 快速啟動指南

## 🚀 一鍵啟動所有服務

本專案使用統一的 Docker Compose 配置，包含所有基礎設施和微服務。

> **注意**: 前端應用獨立運行，不在 Docker Compose 中。

### 前置需求

- Docker >= 20.10
- Docker Compose >= 2.0

### 啟動所有服務

```bash
# 在專案根目錄執行
docker-compose up -d
```

這個命令會啟動：
- ✅ Redis (Port 6379) - 快取和 Pub/Sub
- ✅ InfluxDB (Port 8086) - 時序資料庫
- ✅ PostgreSQL (Port 5432) - 關聯式資料庫
- ✅ Grafana (Port 3000) - 資料視覺化
- ✅ Price Service (Port 50051) - 價格服務
- ✅ Platform Service (Port 8080) - 平台閘道服務

### 查看服務狀態

```bash
# 查看所有服務
docker-compose ps

# 查看服務日誌
docker-compose logs -f

# 查看特定服務日誌
docker-compose logs -f price-service
docker-compose logs -f platform-service
```

### 健康檢查

```bash
# 檢查 Platform Service
curl http://localhost:8080/health

# 預期輸出：
# {"success":true,"data":{"status":"healthy","service":"platform-gateway","timestamp":1234567890}}
```

### 停止所有服務

```bash
# 停止服務（保留數據）
docker-compose stop

# 停止並刪除容器（保留數據卷）
docker-compose down

# 完全清理（包含數據卷）
docker-compose down -v
```

## 📊 服務端點

### Platform Gateway (HTTP API)

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

### WebSocket 連接

```javascript
// 連接 WebSocket
const ws = new WebSocket('ws://localhost:8080/ws/prices');

// 訂閱所有商品
ws.onopen = () => {
  ws.send(JSON.stringify({
    type: 'subscribe',
    symbols: ['GOLD', 'SILVER', 'PLATINUM', 'PALLADIUM']
  }));
};

// 接收價格更新
ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  console.log(message);
};
```

### Grafana 監控

訪問 http://localhost:3000
- 用戶名：`admin`
- 密碼：`admin123`

### InfluxDB 管理

訪問 http://localhost:8086
- 用戶名：`admin`
- 密碼：`admin123456`
- Token：`my-super-secret-auth-token`

## 🔧 進階操作

### 重新構建服務

```bash
# 重新構建所有服務
docker-compose build

# 重新構建特定服務
docker-compose build price-service
docker-compose build platform-service

# 重新構建並啟動
docker-compose up -d --build
```

### 擴展服務

```bash
# 擴展 Platform Service 到 3 個實例
docker-compose up -d --scale platform-service=3
```

### 查看資源使用

```bash
# 查看資源使用情況
docker stats

# 只查看 Golden Buy 服務
docker stats $(docker ps --filter name=golden-buy -q)
```

### 進入容器

```bash
# 進入 Redis 容器
docker exec -it golden-buy-redis redis-cli

# 進入 Price Service 容器
docker exec -it golden-buy-price-service sh

# 進入 Platform Service 容器
docker exec -it golden-buy-platform-service sh
```

## 🐛 故障排除

### 問題 1: 端口被佔用

```bash
# 檢查端口佔用
lsof -i :8080
lsof -i :50051
lsof -i :6379

# 修改端口（在 docker-compose.yml 中）
# 例如：將 8080 改為 8081
ports:
  - "8081:8080"
```

### 問題 2: 服務無法啟動

```bash
# 查看詳細日誌
docker-compose logs service-name

# 重新啟動服務
docker-compose restart service-name

# 完全重建
docker-compose down
docker-compose up -d --build
```

### 問題 3: 數據持久化問題

```bash
# 查看數據卷
docker volume ls | grep golden-buy

# 備份數據卷
docker run --rm -v golden-buy_redis-data:/data -v $(pwd):/backup alpine tar czf /backup/redis-backup.tar.gz -C /data .

# 恢復數據卷
docker run --rm -v golden-buy_redis-data:/data -v $(pwd):/backup alpine tar xzf /backup/redis-backup.tar.gz -C /data
```

## 📈 性能優化

### 生產環境建議

1. **限制資源使用**：

```yaml
services:
  price-service:
    deploy:
      resources:
        limits:
          cpus: '1'
          memory: 512M
        reservations:
          cpus: '0.5'
          memory: 256M
```

2. **設置日誌限制**：

```yaml
services:
  price-service:
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
```

3. **使用生產環境配置**：

```bash
# 設置環境變數
export LOG_LEVEL=warn
export PRICE_STRATEGY=best

# 啟動服務
docker-compose up -d
```

## 🔐 安全建議

1. 修改預設密碼（在 docker-compose.yml 中）
2. 使用環境變數文件（.env）存儲敏感資訊
3. 限制對外暴露的端口
4. 使用 Docker secrets 管理敏感數據（Swarm 模式）

## 📝 下一步

服務啟動成功後，可以：

1. 開發前端應用（Vue 3 + TypeScript）
2. 開發 Order Service（訂單服務）
3. 配置 Grafana 監控儀表板
4. 添加認證和授權機制
5. 部署到生產環境

## 🔗 相關文檔

- [Price Service README](./price/README.md)
- [Platform Service README](./platform/README.md)
- [專案總覽 README](./README.md)
- [開發日誌 CHANGELOG](./CHANGELOG.md)

