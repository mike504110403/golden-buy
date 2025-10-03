package simulator

import (
	"context"
	"log"
	"math"
	"math/rand"
	"sync"
	"time"

	"golden-buy/price/internal/model"
)

// PriceSimulator 價格模擬器
type PriceSimulator struct {
	prices      map[model.Symbol]*PriceState
	mu          sync.RWMutex
	interval    time.Duration
	tickCount   int // 每秒內的計數器
	volatility  float64
	subscribers []chan *model.Price
	subMu       sync.Mutex
}

// PriceState 價格狀態
type PriceState struct {
	CurrentPrice  float64
	PreviousPrice float64
	LastUpdate    time.Time
}

// NewPriceSimulator 創建價格模擬器
func NewPriceSimulator(interval time.Duration, volatility float64) *PriceSimulator {
	sim := &PriceSimulator{
		prices:      make(map[model.Symbol]*PriceState),
		interval:    interval,
		volatility:  volatility,
		tickCount:   0,
		subscribers: make([]chan *model.Price, 0),
	}

	// 初始化所有商品的價格
	for _, symbol := range model.AllSymbols {
		initialPrice := model.GetInitialPrice(symbol)
		sim.prices[symbol] = &PriceState{
			CurrentPrice:  initialPrice,
			PreviousPrice: initialPrice,
			LastUpdate:    time.Now(),
		}
	}

	return sim
}

// Start 啟動價格模擬器
func (s *PriceSimulator) Start(ctx context.Context) {
	// 每秒觸發 3 次更新，間隔 333ms
	ticker := time.NewTicker(333 * time.Millisecond)
	defer ticker.Stop()

	log.Printf("價格模擬器已啟動，每秒 3 次更新，間隔: 333ms")

	for {
		select {
		case <-ctx.Done():
			log.Println("價格模擬器停止")
			return
		case <-ticker.C:
			s.generatePrices()
		}
	}
}

// generatePrices 生成所有商品的新價格
func (s *PriceSimulator) generatePrices() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	// 增加計數器，每 3 次重置為 1（表示新的一秒開始）
	s.tickCount++
	if s.tickCount > 3 {
		s.tickCount = 1
	}

	prices := make([]*model.Price, 0, len(model.AllSymbols))

	for _, symbol := range model.AllSymbols {
		state := s.prices[symbol]

		// 使用幾何布朗運動 (Geometric Brownian Motion) 生成新價格
		// S(t+Δt) = S(t) * exp((μ - σ²/2)Δt + σ√Δt * Z)
		// 簡化版本：S(t+1) = S(t) * (1 + σ * Z)
		// Z 是標準正態分佈的隨機數

		dt := 1.0    // 時間增量（秒）
		drift := 0.0 // 漂移率（趨勢），設為 0 表示無明顯趨勢
		volatility := s.volatility

		// 生成標準正態分佈隨機數
		z := rand.NormFloat64()

		// 計算價格變化
		changePercent := (drift-0.5*volatility*volatility)*dt + volatility*math.Sqrt(dt)*z
		newPrice := state.CurrentPrice * math.Exp(changePercent)

		// 確保價格在合理範圍內（不低於初始價格的 50%，不高於初始價格的 200%）
		initialPrice := model.GetInitialPrice(symbol)
		if newPrice < initialPrice*0.5 {
			newPrice = initialPrice * 0.5
		} else if newPrice > initialPrice*2.0 {
			newPrice = initialPrice * 2.0
		}

		// 計算變化量和百分比
		change := newPrice - state.PreviousPrice
		changePercentValue := (change / state.PreviousPrice) * 100

		// 更新狀態
		state.PreviousPrice = state.CurrentPrice
		state.CurrentPrice = newPrice
		state.LastUpdate = now

		// 創建價格對象
		price := &model.Price{
			Symbol:        symbol,
			Price:         newPrice,
			Timestamp:     now,
			Change:        change,
			ChangePercent: changePercentValue,
		}

		prices = append(prices, price)
	}

	// 通知所有訂閱者
	s.notifySubscribers(prices)

	if len(prices) > 0 {
		log.Printf("更新了 %d 種商品的價格 (第 %d/3 次)", len(prices), s.tickCount)
	}
}

// GetCurrentPrice 獲取指定商品的當前價格
func (s *PriceSimulator) GetCurrentPrice(symbol model.Symbol) *model.Price {
	s.mu.RLock()
	defer s.mu.RUnlock()

	state, ok := s.prices[symbol]
	if !ok {
		return nil
	}

	return &model.Price{
		Symbol:        symbol,
		Price:         state.CurrentPrice,
		Timestamp:     state.LastUpdate,
		Change:        state.CurrentPrice - state.PreviousPrice,
		ChangePercent: ((state.CurrentPrice - state.PreviousPrice) / state.PreviousPrice) * 100,
	}
}

// GetAllPrices 獲取所有商品的當前價格
func (s *PriceSimulator) GetAllPrices() []*model.Price {
	s.mu.RLock()
	defer s.mu.RUnlock()

	prices := make([]*model.Price, 0, len(s.prices))
	for symbol, state := range s.prices {
		prices = append(prices, &model.Price{
			Symbol:        symbol,
			Price:         state.CurrentPrice,
			Timestamp:     state.LastUpdate,
			Change:        state.CurrentPrice - state.PreviousPrice,
			ChangePercent: ((state.CurrentPrice - state.PreviousPrice) / state.PreviousPrice) * 100,
		})
	}

	return prices
}

// Subscribe 訂閱價格更新
func (s *PriceSimulator) Subscribe() chan *model.Price {
	s.subMu.Lock()
	defer s.subMu.Unlock()

	ch := make(chan *model.Price, 100) // 緩衝區
	s.subscribers = append(s.subscribers, ch)
	return ch
}

// Unsubscribe 取消訂閱
func (s *PriceSimulator) Unsubscribe(ch chan *model.Price) {
	s.subMu.Lock()
	defer s.subMu.Unlock()

	for i, sub := range s.subscribers {
		if sub == ch {
			close(ch)
			s.subscribers = append(s.subscribers[:i], s.subscribers[i+1:]...)
			break
		}
	}
}

// notifySubscribers 通知所有訂閱者
func (s *PriceSimulator) notifySubscribers(prices []*model.Price) {
	s.subMu.Lock()
	defer s.subMu.Unlock()

	for _, price := range prices {
		for _, ch := range s.subscribers {
			select {
			case ch <- price:
			default:
				// 如果通道已滿，跳過
				log.Printf("訂閱者通道已滿，跳過價格推送: %s", price.Symbol)
			}
		}
	}
}
