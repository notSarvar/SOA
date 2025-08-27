package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"promoservice/api-gateway/internal/client"
	"promoservice/api-gateway/internal/config"
	"promoservice/api-gateway/internal/handlers"
	authmiddleware "promoservice/api-gateway/internal/middleware"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

func main() {
	cfg := config.NewConfig()

	// Инициализация клиентов
	userClient, err := client.NewUserClient(cfg.UserServiceAddress)
	if err != nil {
		log.Fatalf("Failed to create user client: %v", err)
	}

	promoClient, err := client.NewPromoClient(cfg.PromoServiceAddress)
	if err != nil {
		log.Fatalf("Failed to create promo client: %v", err)
	}

	// Инициализация обработчиков
	server := handlers.NewServer(userClient, promoClient)
	promoHandler := handlers.NewPromoHandler(promoClient)

	e := echo.New()

	// Middleware
	e.Use(echomiddleware.Logger())
	e.Use(echomiddleware.Recover())

	// API v1 группа
	v1 := e.Group("/api/v1")

	// Публичные маршруты
	auth := v1.Group("/users")
	auth.POST("/register", server.PostUsersRegister)
	auth.POST("/login", server.PostUsersLogin)

	// JWT middleware для защищенных маршрутов
	jwtConfig := authmiddleware.NewJWTConfig(cfg.JWTSecret)
	protected := v1.Group("")
	protected.Use(jwtConfig.JWTAuth())

	// Защищенные маршруты для пользователей
	users := protected.Group("/users")
	users.GET("/profile", server.GetUsersProfile)
	users.PUT("/profile", server.PutUsersProfile)

	// Защищенные маршруты для промокодов
	promos := protected.Group("/promos")
	promos.POST("", promoHandler.CreatePromo)
	promos.GET("", promoHandler.ListPromos)
	promos.GET("/:id", promoHandler.GetPromo)
	promos.PUT("/:id", promoHandler.UpdatePromo)
	promos.DELETE("/:id", promoHandler.DeletePromo)

	go func() {
		if err := e.Start(":" + cfg.Port); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
