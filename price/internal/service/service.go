package service

import (
	"context"
	"log"
	"time"

	"golden-buy/price/internal/model"
	"golden-buy/price/internal/pubsub"
	"golden-buy/price/internal/repository"
	"golden-buy/price/internal/simulator"
)

// PriceService 價格服務（業務邏輯層）
type PriceService struct {
	simulator  *simulator.PriceSimulator
	influxRepo *repository.InfluxDBRepository
	publisher  *pubsub.Publisher
}

// NewPriceService 創建價格服務
func NewPriceService(
	sim *simulator.PriceSimulator,
	influxRepo *repository.InfluxDBRepository,
	publisher *pubsub.Publisher,
) *PriceService {
	return &PriceService{
		simulator:  sim,
		influxRepo: influxRepo,
		publisher:  publisher,
	}
}

// Start 啟動價格服務（監聽模擬器並處理價格更新）
func (s *PriceService) Start(ctx context.Context) {
	// 訂閱價格模擬器
	priceChan := s.simulator.Subscribe()
	defer s.simulator.Unsubscribe(priceChan)

	for {
		select {
		case <-ctx.Done():
			return
		case price := <-priceChan:
			if price == nil {
				continue
			}

			// 寫入 InfluxDB
			if err := s.influxRepo.WritePrice(ctx, price); err != nil {
				log.Printf("寫入 InfluxDB 失敗: %v", err)
			}

			// 發布到 Redis Pub/Sub
			if err := s.publisher.Publish(ctx, price); err != nil {
				log.Printf("發布到 Redis 失敗: %v", err)
			}

			// 設置 Redis 快取
			if err := s.publisher.SetCache(ctx, price.Symbol, price); err != nil {
				log.Printf("設置 Redis 快取失敗: %v", err)
			}

			// 添加每秒價格記錄
			if err := s.publisher.AddSecondPrice(ctx, price); err != nil {
				log.Printf("添加每秒價格記錄失敗: %v", err)
			}
		}
	}
}

// GetCurrentPrice 獲取當前價格
func (s *PriceService) GetCurrentPrice(ctx context.Context, symbol model.Symbol) (*model.Price, error) {
	// 1. 先從模擬器獲取最新價格
	if price := s.simulator.GetCurrentPrice(symbol); price != nil {
		return price, nil
	}

	// 2. 如果模擬器沒有，從 Redis 快取讀取
	if price, err := s.publisher.GetCache(ctx, symbol); err == nil && price != nil {
		return price, nil
	}

	// 3. 如果快取沒有，從 InfluxDB 查詢
	return s.influxRepo.GetLatestPrice(ctx, symbol)
}

// GetCurrentPrices 獲取多個商品的當前價格
func (s *PriceService) GetCurrentPrices(ctx context.Context, symbols []model.Symbol) ([]*model.Price, error) {
	var prices []*model.Price

	for _, symbol := range symbols {
		price, err := s.GetCurrentPrice(ctx, symbol)
		if err != nil {
			log.Printf("獲取 %s 價格失敗: %v", symbol, err)
			continue
		}
		prices = append(prices, price)
	}

	return prices, nil
}

// GetKlines 獲取 K 線資料
func (s *PriceService) GetKlines(ctx context.Context, symbol model.Symbol, interval string, startTime, endTime int64, limit int) ([]*model.Kline, error) {
	return s.influxRepo.GetKlines(ctx, symbol, interval, startTime, endTime, limit)
}

// SubscribePrices 訂閱價格更新
func (s *PriceService) SubscribePrices(symbols []model.Symbol) chan *model.Price {
	// 直接從模擬器訂閱
	return s.simulator.Subscribe()
}

// UnsubscribePrices 取消訂閱
func (s *PriceService) UnsubscribePrices(ch chan *model.Price) {
	// 從模擬器取消訂閱
	s.simulator.Unsubscribe(ch)
}

// GetSecondPrices 獲取指定秒內的所有價格
func (s *PriceService) GetSecondPrices(ctx context.Context, symbol model.Symbol, timestamp time.Time) ([]*model.Price, error) {
	return s.publisher.GetSecondPrices(ctx, symbol, timestamp)
}
