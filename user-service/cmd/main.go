package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"promoservice/user-service/internal/config"
	"promoservice/user-service/internal/repository"
	"promoservice/user-service/internal/server"
	"promoservice/user-service/internal/service"

	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	db, err := repository.NewPostgresDB(cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	userRepo := repository.NewPostgresUserRepository(db)
	userService := service.NewUserService(userRepo, cfg.JWTSecret)

	grpcServer := server.NewServer(userService, cfg)

	go func() {
		if err := grpcServer.Start(); err != nil {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	grpcServer.Stop()
	<-ctx.Done()
	logger.Info("Server stopped")
}
