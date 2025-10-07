package grpc

import (
	"context"
	"fmt"
	"time"

	"github.com/mike/golden-buy/platform/internal/config"
	"github.com/mike/golden-buy/platform/internal/model"
	pb "github.com/mike/golden-buy/platform/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// PriceClient Price Service 的 gRPC 客戶端
type PriceClient struct {
	conn   *grpc.ClientConn
	client pb.PriceServiceClient
	cfg    *config.Config
}

// NewPriceClient 創建新的 Price Service 客戶端
func NewPriceClient(cfg *config.Config) (*PriceClient, error) {
	// 連接 Price Service
	conn, err := grpc.Dial(
		cfg.PriceServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(cfg.GRPCTimeout),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to price service at %s: %w", cfg.PriceServiceAddr, err)
	}

	client := pb.NewPriceServiceClient(conn)

	return &PriceClient{
		conn:   conn,
		client: client,
		cfg:    cfg,
	}, nil
}

// GetCurrentPrice 獲取單個商品當前價格
func (pc *PriceClient) GetCurrentPrice(ctx context.Context, symbol string) (*model.Price, error) {
	ctx, cancel := context.WithTimeout(ctx, pc.cfg.GRPCTimeout)
	defer cancel()

	resp, err := pc.client.GetCurrentPrice(ctx, &pb.GetPriceRequest{
		Symbol: symbol,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get current price for %s: %w", symbol, err)
	}

	return &model.Price{
		Symbol:        resp.Symbol,
		Price:         resp.Price,
		Timestamp:     resp.Timestamp,
		Change:        resp.Change,
		ChangePercent: resp.ChangePercent,
	}, nil
}

// GetCurrentPrices 獲取多個商品當前價格
func (pc *PriceClient) GetCurrentPrices(ctx context.Context, symbols []string) ([]*model.Price, error) {
	ctx, cancel := context.WithTimeout(ctx, pc.cfg.GRPCTimeout)
	defer cancel()

	resp, err := pc.client.GetCurrentPrices(ctx, &pb.GetPricesRequest{
		Symbols: symbols,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get current prices: %w", err)
	}

	prices := make([]*model.Price, len(resp.Prices))
	for i, p := range resp.Prices {
		prices[i] = &model.Price{
			Symbol:        p.Symbol,
			Price:         p.Price,
			Timestamp:     p.Timestamp,
			Change:        p.Change,
			ChangePercent: p.ChangePercent,
		}
	}

	return prices, nil
}

// GetKlines 獲取 K 線資料
func (pc *PriceClient) GetKlines(ctx context.Context, symbol, interval string, startTime, endTime int64, limit int32) ([]*model.Kline, error) {
	ctx, cancel := context.WithTimeout(ctx, pc.cfg.GRPCTimeout)
	defer cancel()

	resp, err := pc.client.GetKlines(ctx, &pb.GetKlinesRequest{
		Symbol:    symbol,
		Interval:  interval,
		StartTime: startTime,
		EndTime:   endTime,
		Limit:     limit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get klines for %s: %w", symbol, err)
	}

	klines := make([]*model.Kline, len(resp.Klines))
	for i, k := range resp.Klines {
		klines[i] = &model.Kline{
			Timestamp: k.Timestamp,
			Open:      k.Open,
			High:      k.High,
			Low:       k.Low,
			Close:     k.Close,
			Volume:    k.Volume,
		}
	}

	return klines, nil
}

// SubscribePrices 訂閱價格流（Server Streaming）
func (pc *PriceClient) SubscribePrices(ctx context.Context, symbols []string, callback func(*model.Price)) error {
	stream, err := pc.client.SubscribePrices(ctx, &pb.SubscribeRequest{
		Symbols: symbols,
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe prices: %w", err)
	}

	for {
		update, err := stream.Recv()
		if err != nil {
			return fmt.Errorf("stream receive error: %w", err)
		}

		price := &model.Price{
			Symbol:        update.Symbol,
			Price:         update.Price,
			Timestamp:     update.Timestamp,
			Change:        update.Change,
			ChangePercent: update.ChangePercent,
		}

		callback(price)
	}
}

// Close 關閉連接
func (pc *PriceClient) Close() error {
	if pc.conn != nil {
		return pc.conn.Close()
	}
	return nil
}

// Ping 檢查 Price Service 連接是否正常
func (pc *PriceClient) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// 嘗試獲取一個價格來確認服務是否正常
	_, err := pc.client.GetCurrentPrice(ctx, &pb.GetPriceRequest{
		Symbol: "GOLD",
	})
	if err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	return nil
}
