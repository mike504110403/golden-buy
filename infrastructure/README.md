# 基礎設施服務

這個目錄包含所有周邊服務的 Docker Compose 配置。

## 啟動服務

```bash
# 啟動所有服務
docker-compose up -d

# 查看服務狀態
docker-compose ps

# 查看日誌
docker-compose logs -f

# 停止所有服務
docker-compose down

# 停止並刪除資料卷（⚠️ 會刪除所有資料）
docker-compose down -v
```

## 服務列表

| 服務 | 端口 | 用途 | 訪問地址 |
|------|------|------|----------|
| **Redis** | 6379 | 快取 + Pub/Sub | `localhost:6379` |
| **InfluxDB** | 8086 | 時序資料庫 | http://localhost:8086 |
| **PostgreSQL** | 5432 | 關聯式資料庫 | `localhost:5432` |
| **Grafana** | 3000 | 資料視覺化 | http://localhost:3000 |

## 預設帳號密碼

### InfluxDB
- URL: http://localhost:8086
- 使用者名稱: `admin`
- 密碼: `admin123456`
- Organization: `golden-buy`
- Bucket: `golden_buy`
- Token: `my-super-secret-auth-token`

### PostgreSQL
- Host: `localhost`
- Port: `5432`
- Database: `golden_buy`
- 使用者名稱: `golden_buy`
- 密碼: `golden_buy_password`

### Grafana
- URL: http://localhost:3000
- 使用者名稱: `admin`
- 密碼: `admin123`

## Grafana 配置 InfluxDB 資料源

1. 訪問 http://localhost:3000
2. 登入（admin/admin123）
3. 左側選單 → Configuration → Data Sources
4. Add data source → InfluxDB
5. 配置：
   ```
   Query Language: Flux
   URL: http://influxdb:8086
   Organization: golden-buy
   Token: my-super-secret-auth-token
   Default Bucket: golden_buy
   ```
6. 點擊 "Save & Test"

## 常用指令

```bash
# 重啟單個服務
docker-compose restart redis

# 查看 Redis 連接
docker exec -it golden-buy-redis redis-cli

# 查看 PostgreSQL
docker exec -it golden-buy-postgres psql -U golden_buy

# 進入 InfluxDB CLI
docker exec -it golden-buy-influxdb influx
```

## 資料持久化

所有資料存儲在 Docker volumes 中：
- `redis-data`: Redis 資料
- `influxdb-data`: InfluxDB 資料
- `postgres-data`: PostgreSQL 資料
- `grafana-data`: Grafana 配置和儀表板

## 網路

所有服務連接到 `golden-buy-network` 橋接網路，服務間可以通過容器名稱互相訪問。

例如：
- Price Service 連接 Redis: `redis:6379`
- Price Service 連接 InfluxDB: `http://influxdb:8086`

