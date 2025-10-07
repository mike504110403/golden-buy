package model

import "time"

// Kline K 線資料
type Kline struct {
	Timestamp time.Time `json:"timestamp"` // K 線開始時間
	Open      float64   `json:"open"`
	High      float64   `json:"high"`
	Low       float64   `json:"low"`
	Close     float64   `json:"close"`
	Volume    float64   `json:"volume"`
}

// Interval K 線時間週期
type Interval string

const (
	Interval1m  Interval = "1m"
	Interval5m  Interval = "5m"
	Interval15m Interval = "15m"
	Interval30m Interval = "30m"
	Interval1h  Interval = "1h"
	Interval4h  Interval = "4h"
	Interval1d  Interval = "1d"
)

// IsValidInterval 驗證時間週期是否有效
func IsValidInterval(interval string) bool {
	switch Interval(interval) {
	case Interval1m, Interval5m, Interval15m, Interval30m, Interval1h, Interval4h, Interval1d:
		return true
	default:
		return false
	}
}
