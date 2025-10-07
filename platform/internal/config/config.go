package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config 平台服務配置
type Config struct {
	// gRPC 客戶端配置
	PriceServiceAddr string
	GRPCTimeout      time.Duration

	// Redis 配置
	RedisAddr     string
	RedisPassword string
	RedisDB       int

	// 價格策略配置
	PriceStrategy string // "best" 或 "worst"，用於選擇每秒內的最佳或最差價格

	// WebSocket 配置（未來使用）
	HTTPPort string
	WSPath   string

	// 日誌配置
	LogLevel string
}

// Load 從環境變數載入配置
func Load() (*Config, error) {
	cfg := &Config{
		// gRPC 預設值
		PriceServiceAddr: getEnv("PRICE_SERVICE_ADDR", "localhost:50051"),
		GRPCTimeout:      getDurationEnv("GRPC_TIMEOUT", 10*time.Second),

		// Redis 預設值
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getIntEnv("REDIS_DB", 0),

		// 價格策略（預設最佳價格）
		PriceStrategy: getEnv("PRICE_STRATEGY", "best"),

		// HTTP/WebSocket 預設值
		HTTPPort: getEnv("HTTP_PORT", "8080"),
		WSPath:   getEnv("WS_PATH", "/ws/prices"),

		// 日誌配置
		LogLevel: getEnv("LOG_LEVEL", "info"),
	}

	// 驗證配置
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

// validate 驗證配置
func (c *Config) validate() error {
	if c.PriceServiceAddr == "" {
		return fmt.Errorf("PRICE_SERVICE_ADDR is required")
	}

	if c.RedisAddr == "" {
		return fmt.Errorf("REDIS_ADDR is required")
	}

	if c.PriceStrategy != "best" && c.PriceStrategy != "worst" {
		return fmt.Errorf("PRICE_STRATEGY must be 'best' or 'worst', got: %s", c.PriceStrategy)
	}

	return nil
}

// getEnv 獲取環境變數，若不存在則返回預設值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getIntEnv 獲取整數類型環境變數
func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getDurationEnv 獲取時間類型環境變數
func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
