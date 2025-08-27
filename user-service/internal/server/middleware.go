package server

import (
	"context"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	authorizationHeader = "authorization"
	bearerPrefix        = "Bearer "
)

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

type AuthMiddleware interface {
	UnaryServerInterceptor() grpc.UnaryServerInterceptor
}

type authMiddleware struct {
	jwtSecret string
}

func NewAuthMiddleware(jwtSecret string) AuthMiddleware {
	return &authMiddleware{
		jwtSecret: jwtSecret,
	}
}

func (m *authMiddleware) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if info.FullMethod == "/user.UserService/Register" || info.FullMethod == "/user.UserService/Login" {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
		}

		authHeader := md.Get(authorizationHeader)
		if len(authHeader) == 0 {
			return nil, status.Errorf(codes.Unauthenticated, "authorization token is not provided")
		}

		tokenString := authHeader[0]
		if !strings.HasPrefix(tokenString, bearerPrefix) {
			return nil, status.Errorf(codes.Unauthenticated, "invalid token format")
		}

		tokenString = strings.TrimPrefix(tokenString, bearerPrefix)

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(m.jwtSecret), nil
		})

		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
		}

		if !token.Valid {
			return nil, status.Errorf(codes.Unauthenticated, "invalid token")
		}

		// Добавляем user_id в контекст
		ctx = context.WithValue(ctx, "user_id", claims.UserID)

		return handler(ctx, req)
	}
}
