package http

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mike/golden-buy/platform/internal/service"
	"github.com/mike/golden-buy/platform/internal/websocket"
)

// Server HTTP 服務器
type Server struct {
	handler   *Handler
	wsHandler *websocket.Handler
	engine    *gin.Engine
	addr      string
}

// NewServer 創建新的 HTTP 服務器
func NewServer(addr string, svc *service.PlatformService, wsHub *websocket.Hub) *Server {
	// 設置 Gin 為 release 模式（生產環境）
	// gin.SetMode(gin.ReleaseMode)

	// 創建 Gin 引擎
	engine := gin.Default()

	// 添加 CORS 中間件
	engine.Use(corsMiddleware())

	// 創建處理器
	handler := NewHandler(svc)
	wsHandler := websocket.NewHandler(wsHub)

	// 健康檢查
	engine.GET("/health", handler.HandleHealthCheck)

	// API 路由組
	api := engine.Group("/api")
	{
		// 價格相關路由
		prices := api.Group("/prices")
		{
			prices.GET("/current", handler.HandleGetCurrentPrice)
			prices.GET("/history", handler.HandleGetHistory)
		}

		// 用戶相關路由
		user := api.Group("/user")
		{
			user.GET("/info", handler.HandleGetUserInfo)
		}
	}

	// WebSocket 路由
	engine.GET("/ws/prices", func(c *gin.Context) {
		wsHandler.ServeWS(c.Writer, c.Request)
	})

	return &Server{
		handler:   handler,
		wsHandler: wsHandler,
		engine:    engine,
		addr:      addr,
	}
}

// Start 啟動 HTTP 服務器
func (s *Server) Start() error {
	log.Printf("🌐 Starting HTTP server on %s", s.addr)
	log.Println("📍 Available endpoints:")
	log.Println("   GET  /health                  - Health check")
	log.Println("   GET  /api/prices/current      - Get current prices")
	log.Println("   GET  /api/prices/history      - Get historical klines")
	log.Println("   GET  /api/user/info           - Get user info (demo)")
	log.Println("   WS   /ws/prices               - WebSocket price stream")

	if err := s.engine.Run(s.addr); err != nil {
		return err
	}

	return nil
}

// Stop 停止 HTTP 服務器
func (s *Server) Stop() error {
	log.Println("🛑 Stopping HTTP server...")
	// Gin 沒有內建的優雅關閉，可以使用 http.Server 包裝
	return nil
}

// corsMiddleware CORS 中間件
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		// 處理預檢請求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// loggingMiddleware 自定義日誌中間件（可選，Gin 已有內建）
func loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		log.Printf("📥 %s %s from %s - %d (%v)",
			c.Request.Method,
			path,
			c.ClientIP(),
			c.Writer.Status(),
			time.Since(start),
		)
	}
}

