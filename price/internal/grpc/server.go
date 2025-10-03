package grpc

import (
	"context"
	"fmt"
	"log"
	"time"

	"golden-buy/price/internal/model"
	"golden-buy/price/internal/service"
	pb "golden-buy/price/proto"
)

// PriceServiceServer gRPC 服務實現
type PriceServiceServer struct {
	pb.UnimplementedPriceServiceServer
	priceService *service.PriceService
}

// NewPriceServiceServer 創建 gRPC 服務器
func NewPriceServiceServer(priceService *service.PriceService) *PriceServiceServer {
	return &PriceServiceServer{
		priceService: priceService,
	}
}

// GetCurrentPrice 獲取當前價格
func (s *PriceServiceServer) GetCurrentPrice(ctx context.Context, req *pb.GetPriceRequest) (*pb.PriceResponse, error) {
	// 驗證 symbol
	if req.Symbol == "" {
		return nil, fmt.Errorf("symbol 不能為空")
	}

	symbol := model.Symbol(req.Symbol)
	if !isValidSymbol(symbol) {
		return nil, fmt.Errorf("不支援的商品代碼: %s", req.Symbol)
	}

	// 調用 service 層獲取價格
	price, err := s.priceService.GetCurrentPrice(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("獲取價格失敗: %v", err)
	}

	// 轉換為 protobuf 響應
	return &pb.PriceResponse{
		Symbol:        string(price.Symbol),
		Price:         price.Price,
		Timestamp:     price.Timestamp.UnixMilli(),
		Change:        price.Change,
		ChangePercent: price.ChangePercent,
	}, nil
}

// GetCurrentPrices 獲取多個商品的當前價格
func (s *PriceServiceServer) GetCurrentPrices(ctx context.Context, req *pb.GetPricesRequest) (*pb.PricesResponse, error) {
	var symbols []model.Symbol
	for _, symbolStr := range req.Symbols {
		symbol := model.Symbol(symbolStr)
		if isValidSymbol(symbol) {
			symbols = append(symbols, symbol)
		}
	}

	// 如果沒有指定 symbols，返回所有商品
	if len(symbols) == 0 {
		symbols = model.AllSymbols
	}

	// 調用 service 層獲取價格
	prices, err := s.priceService.GetCurrentPrices(ctx, symbols)
	if err != nil {
		return nil, fmt.Errorf("獲取價格失敗: %v", err)
	}

	// 轉換為 protobuf 響應
	var responses []*pb.PriceResponse
	for _, price := range prices {
		responses = append(responses, &pb.PriceResponse{
			Symbol:        string(price.Symbol),
			Price:         price.Price,
			Timestamp:     price.Timestamp.UnixMilli(),
			Change:        price.Change,
			ChangePercent: price.ChangePercent,
		})
	}

	return &pb.PricesResponse{
		Prices: responses,
	}, nil
}

// SubscribePrices 訂閱價格流（Server Streaming）
func (s *PriceServiceServer) SubscribePrices(req *pb.SubscribeRequest, stream pb.PriceService_SubscribePricesServer) error {
	// 處理訂閱的 symbols
	var symbols []model.Symbol
	if len(req.Symbols) > 0 {
		for _, symbolStr := range req.Symbols {
			symbol := model.Symbol(symbolStr)
			if isValidSymbol(symbol) {
				symbols = append(symbols, symbol)
			}
		}
	} else {
		// 如果沒有指定，訂閱所有商品
		symbols = model.AllSymbols
	}

	// 從 service 層訂閱價格
	priceChan := s.priceService.SubscribePrices(symbols)
	defer s.priceService.UnsubscribePrices(priceChan)

	// 循環接收價格更新並推送給客戶端
	ctx := stream.Context()
	for {
		select {
		case <-ctx.Done():
			log.Printf("客戶端斷開連接: %v", ctx.Err())
			return ctx.Err()
		case price := <-priceChan:
			if price == nil {
				continue
			}

			// 轉換為 protobuf 響應
			response := &pb.PriceUpdate{
				Symbol:        string(price.Symbol),
				Price:         price.Price,
				Timestamp:     price.Timestamp.UnixMilli(),
				Change:        price.Change,
				ChangePercent: price.ChangePercent,
			}

			// 推送給客戶端
			if err := stream.Send(response); err != nil {
				log.Printf("推送價格更新失敗: %v", err)
				return err
			}
		}
	}
}

// GetKlines 獲取 K 線資料
func (s *PriceServiceServer) GetKlines(ctx context.Context, req *pb.GetKlinesRequest) (*pb.KlinesResponse, error) {
	// 驗證參數
	if req.Symbol == "" {
		return nil, fmt.Errorf("symbol 不能為空")
	}

	symbol := model.Symbol(req.Symbol)
	if !isValidSymbol(symbol) {
		return nil, fmt.Errorf("不支援的商品代碼: %s", req.Symbol)
	}

	if req.Interval == "" {
		req.Interval = "1m" // 預設 1 分鐘
	}

	if !model.IsValidInterval(req.Interval) {
		return nil, fmt.Errorf("不支援的時間週期: %s", req.Interval)
	}

	// 設定預設值
	if req.Limit <= 0 {
		req.Limit = 100
	}
	if req.Limit > 1000 {
		req.Limit = 1000
	}

	if req.StartTime == 0 {
		req.StartTime = time.Now().Add(-24 * time.Hour).UnixMilli()
	}
	if req.EndTime == 0 {
		req.EndTime = time.Now().UnixMilli()
	}

	// 調用 service 層查詢 K 線
	klines, err := s.priceService.GetKlines(ctx, symbol, req.Interval, req.StartTime, req.EndTime, int(req.Limit))
	if err != nil {
		return nil, fmt.Errorf("查詢 K 線失敗: %v", err)
	}

	// 轉換為 protobuf 響應
	var responses []*pb.Kline
	for _, kline := range klines {
		responses = append(responses, &pb.Kline{
			Timestamp: kline.Timestamp.UnixMilli(),
			Open:      kline.Open,
			High:      kline.High,
			Low:       kline.Low,
			Close:     kline.Close,
			Volume:    kline.Volume,
		})
	}

	return &pb.KlinesResponse{
		Symbol:   string(symbol),
		Interval: req.Interval,
		Klines:   responses,
		Total:    int32(len(responses)),
	}, nil
}

// isValidSymbol 驗證商品代碼是否有效
func isValidSymbol(symbol model.Symbol) bool {
	for _, validSymbol := range model.AllSymbols {
		if symbol == validSymbol {
			return true
		}
	}
	return false
}
