package http

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mike/golden-buy/platform/internal/model"
	"github.com/mike/golden-buy/platform/internal/service"
)

// Handler HTTP 處理器
type Handler struct {
	service *service.PlatformService
}

// NewHandler 創建新的 HTTP 處理器
func NewHandler(svc *service.PlatformService) *Handler {
	return &Handler{
		service: svc,
	}
}

// Response 統一回應結構
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// PriceResponse 價格回應
type PriceResponse struct {
	Symbol        string    `json:"symbol"`
	Price         float64   `json:"price"`
	ChangePercent float64   `json:"change_percent"`
	Timestamp     int64     `json:"timestamp"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// KlineResponse K 線回應
type KlineResponse struct {
	Symbol   string         `json:"symbol"`
	Interval string         `json:"interval"`
	Klines   []*model.Kline `json:"klines"`
	Count    int            `json:"count"`
}

// UserResponse 用戶回應
type UserResponse struct {
	ID       string  `json:"id"`
	Username string  `json:"username"`
	Email    string  `json:"email"`
	Balance  float64 `json:"balance"`
	Role     string  `json:"role"`
}

// HandleGetCurrentPrice 獲取當前價格
// GET /api/prices/current?symbol=GOLD
// GET /api/prices/current (返回所有商品)
func (h *Handler) HandleGetCurrentPrice(c *gin.Context) {
	// 獲取查詢參數
	symbol := strings.ToUpper(c.Query("symbol"))

	// 如果指定了商品，返回單個商品價格
	if symbol != "" {
		// 優先從緩存獲取
		price, err := h.service.GetLatestPrice(symbol)
		if err != nil {
			// 緩存沒有，從 Price Service 獲取
			price, err = h.service.GetCurrentPriceFromService(c.Request.Context(), symbol)
			if err != nil {
				log.Printf("❌ Failed to get price for %s: %v", symbol, err)
				c.JSON(http.StatusInternalServerError, Response{
					Success: false,
					Error:   "Failed to get price",
				})
				return
			}
		}

		c.JSON(http.StatusOK, Response{
			Success: true,
			Data:    convertToResponse(price),
		})
		return
	}

	// 否則返回所有商品價格
	prices := h.service.GetLatestPrices()

	// 如果緩存為空，從 Price Service 獲取
	if len(prices) == 0 {
		symbols := []string{"GOLD", "SILVER", "PLATINUM", "PALLADIUM"}
		servicePrices, err := h.service.GetCurrentPricesFromService(c.Request.Context(), symbols)
		if err != nil {
			log.Printf("❌ Failed to get prices: %v", err)
			c.JSON(http.StatusInternalServerError, Response{
				Success: false,
				Error:   "Failed to get prices",
			})
			return
		}

		// 轉換為 map
		priceMap := make(map[string]*PriceResponse)
		for _, p := range servicePrices {
			priceMap[p.Symbol] = convertToResponse(p)
		}

		c.JSON(http.StatusOK, Response{
			Success: true,
			Data:    priceMap,
		})
		return
	}

	// 轉換為回應格式
	priceMap := make(map[string]*PriceResponse)
	for symbol, price := range prices {
		priceMap[symbol] = convertToResponse(price)
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    priceMap,
	})
}

// HandleGetHistory 獲取歷史 K 線資料
// GET /api/prices/history?symbol=GOLD&interval=1m&start=1234567890000&end=1234567899000&limit=100
func (h *Handler) HandleGetHistory(c *gin.Context) {
	// 解析查詢參數
	symbol := strings.ToUpper(c.Query("symbol"))
	interval := c.DefaultQuery("interval", "1m")

	// 必需參數檢查
	if symbol == "" {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Error:   "symbol is required",
		})
		return
	}

	// 解析時間範圍
	var startTime, endTime int64
	var limit int32 = 100 // 預設 100 筆

	if startStr := c.Query("start"); startStr != "" {
		if val, err := strconv.ParseInt(startStr, 10, 64); err == nil {
			startTime = val
		}
	}

	if endStr := c.Query("end"); endStr != "" {
		if val, err := strconv.ParseInt(endStr, 10, 64); err == nil {
			endTime = val
		}
	}

	// 如果沒有指定時間範圍，使用最近 1 小時
	if startTime == 0 || endTime == 0 {
		endTime = time.Now().UnixMilli()
		startTime = endTime - (60 * 60 * 1000) // 1 小時前
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if val, err := strconv.ParseInt(limitStr, 10, 32); err == nil {
			limit = int32(val)
		}
	}

	// 從服務獲取 K 線資料
	klines, err := h.service.GetKlines(c.Request.Context(), symbol, interval, startTime, endTime, limit)
	if err != nil {
		log.Printf("❌ Failed to get klines for %s %s: %v", symbol, interval, err)

		// 生成填充的空數據，而不是返回錯誤
		filledKlines := h.generateFilledKlines(symbol, interval, startTime, endTime, limit)
		log.Printf("📊 Generated %d filled klines for %s %s", len(filledKlines), symbol, interval)

		response := &KlineResponse{
			Symbol:   symbol,
			Interval: interval,
			Klines:   filledKlines,
			Count:    len(filledKlines),
		}
		c.JSON(http.StatusOK, Response{
			Success: true,
			Data:    response,
			Message: fmt.Sprintf("Generated filled data for %s %s (no historical data available)", symbol, interval),
		})
		return
	}

	response := &KlineResponse{
		Symbol:   symbol,
		Interval: interval,
		Klines:   klines,
		Count:    len(klines),
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    response,
	})
}

// generateFilledKlines 生成填充的空 K 線數據
func (h *Handler) generateFilledKlines(symbol, interval string, startTime, endTime int64, limit int32) []*model.Kline {
	var klines []*model.Kline

	// 計算時間間隔（毫秒）
	intervalMs := h.getIntervalMs(interval)
	if intervalMs == 0 {
		intervalMs = 60 * 1000 // 預設 1 分鐘
	}

	// 生成時間序列
	currentTime := startTime
	count := 0

	for currentTime < endTime && count < int(limit) {
		klines = append(klines, &model.Kline{
			Timestamp: currentTime,
			Open:      0,
			High:      0,
			Low:       0,
			Close:     0,
		})
		currentTime += intervalMs
		count++
	}

	return klines
}

// getIntervalMs 獲取時間間隔的毫秒數
func (h *Handler) getIntervalMs(interval string) int64 {
	switch interval {
	case "1m":
		return 60 * 1000
	case "5m":
		return 5 * 60 * 1000
	case "15m":
		return 15 * 60 * 1000
	case "30m":
		return 30 * 60 * 1000
	case "1h":
		return 60 * 60 * 1000
	case "4h":
		return 4 * 60 * 60 * 1000
	case "1d":
		return 24 * 60 * 60 * 1000
	default:
		return 60 * 1000 // 預設 1 分鐘
	}
}

// HandleGetUserInfo 獲取用戶資訊（Demo 版本）
// GET /api/user/info
func (h *Handler) HandleGetUserInfo(c *gin.Context) {
	// Demo 用戶資料（未來可以從資料庫獲取）
	user := &UserResponse{
		ID:       "demo-user-001",
		Username: "demo_user",
		Email:    "demo@golden-buy.com",
		Balance:  10000.00,
		Role:     "demo",
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    user,
	})
}

// HandleHealthCheck 健康檢查
// GET /health
func (h *Handler) HandleHealthCheck(c *gin.Context) {
	health := map[string]interface{}{
		"status":    "healthy",
		"service":   "platform-gateway",
		"timestamp": time.Now().Unix(),
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    health,
	})
}

// convertToResponse 轉換 Price 為回應格式
func convertToResponse(price *model.Price) *PriceResponse {
	return &PriceResponse{
		Symbol:        price.Symbol,
		Price:         price.Price,
		ChangePercent: price.ChangePercent,
		Timestamp:     price.Timestamp,
		UpdatedAt:     time.Unix(0, price.Timestamp*int64(time.Millisecond)),
	}
}
