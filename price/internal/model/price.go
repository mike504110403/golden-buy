package model

import "time"

// Symbol 商品代碼
type Symbol string

const (
	SymbolGold      Symbol = "GOLD"
	SymbolSilver    Symbol = "SILVER"
	SymbolPlatinum  Symbol = "PLATINUM"
	SymbolPalladium Symbol = "PALLADIUM"
)

// AllSymbols 所有支援的商品
var AllSymbols = []Symbol{
	SymbolGold,
	SymbolSilver,
	SymbolPlatinum,
	SymbolPalladium,
}

// Price 價格資料
type Price struct {
	Symbol        Symbol    `json:"symbol"`
	Price         float64   `json:"price"`
	Timestamp     time.Time `json:"timestamp"`
	Change        float64   `json:"change"`         // 變化量
	ChangePercent float64   `json:"change_percent"` // 變化百分比
}

// InitialPrices 初始價格配置
var InitialPrices = map[Symbol]float64{
	SymbolGold:      1850.0,
	SymbolSilver:    24.0,
	SymbolPlatinum:  950.0,
	SymbolPalladium: 1280.0,
}

// GetInitialPrice 獲取商品的初始價格
func GetInitialPrice(symbol Symbol) float64 {
	if price, ok := InitialPrices[symbol]; ok {
		return price
	}
	return 100.0 // 預設值
}
