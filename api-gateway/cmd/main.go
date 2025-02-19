package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"promoservice/api-gateway/internal/api"
	"promoservice/api-gateway/internal/client"
	"promoservice/api-gateway/internal/config"
	authmiddleware "promoservice/api-gateway/internal/middleware"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

type Server struct {
	userClient *client.UserClient
	cfg        *config.Config
}

func NewServer(cfg *config.Config) (*Server, error) {
	userClient, err := client.NewUserClient(cfg.UserServiceAddress)
	if err != nil {
		return nil, err
	}
	return &Server{
		userClient: userClient,
		cfg:        cfg,
	}, nil
}

func (s *Server) PostUsersRegister(c echo.Context) error {
	var req api.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	user, token, err := s.userClient.Register(c.Request().Context(), req.Login, req.Password, string(req.Email))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"user":  user,
		"token": token,
	})
}

func (s *Server) PostUsersLogin(c echo.Context) error {
	var req api.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if req.Login == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Login is required"})
	}

	user, token, err := s.userClient.Login(c.Request().Context(), req.Login, req.Password, "")
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "not found"):
			return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
		case strings.Contains(err.Error(), "invalid password"):
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid password"})
		default:
			log.Printf("Login error: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"user":  user,
		"token": token,
	})
}

func (s *Server) GetUsersProfile(c echo.Context) error {
	token := c.Request().Header.Get("Authorization")
	if token == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}

	user, err := s.userClient.GetProfile(c.Request().Context(), token)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, user)
}

func (s *Server) PutUsersProfile(c echo.Context) error {
	token := c.Request().Header.Get("Authorization")
	if token == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}

	var req api.UpdateProfileRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	var email, firstName, lastName, phone string
	if req.Email != nil {
		email = string(*req.Email)
	}
	if req.FirstName != nil {
		firstName = *req.FirstName
	}
	if req.LastName != nil {
		lastName = *req.LastName
	}
	if req.Phone != nil {
		phone = *req.Phone
	}

	user, err := s.userClient.UpdateProfile(c.Request().Context(), token, email, firstName, lastName, phone, req.BirthDate)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, user)
}

func main() {
	cfg := config.NewConfig()

	server, err := NewServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	e := echo.New()

	// Middleware
	e.Use(echomiddleware.Logger())
	e.Use(echomiddleware.Recover())

	// JWT middleware
	jwtConfig := authmiddleware.NewJWTConfig(cfg.JWTSecret)
	e.Use(jwtConfig.JWTAuth())

	api.RegisterHandlers(e, server)

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
