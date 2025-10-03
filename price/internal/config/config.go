package config

import (
	"log"
	"os"
	"time"
)

type Config struct {
	// InfluxDB 配置
	InfluxDB InfluxDBConfig

	// Redis 配置
	Redis RedisConfig

	// gRPC 配置
	GRPC GRPCConfig

	// 模擬器配置
	Simulator SimulatorConfig

	// 日誌配置
	LogLevel string
}

type InfluxDBConfig struct {
	URL    string
	Token  string
	Org    string
	Bucket string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type GRPCConfig struct {
	Port string
}

type SimulatorConfig struct {
	Interval   time.Duration
	Volatility float64
}

// Load 從環境變量載入配置
func Load() *Config {
	return &Config{
		InfluxDB: InfluxDBConfig{
			URL:    getEnv("INFLUXDB_URL", "http://localhost:8086"),
			Token:  getEnv("INFLUXDB_TOKEN", "my-super-secret-auth-token"),
			Org:    getEnv("INFLUXDB_ORG", "golden-buy"),
			Bucket: getEnv("INFLUXDB_BUCKET", "golden_buy"),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       0,
		},
		GRPC: GRPCConfig{
			Port: getEnv("GRPC_PORT", "50051"),
		},
		Simulator: SimulatorConfig{
			Interval:   parseDuration(getEnv("SIMULATOR_INTERVAL", "1s")),
			Volatility: 0.01, // 1% 波動率
		},
		LogLevel: getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		log.Printf("解析時間失敗，使用預設值: %v", err)
		return 1 * time.Second
	}
	return d
}
