package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mike/golden-buy/platform/internal/config"
	httpserver "github.com/mike/golden-buy/platform/internal/http"
	"github.com/mike/golden-buy/platform/internal/service"
)

func main() {
	// 設置日誌格式
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// 載入配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}

	log.Println("🎯 Golden Buy - Platform Service")
	log.Println("========================================")
	log.Printf("Price Service: %s", cfg.PriceServiceAddr)
	log.Printf("Redis: %s", cfg.RedisAddr)
	log.Printf("Price Strategy: %s", cfg.PriceStrategy)
	log.Printf("HTTP Port: %s", cfg.HTTPPort)
	log.Println("========================================")

	// 創建服務
	svc, err := service.New(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to create service: %v", err)
	}

	// 啟動服務
	if err := svc.Start(); err != nil {
		log.Fatalf("❌ Failed to start service: %v", err)
	}

	// 創建 HTTP 服務器
	httpAddr := fmt.Sprintf(":%s", cfg.HTTPPort)
	wsHub := svc.GetWebSocketHub()
	httpServer := httpserver.NewServer(httpAddr, svc, wsHub)

	// 啟動 HTTP 服務器
	go func() {
		if err := httpServer.Start(); err != nil {
			log.Fatalf("❌ Failed to start HTTP server: %v", err)
		}
	}()

	// 測試：獲取 K 線資料
	go testKlines(svc)

	// 等待中斷信號
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	log.Println("\n🔄 Received shutdown signal...")

	// 優雅關閉
	if err := httpServer.Stop(); err != nil {
		log.Printf("❌ Error stopping HTTP server: %v", err)
	}

	if err := svc.Stop(); err != nil {
		log.Printf("❌ Error during shutdown: %v", err)
		os.Exit(1)
	}

	log.Println("👋 Platform Service shut down gracefully")
}

// testKlines 測試 K 線資料獲取功能
func testKlines(svc *service.PlatformService) {
	// 等待 5 秒讓服務完全啟動
	time.Sleep(5 * time.Second)

	log.Println("\n🧪 Testing K-line data retrieval...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	symbols := []string{"GOLD", "SILVER", "PLATINUM", "PALLADIUM"}

	for _, symbol := range symbols {
		// 獲取最近 1 小時的 1 分鐘 K 線
		endTime := time.Now().UnixMilli()
		startTime := endTime - (60 * 60 * 1000) // 1 小時前

		klines, err := svc.GetKlines(ctx, symbol, "1m", startTime, endTime, 10)
		if err != nil {
			log.Printf("❌ Failed to get klines for %s: %v", symbol, err)
			continue
		}

		if len(klines) > 0 {
			log.Printf("✅ [%s] Retrieved %d klines", symbol, len(klines))
			log.Printf("   Latest kline: Open=%.2f, High=%.2f, Low=%.2f, Close=%.2f",
				klines[0].Open, klines[0].High, klines[0].Low, klines[0].Close)
		} else {
			log.Printf("⚠️  [%s] No klines available yet", symbol)
		}
	}

	// 每 10 秒顯示最新價格
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			displayLatestPrices(svc)
		}
	}
}

// displayLatestPrices 顯示最新價格
func displayLatestPrices(svc *service.PlatformService) {
	prices := svc.GetLatestPrices()

	if len(prices) == 0 {
		log.Println("⏳ No prices received yet...")
		return
	}

	log.Println("\n📊 Latest Prices (after processing):")
	log.Println("----------------------------------------")

	symbols := []string{"GOLD", "SILVER", "PLATINUM", "PALLADIUM"}
	for _, symbol := range symbols {
		if price, exists := prices[symbol]; exists {
			log.Printf("%s: $%.2f (%.2f%%)",
				formatSymbol(symbol),
				price.Price,
				price.ChangePercent,
			)
		}
	}

	log.Println("----------------------------------------")
}

// formatSymbol 格式化商品名稱
func formatSymbol(symbol string) string {
	names := map[string]string{
		"GOLD":      "🥇 黃金  ",
		"SILVER":    "🥈 白銀  ",
		"PLATINUM":  "⚪ 鉑金  ",
		"PALLADIUM": "⚫ 鈀金  ",
	}

	if name, exists := names[symbol]; exists {
		return name
	}

	return fmt.Sprintf("   %s", symbol)
}

