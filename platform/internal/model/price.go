package model

// Price 價格資料結構
type Price struct {
	Symbol        string  `json:"symbol"`
	Price         float64 `json:"price"`
	Timestamp     int64   `json:"timestamp"`
	Change        float64 `json:"change"`
	ChangePercent float64 `json:"change_percent"`
}

// Kline K 線資料結構
type Kline struct {
	Timestamp int64   `json:"timestamp"`
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Volume    float64 `json:"volume"`
}

// PriceBuffer 每秒內價格緩衝區
type PriceBuffer struct {
	Prices    []Price
	Symbol    string
	Timestamp int64 // 秒級時間戳
}

// GetBestPrice 獲取緩衝區內最佳價格（最低買入價）
func (pb *PriceBuffer) GetBestPrice() *Price {
	if len(pb.Prices) == 0 {
		return nil
	}

	best := pb.Prices[0]
	for _, p := range pb.Prices[1:] {
		if p.Price < best.Price {
			best = p
		}
	}

	return &best
}

// GetWorstPrice 獲取緩衝區內最差價格（最高買入價）
func (pb *PriceBuffer) GetWorstPrice() *Price {
	if len(pb.Prices) == 0 {
		return nil
	}

	worst := pb.Prices[0]
	for _, p := range pb.Prices[1:] {
		if p.Price > worst.Price {
			worst = p
		}
	}

	return &worst
}
