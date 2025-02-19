package config

import (
	"os"
)

type Config struct {
	Port               string
	JWTSecret          string
	UserServiceAddress string
}

func NewConfig() *Config {
	return &Config{
		Port:               getEnv("PORT", "8080"),
		JWTSecret:          getEnv("JWT_SECRET", "your-secret-key"),
		UserServiceAddress: getEnv("USER_SERVICE_ADDRESS", "user-service:50051"),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
