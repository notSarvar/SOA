package server

import (
	"context"
	"fmt"
	"net"

	"promoservice/promo-service/internal/service"
	"promoservice/proto/promo"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Server struct {
	promo.UnimplementedPromoServiceServer
	service *service.PromoService
	grpc    *grpc.Server
}

func NewServer(service *service.PromoService) *Server {
	return &Server{
		service: service,
	}
}

func (s *Server) Start(port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	s.grpc = grpc.NewServer()
	promo.RegisterPromoServiceServer(s.grpc, s)

	return s.grpc.Serve(lis)
}

func (s *Server) Stop() {
	if s.grpc != nil {
		s.grpc.GracefulStop()
	}
}

func (s *Server) CreatePromo(ctx context.Context, req *promo.CreatePromoRequest) (*promo.CreatePromoResponse, error) {
	// Получаем метаданные из контекста
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "metadata is not provided")
	}

	// Получаем информацию о пользователе из метаданных
	userIDs := md.Get("user_id")
	if len(userIDs) == 0 {
		return nil, status.Error(codes.Unauthenticated, "user_id is not provided")
	}
	userID := userIDs[0]

	userRoles := md.Get("user_role")
	if len(userRoles) == 0 {
		return nil, status.Error(codes.Unauthenticated, "user_role is not provided")
	}
	userRole := promo.UserRole(promo.UserRole_value[userRoles[0]])

	promoModel, err := s.service.CreatePromo(ctx, req, userID, userRole)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &promo.CreatePromoResponse{
		Promo: promoModel.ToProto(),
	}, nil
}

func (s *Server) UpdatePromo(ctx context.Context, req *promo.UpdatePromoRequest) (*promo.UpdatePromoResponse, error) {
	// Получаем метаданные из контекста
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "metadata is not provided")
	}

	// Получаем информацию о пользователе из метаданных
	userIDs := md.Get("user_id")
	if len(userIDs) == 0 {
		return nil, status.Error(codes.Unauthenticated, "user_id is not provided")
	}
	userID := userIDs[0]

	userRoles := md.Get("user_role")
	if len(userRoles) == 0 {
		return nil, status.Error(codes.Unauthenticated, "user_role is not provided")
	}
	userRole := promo.UserRole(promo.UserRole_value[userRoles[0]])

	updatedPromo, err := s.service.UpdatePromo(ctx, req, userID, userRole)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &promo.UpdatePromoResponse{
		Promo: updatedPromo.ToProto(),
	}, nil
}

func (s *Server) DeletePromo(ctx context.Context, req *promo.DeletePromoRequest) (*promo.DeletePromoResponse, error) {
	// Получаем метаданные из контекста
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "metadata is not provided")
	}

	// Получаем информацию о пользователе из метаданных
	userIDs := md.Get("user_id")
	if len(userIDs) == 0 {
		return nil, status.Error(codes.Unauthenticated, "user_id is not provided")
	}
	userID := userIDs[0]

	userRoles := md.Get("user_role")
	if len(userRoles) == 0 {
		return nil, status.Error(codes.Unauthenticated, "user_role is not provided")
	}
	userRole := promo.UserRole(promo.UserRole_value[userRoles[0]])

	err := s.service.DeletePromo(ctx, req.Id, userID, userRole)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &promo.DeletePromoResponse{}, nil
}

func (s *Server) GetPromo(ctx context.Context, req *promo.GetPromoRequest) (*promo.GetPromoResponse, error) {
	promoModel, err := s.service.GetPromo(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &promo.GetPromoResponse{
		Promo: promoModel.ToProto(),
	}, nil
}

func (s *Server) ListPromos(ctx context.Context, req *promo.ListPromosRequest) (*promo.ListPromosResponse, error) {
	promos, total, err := s.service.ListPromos(ctx, req.CreatorId, int(req.Page), int(req.PageSize), req.OnlyActive)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	protoPromos := make([]*promo.Promo, len(promos))
	for i, p := range promos {
		protoPromos[i] = p.ToProto()
	}

	return &promo.ListPromosResponse{
		Promos: protoPromos,
		Total:  int32(total),
	}, nil
}
