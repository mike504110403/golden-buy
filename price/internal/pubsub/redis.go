package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"golden-buy/price/internal/model"

	"github.com/redis/go-redis/v9"
)

const (
	// PriceUpdatesChannel Redis Pub/Sub 頻道名稱
	PriceUpdatesChannel = "price:updates"

	// PriceSecondChannel Redis 每秒價格記錄頻道
	PriceSecondChannel = "price:second"
)

// Publisher 價格發布者
type Publisher struct {
	client *redis.Client
}

// NewPublisher 創建價格發布者
func NewPublisher(addr, password string, db int) (*Publisher, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// 測試連接
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	log.Println("Redis 連接成功")

	return &Publisher{
		client: client,
	}, nil
}

// Publish 發布價格更新
func (p *Publisher) Publish(ctx context.Context, price *model.Price) error {
	// 將價格轉換為 Unix 毫秒時間戳格式
	priceData := map[string]interface{}{
		"symbol":         price.Symbol,
		"price":          price.Price,
		"timestamp":      price.Timestamp.UnixMilli(), // Unix 毫秒時間戳
		"change":         price.Change,
		"change_percent": price.ChangePercent,
	}

	data, err := json.Marshal(priceData)
	if err != nil {
		return err
	}

	// 發布到 Redis
	return p.client.Publish(ctx, PriceUpdatesChannel, data).Err()
}

// SetCache 設置價格快取
func (p *Publisher) SetCache(ctx context.Context, symbol model.Symbol, price *model.Price) error {
	key := "price:" + string(symbol)

	// 將價格轉換為 Unix 毫秒時間戳格式
	priceData := map[string]interface{}{
		"symbol":         price.Symbol,
		"price":          price.Price,
		"timestamp":      price.Timestamp.UnixMilli(), // Unix 毫秒時間戳
		"change":         price.Change,
		"change_percent": price.ChangePercent,
	}

	data, err := json.Marshal(priceData)
	if err != nil {
		return err
	}

	// 設置快取，TTL 60 秒
	return p.client.Set(ctx, key, data, 60*time.Second).Err()
}

// AddSecondPrice 添加每秒內的價格記錄
func (p *Publisher) AddSecondPrice(ctx context.Context, price *model.Price) error {
	// 生成每秒的 key，格式：price:second:GOLD:1640995200000 (Unix 毫秒時間戳)
	secondTimestamp := price.Timestamp.Truncate(time.Second).UnixMilli()
	secondKey := fmt.Sprintf("price:second:%s:%d",
		string(price.Symbol),
		secondTimestamp)

	// 將價格轉換為 Unix 毫秒時間戳格式
	priceData := map[string]interface{}{
		"symbol":         price.Symbol,
		"price":          price.Price,
		"timestamp":      price.Timestamp.UnixMilli(), // Unix 毫秒時間戳
		"change":         price.Change,
		"change_percent": price.ChangePercent,
	}

	data, err := json.Marshal(priceData)
	if err != nil {
		return err
	}

	// 使用 Redis List 存儲每秒內的價格，最多 3 筆
	pipe := p.client.Pipeline()
	pipe.RPush(ctx, secondKey, data)
	pipe.Expire(ctx, secondKey, 10*time.Minute) // 10 分鐘後自動過期
	pipe.LTrim(ctx, secondKey, -3, -1)          // 只保留最後 3 筆

	_, err = pipe.Exec(ctx)
	return err
}

// GetSecondPrices 獲取指定秒內的所有價格
func (p *Publisher) GetSecondPrices(ctx context.Context, symbol model.Symbol, timestamp time.Time) ([]*model.Price, error) {
	// 使用 Unix 毫秒時間戳作為 key
	secondTimestamp := timestamp.Truncate(time.Second).UnixMilli()
	secondKey := fmt.Sprintf("price:second:%s:%d",
		string(symbol),
		secondTimestamp)

	// 獲取 List 中的所有價格
	results, err := p.client.LRange(ctx, secondKey, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	var prices []*model.Price
	for _, result := range results {
		// 解析包含 Unix 毫秒時間戳的資料
		var priceData map[string]interface{}
		if err := json.Unmarshal([]byte(result), &priceData); err != nil {
			log.Printf("解析價格失敗: %v", err)
			continue
		}

		// 轉換回 model.Price 格式
		timestampMs := int64(priceData["timestamp"].(float64))
		price := &model.Price{
			Symbol:        model.Symbol(priceData["symbol"].(string)),
			Price:         priceData["price"].(float64),
			Timestamp:     time.UnixMilli(timestampMs),
			Change:        priceData["change"].(float64),
			ChangePercent: priceData["change_percent"].(float64),
		}

		prices = append(prices, price)
	}

	return prices, nil
}

// GetCache 獲取價格快取
func (p *Publisher) GetCache(ctx context.Context, symbol model.Symbol) (*model.Price, error) {
	key := "price:" + string(symbol)
	data, err := p.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	// 解析包含 Unix 毫秒時間戳的資料
	var priceData map[string]interface{}
	if err := json.Unmarshal([]byte(data), &priceData); err != nil {
		return nil, err
	}

	// 轉換回 model.Price 格式
	timestampMs := int64(priceData["timestamp"].(float64))
	price := &model.Price{
		Symbol:        model.Symbol(priceData["symbol"].(string)),
		Price:         priceData["price"].(float64),
		Timestamp:     time.UnixMilli(timestampMs),
		Change:        priceData["change"].(float64),
		ChangePercent: priceData["change_percent"].(float64),
	}

	return price, nil
}

// Close 關閉連接
func (p *Publisher) Close() error {
	return p.client.Close()
}
