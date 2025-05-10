package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	DatabaseURL           string        `env:"DATABASE_URL"`
	JWTSecret             string        `env:"JWT_SECRET"`
	GRPCPort              string        `env:"GRPC_PORT"`
	TokenTTL              time.Duration `env:"TOKEN_TTL"`
	BCryptCost            int           `env:"BCRYPT_COST"`
	MaxRequestSize        int           `env:"MAX_REQUEST_SIZE"`
	RateLimit             int           `env:"RATE_LIMIT"`
	RateLimitWindow       int           `env:"RATE_LIMIT_WINDOW"`
	MaxConnectionIdle     int           `env:"MAX_CONNECTION_IDLE"`
	MaxConnectionAge      int           `env:"MAX_CONNECTION_AGE"`
	MaxConnectionAgeGrace int           `env:"MAX_CONNECTION_AGE_GRACE"`
	KeepAliveTime         int           `env:"KEEP_ALIVE_TIME"`
	KeepAliveTimeout      int           `env:"KEEP_ALIVE_TIMEOUT"`
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		JWTSecret:       os.Getenv("JWT_SECRET"),
		GRPCPort:        os.Getenv("GRPC_PORT"),
		MaxRequestSize:  1024 * 1024, // 1MB
		RateLimit:       100,
		RateLimitWindow: 60, // 60 seconds
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	if cfg.GRPCPort == "" {
		cfg.GRPCPort = "50051"
	}
	return cfg, nil
}
