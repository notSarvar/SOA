package config

import (
	"os"
)

type Config struct {
	Port                string
	UserServiceAddress  string
	PromoServiceAddress string
	JWTSecret           string
}

func NewConfig() *Config {
	return &Config{
		Port:                getEnv("PORT", "8080"),
		UserServiceAddress:  getEnv("USER_SERVICE_ADDRESS", "localhost:50051"),
		PromoServiceAddress: getEnv("PROMO_SERVICE_ADDRESS", "localhost:50052"),
		JWTSecret:           getEnv("JWT_SECRET", "your-secret-key"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
