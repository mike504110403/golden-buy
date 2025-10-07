package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golden-buy/price/internal/config"
	grpcServer "golden-buy/price/internal/grpc"
	"golden-buy/price/internal/pubsub"
	"golden-buy/price/internal/repository"
	"golden-buy/price/internal/service"
	"golden-buy/price/internal/simulator"
	pb "golden-buy/price/proto"

	"google.golang.org/grpc"
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

	log.Println("Price Service 啟動中...")

	// 1. 載入配置
	cfg := config.Load()
	log.Printf("配置載入完成: %+v", cfg)

	// 2. 連接 InfluxDB
	influxRepo, err := repository.NewInfluxDBRepository(
		cfg.InfluxDB.URL,
		cfg.InfluxDB.Token,
		cfg.InfluxDB.Org,
		cfg.InfluxDB.Bucket,
	)
	if err != nil {
		log.Fatalf("連接 InfluxDB 失敗: %v", err)
	}
	defer influxRepo.Close()
	log.Println("InfluxDB 連接成功")

	// 3. 連接 Redis
	redisPublisher, err := pubsub.NewPublisher(
		cfg.Redis.Addr,
		cfg.Redis.Password,
		cfg.Redis.DB,
	)
	if err != nil {
		log.Fatalf("連接 Redis 失敗: %v", err)
	}
	defer redisPublisher.Close()
	log.Println("Redis 連接成功")

	// 4. 創建價格模擬器
	simulator := simulator.NewPriceSimulator(cfg.Simulator.Interval, cfg.Simulator.Volatility)
	log.Println("價格模擬器創建成功")

	// 5. 創建業務邏輯服務
	priceService := service.NewPriceService(simulator, influxRepo, redisPublisher)
	log.Println("業務邏輯服務創建成功")

	// 6. 啟動價格模擬器（goroutine）
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go simulator.Start(ctx)
	log.Println("價格模擬器已啟動")

	// 啟動價格處理服務
	go priceService.Start(ctx)
	log.Println("價格處理服務已啟動")

	// 7. 啟動 gRPC 服務器
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPC.Port))
	if err != nil {
		log.Fatalf("啟動 gRPC 監聽失敗: %v", err)
	}

	server := grpc.NewServer()
	priceServer := grpcServer.NewPriceServiceServer(priceService)
	pb.RegisterPriceServiceServer(server, priceServer)

	log.Printf("gRPC 服務器啟動在端口 %s", cfg.GRPC.Port)

	// 在 goroutine 中啟動 gRPC 服務器
	go func() {
		if err := server.Serve(lis); err != nil {
			log.Fatalf("gRPC 服務器啟動失敗: %v", err)
		}
	}()

	// 8. 等待關閉信號
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	log.Println("收到關閉信號，優雅關閉中...")

	// 9. 清理資源
	cancel() // 停止所有 goroutine
	server.GracefulStop()
	log.Println("gRPC 服務器已停止")

	log.Println("Price Service 已停止")
}
