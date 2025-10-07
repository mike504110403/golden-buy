package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/mike/golden-buy/platform/internal/config"
	"github.com/mike/golden-buy/platform/internal/model"
	"github.com/redis/go-redis/v9"
)

const (
	// PriceUpdatesChannel Redis Pub/Sub 頻道名稱
	PriceUpdatesChannel = "price:updates"
)

// PriceHandler 價格處理回調函數
type PriceHandler func(*model.Price)

// Subscriber Redis 訂閱器
type Subscriber struct {
	client  *redis.Client
	cfg     *config.Config
	mu      sync.RWMutex
	buffers map[string]*model.PriceBuffer // symbol -> buffer
	ticker  *time.Ticker
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

// NewSubscriber 創建新的 Redis 訂閱器
func NewSubscriber(cfg *config.Config) (*Subscriber, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	// 測試連接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis at %s: %w", cfg.RedisAddr, err)
	}

	ctx, cancelFunc := context.WithCancel(context.Background())

	sub := &Subscriber{
		client:  client,
		cfg:     cfg,
		buffers: make(map[string]*model.PriceBuffer),
		ticker:  time.NewTicker(1 * time.Second),
		ctx:     ctx,
		cancel:  cancelFunc,
	}

	return sub, nil
}

// Start 開始訂閱 Redis 價格更新
func (s *Subscriber) Start(handler PriceHandler) error {
	// 訂閱價格更新頻道
	pubsub := s.client.Subscribe(s.ctx, PriceUpdatesChannel)
	defer pubsub.Close()

	// 確認訂閱成功
	_, err := pubsub.Receive(s.ctx)
	if err != nil {
		return fmt.Errorf("failed to subscribe to channel %s: %w", PriceUpdatesChannel, err)
	}

	log.Printf("✅ Subscribed to Redis channel: %s", PriceUpdatesChannel)
	log.Printf("📊 Price strategy: %s", s.cfg.PriceStrategy)

	// 啟動定時處理器（每秒處理一次緩衝區）
	s.wg.Add(1)
	go s.processBuffers(handler)

	// 接收訊息
	ch := pubsub.Channel()
	for {
		select {
		case <-s.ctx.Done():
			log.Println("📴 Subscriber context cancelled")
			return s.ctx.Err()

		case msg, ok := <-ch:
			if !ok {
				return fmt.Errorf("redis channel closed")
			}

			// 解析價格更新
			var price model.Price
			if err := json.Unmarshal([]byte(msg.Payload), &price); err != nil {
				log.Printf("❌ Failed to unmarshal price: %v", err)
				continue
			}

			// 將價格加入緩衝區
			s.addToBuffer(&price)
		}
	}
}

// addToBuffer 將價格加入緩衝區
func (s *Subscriber) addToBuffer(price *model.Price) {
	s.mu.Lock()
	defer s.mu.Unlock()

	symbol := price.Symbol
	currentSecond := price.Timestamp / 1000 // 轉換為秒級時間戳

	// 獲取或創建該商品的緩衝區
	buffer, exists := s.buffers[symbol]
	if !exists || buffer.Timestamp != currentSecond {
		// 創建新的緩衝區（新的一秒）
		buffer = &model.PriceBuffer{
			Symbol:    symbol,
			Timestamp: currentSecond,
			Prices:    make([]model.Price, 0, 3), // 預分配空間給 3 筆價格
		}
		s.buffers[symbol] = buffer
	}

	// 加入價格到緩衝區
	buffer.Prices = append(buffer.Prices, *price)

	// 日誌記錄（調試用）
	if len(buffer.Prices) == 1 {
		log.Printf("📝 [%s] New second buffer: %d", symbol, currentSecond)
	}
}

// processBuffers 定時處理緩衝區（每秒執行一次）
func (s *Subscriber) processBuffers(handler PriceHandler) {
	defer s.wg.Done()

	for {
		select {
		case <-s.ctx.Done():
			return

		case <-s.ticker.C:
			s.flushBuffers(handler)
		}
	}
}

// flushBuffers 清空緩衝區並選擇最佳/最差價格
func (s *Subscriber) flushBuffers(handler PriceHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()

	currentSecond := time.Now().Unix()

	for symbol, buffer := range s.buffers {
		// 只處理上一秒的數據（確保該秒已收集完整）
		if buffer.Timestamp < currentSecond-1 {
			// 該緩衝區太舊，直接刪除
			delete(s.buffers, symbol)
			continue
		}

		if buffer.Timestamp == currentSecond-1 {
			// 處理上一秒的完整數據
			if len(buffer.Prices) > 0 {
				var selectedPrice *model.Price

				// 根據策略選擇價格
				if s.cfg.PriceStrategy == "best" {
					selectedPrice = buffer.GetBestPrice()
				} else {
					selectedPrice = buffer.GetWorstPrice()
				}

				if selectedPrice != nil {
					log.Printf("💰 [%s] Selected %s price: %.2f (from %d prices)",
						symbol, s.cfg.PriceStrategy, selectedPrice.Price, len(buffer.Prices))

					// 調用處理器
					handler(selectedPrice)
				}
			}

			// 刪除已處理的緩衝區
			delete(s.buffers, symbol)
		}
		// 如果 buffer.Timestamp == currentSecond，則保留（還在收集中）
	}
}

// GetCurrentBuffer 獲取當前緩衝區狀態（調試用）
func (s *Subscriber) GetCurrentBuffer(symbol string) *model.PriceBuffer {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if buffer, exists := s.buffers[symbol]; exists {
		// 返回副本避免併發問題
		bufferCopy := &model.PriceBuffer{
			Symbol:    buffer.Symbol,
			Timestamp: buffer.Timestamp,
			Prices:    make([]model.Price, len(buffer.Prices)),
		}
		copy(bufferCopy.Prices, buffer.Prices)
		return bufferCopy
	}

	return nil
}

// Stop 停止訂閱器
func (s *Subscriber) Stop() error {
	log.Println("🛑 Stopping Redis subscriber...")

	// 停止定時器
	if s.ticker != nil {
		s.ticker.Stop()
	}

	// 取消 context
	if s.cancel != nil {
		s.cancel()
	}

	// 等待 goroutines 結束
	s.wg.Wait()

	// 關閉 Redis 連接
	if s.client != nil {
		if err := s.client.Close(); err != nil {
			return fmt.Errorf("failed to close redis client: %w", err)
		}
	}

	log.Println("✅ Redis subscriber stopped")
	return nil
}

// Ping 檢查 Redis 連接是否正常
func (s *Subscriber) Ping(ctx context.Context) error {
	return s.client.Ping(ctx).Err()
}

