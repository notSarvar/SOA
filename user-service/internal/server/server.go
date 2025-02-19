package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"promoservice/proto/user"
	"promoservice/user-service/internal/config"
	"promoservice/user-service/internal/service"

	"google.golang.org/grpc"
)

type Server struct {
	user.UnimplementedUserServiceServer
	userService *service.UserService
	grpcServer  *grpc.Server
	cfg         *config.Config
}

func NewServer(userService *service.UserService, cfg *config.Config) *Server {
	authMiddleware := NewAuthMiddleware(cfg.JWTSecret)

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(authMiddleware.UnaryServerInterceptor()),
	)

	s := &Server{
		grpcServer:  grpcServer,
		userService: userService,
		cfg:         cfg,
	}

	user.RegisterUserServiceServer(grpcServer, s)
	return s
}

func (s *Server) Start() error {
	lis, err := net.Listen("tcp", ":"+s.cfg.GRPCPort)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	log.Println("Starting gRPC server on ", s.cfg.GRPCPort)

	if err := s.grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}

	return nil
}

func (s *Server) Stop() {
	log.Println("Shutting down gRPC server")
	s.grpcServer.GracefulStop()
}

func (s *Server) Register(ctx context.Context, req *user.RegisterRequest) (*user.RegisterResponse, error) {
	log.Println("Registering new user with login and email", req.Login, req.Email)

	u, token, err := s.userService.Register(ctx, req.Login, req.Password, req.Email)
	if err != nil {
		log.Printf("Something went wrong while registering")
		return nil, err
	}

	return &user.RegisterResponse{
		User:  u.ToProto(),
		Token: token,
	}, nil
}

func (s *Server) Login(ctx context.Context, req *user.LoginRequest) (*user.LoginResponse, error) {
	u, token, err := s.userService.Login(ctx, req.GetLogin(), req.GetEmail(), req.Password)
	if err != nil {
		log.Printf("Something went wrong while logging in")
		return nil, err
	}

	return &user.LoginResponse{
		User:  u.ToProto(),
		Token: token,
	}, nil
}

func (s *Server) GetProfile(ctx context.Context, req *user.GetProfileRequest) (*user.GetProfileResponse, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, fmt.Errorf("unauthorized: invalid or missing user_id")
	}
	log.Printf("Getting user profile with ID: %s", userID)

	u, err := s.userService.GetProfile(ctx, userID)
	if err != nil {
		log.Printf("Error getting profile for user %s: %v", userID, err)
		return nil, err
	}

	return &user.GetProfileResponse{
		User: u.ToProto(),
	}, nil
}

func (s *Server) UpdateProfile(ctx context.Context, req *user.UpdateProfileRequest) (*user.UpdateProfileResponse, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, fmt.Errorf("unauthorized: invalid or missing user_id")
	}
	log.Printf("Updating profile for user ID: %s", userID)

	// Валидация входных данных
	if req.Email != nil && !isValidEmail(*req.Email) {
		return nil, fmt.Errorf("invalid email format")
	}
	if req.Phone != nil && !isValidPhone(*req.Phone) {
		return nil, fmt.Errorf("invalid phone format")
	}

	birthDate := time.Time{}
	if req.BirthDate != nil {
		birthDate = req.BirthDate.AsTime()
	}

	u, err := s.userService.UpdateProfile(ctx, userID, req.GetEmail(), req.GetFirstName(), req.GetLastName(), req.GetPhone(), birthDate)
	if err != nil {
		log.Printf("Error updating profile for user %s: %v", userID, err)
		return nil, err
	}

	return &user.UpdateProfileResponse{
		User: u.ToProto(),
	}, nil
}

// Вспомогательные функции для валидации
func isValidEmail(email string) bool {
	// Простая валидация email
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

func isValidPhone(phone string) bool {
	// Простая валидация телефона (только цифры и + в начале)
	if len(phone) < 10 {
		return false
	}
	if phone[0] == '+' {
		phone = phone[1:]
	}
	for _, c := range phone {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
