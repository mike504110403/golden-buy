package service

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/mike/golden-buy/platform/internal/config"
	"github.com/mike/golden-buy/platform/internal/grpc"
	"github.com/mike/golden-buy/platform/internal/model"
	"github.com/mike/golden-buy/platform/internal/redis"
)

// PlatformService 平台服務
type PlatformService struct {
	cfg          *config.Config
	grpcClient   *grpc.PriceClient
	subscriber   *redis.Subscriber
	mu           sync.RWMutex
	latestPrices map[string]*model.Price // 存儲每個商品的最新處理價格
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
}

// New 創建新的平台服務
func New(cfg *config.Config) (*PlatformService, error) {
	// 創建 gRPC 客戶端
	grpcClient, err := grpc.NewPriceClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create grpc client: %w", err)
	}

	// 測試 gRPC 連接
	ctx := context.Background()
	if err := grpcClient.Ping(ctx); err != nil {
		grpcClient.Close()
		return nil, fmt.Errorf("grpc ping failed: %w", err)
	}
	log.Printf("✅ Connected to Price Service at %s", cfg.PriceServiceAddr)

	// 創建 Redis 訂閱器
	subscriber, err := redis.NewSubscriber(cfg)
	if err != nil {
		grpcClient.Close()
		return nil, fmt.Errorf("failed to create redis subscriber: %w", err)
	}

	// 測試 Redis 連接
	if err := subscriber.Ping(ctx); err != nil {
		grpcClient.Close()
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}
	log.Printf("✅ Connected to Redis at %s", cfg.RedisAddr)

	ctx, cancel := context.WithCancel(context.Background())

	return &PlatformService{
		cfg:          cfg,
		grpcClient:   grpcClient,
		subscriber:   subscriber,
		latestPrices: make(map[string]*model.Price),
		ctx:          ctx,
		cancel:       cancel,
	}, nil
}

// Start 啟動服務
func (s *PlatformService) Start() error {
	log.Println("🚀 Starting Platform Service...")

	// 啟動 Redis 訂閱器
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		err := s.subscriber.Start(s.handlePriceUpdate)
		if err != nil && err != context.Canceled {
			log.Printf("❌ Redis subscriber error: %v", err)
		}
	}()

	log.Println("✅ Platform Service started")
	return nil
}

// handlePriceUpdate 處理價格更新（來自 Redis 訂閱器）
func (s *PlatformService) handlePriceUpdate(price *model.Price) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 存儲最新價格
	s.latestPrices[price.Symbol] = price

	// 未來在這裡可以推送到 WebSocket 客戶端
	log.Printf("📊 Latest price updated: %s = %.2f (change: %.2f%%)",
		price.Symbol, price.Price, price.ChangePercent)
}

// GetLatestPrice 獲取最新處理過的價格
func (s *PlatformService) GetLatestPrice(symbol string) (*model.Price, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	price, exists := s.latestPrices[symbol]
	if !exists {
		return nil, fmt.Errorf("no price data for symbol: %s", symbol)
	}

	return price, nil
}

// GetLatestPrices 獲取所有最新價格
func (s *PlatformService) GetLatestPrices() map[string]*model.Price {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 返回副本避免併發問題
	prices := make(map[string]*model.Price, len(s.latestPrices))
	for k, v := range s.latestPrices {
		prices[k] = v
	}

	return prices
}

// GetCurrentPriceFromService 直接從 Price Service 獲取當前價格
func (s *PlatformService) GetCurrentPriceFromService(ctx context.Context, symbol string) (*model.Price, error) {
	return s.grpcClient.GetCurrentPrice(ctx, symbol)
}

// GetCurrentPricesFromService 直接從 Price Service 獲取多個當前價格
func (s *PlatformService) GetCurrentPricesFromService(ctx context.Context, symbols []string) ([]*model.Price, error) {
	return s.grpcClient.GetCurrentPrices(ctx, symbols)
}

// GetKlines 獲取 K 線資料（用於圖表）
func (s *PlatformService) GetKlines(ctx context.Context, symbol, interval string, startTime, endTime int64, limit int32) ([]*model.Kline, error) {
	klines, err := s.grpcClient.GetKlines(ctx, symbol, interval, startTime, endTime, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get klines: %w", err)
	}

	log.Printf("📈 Retrieved %d klines for %s (%s interval)", len(klines), symbol, interval)
	return klines, nil
}

// Stop 停止服務
func (s *PlatformService) Stop() error {
	log.Println("🛑 Stopping Platform Service...")

	// 取消 context
	if s.cancel != nil {
		s.cancel()
	}

	// 停止 Redis 訂閱器
	if err := s.subscriber.Stop(); err != nil {
		log.Printf("❌ Error stopping subscriber: %v", err)
	}

	// 等待所有 goroutines 結束
	s.wg.Wait()

	// 關閉 gRPC 客戶端
	if err := s.grpcClient.Close(); err != nil {
		log.Printf("❌ Error closing grpc client: %v", err)
		return err
	}

	log.Println("✅ Platform Service stopped")
	return nil
}

// GetSubscriber 獲取 Redis 訂閱器（用於調試）
func (s *PlatformService) GetSubscriber() *redis.Subscriber {
	return s.subscriber
}
