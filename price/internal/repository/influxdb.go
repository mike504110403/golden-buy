package repository

import (
	"context"
	"fmt"
	"log"
	"time"

	"golden-buy/price/internal/model"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

// InfluxDBRepository InfluxDB 存儲層
type InfluxDBRepository struct {
	client        influxdb2.Client
	writeAPI      api.WriteAPI
	writeAPIBlock api.WriteAPIBlocking
	queryAPI      api.QueryAPI
	org           string
	bucket        string
}

// NewInfluxDBRepository 創建 InfluxDB 存儲層
func NewInfluxDBRepository(url, token, org, bucket string) (*InfluxDBRepository, error) {
	// 創建 InfluxDB 客戶端
	client := influxdb2.NewClient(url, token)

	// 測試連接
	health, err := client.Health(context.Background())
	if err != nil {
		return nil, fmt.Errorf("InfluxDB 連接失敗: %v", err)
	}

	if health.Status != "pass" {
		return nil, fmt.Errorf("InfluxDB 健康檢查失敗: %s", health.Status)
	}

	log.Printf("InfluxDB 連接成功: %s", health.Status)

	return &InfluxDBRepository{
		client:        client,
		writeAPI:      client.WriteAPI(org, bucket),
		writeAPIBlock: client.WriteAPIBlocking(org, bucket),
		queryAPI:      client.QueryAPI(org),
		org:           org,
		bucket:        bucket,
	}, nil
}

// WritePrice 寫入單個價格
func (r *InfluxDBRepository) WritePrice(ctx context.Context, price *model.Price) error {
	// 創建資料點
	point := write.NewPointWithMeasurement("prices").
		AddTag("symbol", string(price.Symbol)).
		AddField("price", price.Price).
		AddField("change", price.Change).
		AddField("change_percent", price.ChangePercent).
		SetTime(price.Timestamp)

	// 寫入 InfluxDB
	err := r.writeAPIBlock.WritePoint(ctx, point)
	if err != nil {
		return fmt.Errorf("寫入價格失敗: %v", err)
	}

	return nil
}

// GetLatestPrice 獲取最新價格
func (r *InfluxDBRepository) GetLatestPrice(ctx context.Context, symbol model.Symbol) (*model.Price, error) {
	// Flux 查詢語句
	query := fmt.Sprintf(`
		from(bucket: "%s")
		|> range(start: -1h)
		|> filter(fn: (r) => r["_measurement"] == "prices")
		|> filter(fn: (r) => r["symbol"] == "%s")
		|> filter(fn: (r) => r["_field"] == "price")
		|> last()
	`, r.bucket, string(symbol))

	// 執行查詢
	result, err := r.queryAPI.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("查詢最新價格失敗: %v", err)
	}

	// 解析結果
	for result.Next() {
		record := result.Record()
		if record.Table() == 0 {
			price := &model.Price{
				Symbol:    symbol,
				Price:     record.Value().(float64),
				Timestamp: record.Time(),
			}
			return price, nil
		}
	}

	return nil, fmt.Errorf("未找到 %s 的最新價格", symbol)
}

// GetKlines 獲取 K 線資料
func (r *InfluxDBRepository) GetKlines(ctx context.Context, symbol model.Symbol, interval string, startTime, endTime int64, limit int) ([]*model.Kline, error) {
	// 轉換時間戳為 RFC3339 格式
	start := time.UnixMilli(startTime).Format(time.RFC3339)
	end := time.UnixMilli(endTime).Format(time.RFC3339)

	// Flux 查詢語句 - 使用 aggregateWindow 聚合
	query := fmt.Sprintf(`
		from(bucket: "%s")
		|> range(start: %s, stop: %s)
		|> filter(fn: (r) => r["_measurement"] == "prices")
		|> filter(fn: (r) => r["symbol"] == "%s")
		|> filter(fn: (r) => r["_field"] == "price")
		|> aggregateWindow(every: %s, fn: aggregate, createEmpty: false)
		|> limit(n: %d)
	`, r.bucket, start, end, string(symbol), interval, limit)

	// 執行查詢
	result, err := r.queryAPI.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("查詢 K 線失敗: %v", err)
	}

	var klines []*model.Kline
	for result.Next() {
		record := result.Record()
		if record.Table() == 0 {
			kline := &model.Kline{
				Timestamp: record.Time(),
				Close:     record.Value().(float64),
				Volume:    0, // 模擬器暫時不生成成交量
			}
			klines = append(klines, kline)
		}
	}

	return klines, nil
}

// Close 關閉連接
func (r *InfluxDBRepository) Close() {
	r.client.Close()
	log.Println("InfluxDB 連接已關閉")
}
