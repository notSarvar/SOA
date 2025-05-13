package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"promoservice/event-service/internal/config"
	"promoservice/event-service/internal/database"
	"promoservice/event-service/internal/kafka"
	"promoservice/event-service/internal/service"
	"promoservice/proto/event"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func main() {
	// Инициализация логгера
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	// Загрузка конфигурации
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		logger.Fatalf("Failed to load config: %v", err)
	}

	// Инициализация базы данных
	db, err := database.NewPostgres(
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		logger,
	)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Инициализация Kafka producer
	producer, err := kafka.NewProducer(
		cfg.Kafka.Brokers,
		&cfg.Kafka.Producer,
		&cfg.Kafka.DeadLetter,
		logger,
	)
	if err != nil {
		logger.Fatalf("Failed to create Kafka producer: %v", err)
	}
	defer producer.Close()

	// Инициализация сервиса
	svcConfig := &service.Config{
		KafkaTopics: struct {
			PromoCreated       string
			PromoUpdated       string
			PromoDeleted       string
			PromoViewed        string
			PromoClicked       string
			PromoLiked         string
			UserRegistered     string
			UserProfileUpdated string
		}{
			PromoCreated:       cfg.Kafka.Topics.PromoCreated,
			PromoUpdated:       cfg.Kafka.Topics.PromoUpdated,
			PromoDeleted:       cfg.Kafka.Topics.PromoDeleted,
			PromoViewed:        cfg.Kafka.Topics.PromoViewed,
			PromoClicked:       cfg.Kafka.Topics.PromoClicked,
			PromoLiked:         cfg.Kafka.Topics.PromoLiked,
			UserRegistered:     cfg.Kafka.Topics.UserRegistered,
			UserProfileUpdated: cfg.Kafka.Topics.UserProfileUpdated,
		},
	}

	svc := service.NewService(db, producer, logger, svcConfig)

	// Создание gRPC сервера
	grpcServer := grpc.NewServer()

	// Регистрация сервиса
	event.RegisterEventServiceServer(grpcServer, svc)

	// Запуск сервера
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port))
	if err != nil {
		logger.Fatalf("Failed to listen: %v", err)
	}

	go func() {
		logger.Infof("Starting gRPC server on %s:%d", cfg.Server.Host, cfg.Server.Port)
		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Ожидание сигнала для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")
	grpcServer.GracefulStop()
	logger.Info("Server stopped")
}
