package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"promoservice/promo-service/internal/config"
	"promoservice/promo-service/internal/repository"
	"promoservice/promo-service/internal/server"
	"promoservice/promo-service/internal/service"
)

func main() {
	// Загружаем конфигурацию
	cfg := config.NewConfig()

	// Подключаемся к базе данных
	db, err := repository.NewPostgresDB(
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
	)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Инициализируем репозиторий
	promoRepo := repository.NewPostgresPromoRepository(db)

	// Инициализируем сервис
	promoService := service.NewPromoService(promoRepo)

	// Инициализируем и запускаем gRPC сервер
	srv := server.NewServer(promoService)
	go func() {
		if err := srv.Start(cfg.GRPCPort); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Ожидаем сигнал для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Graceful shutdown
	srv.Stop()
	log.Println("Server stopped")
}
