package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type JWTConfig struct {
	SecretKey string
}

func NewJWTConfig(secretKey string) *JWTConfig {
	return &JWTConfig{
		SecretKey: secretKey,
	}
}

func (c *JWTConfig) JWTAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			// Пропускаем запросы на регистрацию и вход без авторизации
			path := ctx.Request().URL.Path
			if path == "/users/register" || path == "/users/login" {
				return next(ctx)
			}

			authHeader := ctx.Request().Header.Get("Authorization")
			if authHeader == "" {
				return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Authorization header is required"})
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid authorization header format"})
			}

			tokenString := parts[1]
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(c.SecretKey), nil
			})

			if err != nil {
				return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
			}

			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				userID, ok := claims["user_id"].(string)
				if !ok {
					return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token claims"})
				}

				ctx.Set("user_id", userID)
				return next(ctx)
			}

			return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
		}
	}
}
