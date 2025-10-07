package main

import (
	"context"
	"log"
	"math"
	"math/rand"
	"time"

	"golden-buy/price/internal/model"
	"golden-buy/price/internal/repository"
)

const (
	influxURL    = "http://localhost:8086"
	influxToken  = "my-super-secret-auth-token"
	influxOrg    = "golden-buy"
	influxBucket = "golden_buy"
)

func main() {
	// 設置時區為 UTC+8 (Asia/Taipei)
	loc, err := time.LoadLocation("Asia/Taipei")
	if err != nil {
		log.Printf("⚠️  載入時區失敗，使用預設時區: %v", err)
	} else {
		time.Local = loc
		log.Printf("✅ 時區設置為: %s", loc.String())
	}

	log.Println("🌱 開始生成歷史數據...")

	// 連接 InfluxDB
	repo, err := repository.NewInfluxDBRepository(influxURL, influxToken, influxOrg, influxBucket)
	if err != nil {
		log.Fatalf("連接 InfluxDB 失敗: %v", err)
	}

	ctx := context.Background()

	// 生成過去 1 天的數據
	// 每秒 3 筆數據，1天 = 24小時 * 3600秒 * 3 = 259,200 筆
	endTime := time.Now()
	startTime := endTime.Add(-24 * time.Hour)

	log.Printf("⏰ 生成時間範圍: %s 到 %s", startTime.Format("2006-01-02 15:04:05"), endTime.Format("2006-01-02 15:04:05"))
	log.Printf("📊 商品數量: %d", len(model.AllSymbols))

	// 初始化每個商品的狀態
	priceStates := make(map[model.Symbol]*PriceState)
	for _, symbol := range model.AllSymbols {
		initialPrice := model.GetInitialPrice(symbol)
		priceStates[symbol] = &PriceState{
			CurrentPrice:  initialPrice,
			PreviousPrice: initialPrice,
		}
	}

	// 每秒生成 3 筆數據
	interval := 333 * time.Millisecond
	volatility := 0.02 // 2% 波動率

	totalPoints := 0
	currentTime := startTime

	for currentTime.Before(endTime) {
		// 為每個商品生成價格
		for _, symbol := range model.AllSymbols {
			state := priceStates[symbol]

			// 使用幾何布朗運動生成新價格
			z := rand.NormFloat64()
			dt := 1.0
			drift := 0.0
			changePercent := (drift-0.5*volatility*volatility)*dt + volatility*math.Sqrt(dt)*z
			newPrice := state.CurrentPrice * math.Exp(changePercent)

			// 確保價格在合理範圍內
			initialPrice := model.GetInitialPrice(symbol)
			if newPrice < initialPrice*0.5 {
				newPrice = initialPrice * 0.5
			} else if newPrice > initialPrice*2.0 {
				newPrice = initialPrice * 2.0
			}

			// 計算變化量
			change := newPrice - state.PreviousPrice
			changePercentValue := (change / state.PreviousPrice) * 100

			// 創建價格對象
			price := &model.Price{
				Symbol:        symbol,
				Price:         newPrice,
				Timestamp:     currentTime,
				Change:        change,
				ChangePercent: changePercentValue,
			}

			// 寫入 InfluxDB
			if err := repo.WritePrice(ctx, price); err != nil {
				log.Printf("❌ 寫入失敗 %s at %s: %v", symbol, currentTime.Format("2006-01-02 15:04:05"), err)
			}

			// 更新狀態
			state.PreviousPrice = state.CurrentPrice
			state.CurrentPrice = newPrice

			totalPoints++
		}

		// 每 10000 筆打印一次進度
		if totalPoints%10000 == 0 {
			progress := float64(currentTime.Sub(startTime)) / float64(endTime.Sub(startTime)) * 100
			log.Printf("📈 進度: %.2f%% (%d 筆數據)", progress, totalPoints)
		}

		// 移動到下一個時間點
		currentTime = currentTime.Add(interval)
	}

	log.Printf("✅ 完成！總共生成 %d 筆歷史數據", totalPoints)
	log.Printf("💾 數據已保存到 InfluxDB")
}

// PriceState 價格狀態
type PriceState struct {
	CurrentPrice  float64
	PreviousPrice float64
}
