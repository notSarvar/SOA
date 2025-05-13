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
	"google.golang.org/protobuf/types/known/timestamppb"
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

func (s *Server) AddComment(ctx context.Context, req *promo.AddCommentRequest) (*promo.AddCommentResponse, error) {
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

	comment, err := s.service.AddComment(ctx, req.PromoId, userID, req.CommentText)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &promo.AddCommentResponse{
		Comment: &promo.PromoComment{
			Id:          comment.ID.String(),
			PromoId:     comment.PromoID.String(),
			ClientId:    comment.ClientID.String(),
			CommentText: comment.CommentText,
			CreatedAt:   timestamppb.New(comment.CreatedAt),
			UpdatedAt:   timestamppb.New(comment.UpdatedAt),
		},
	}, nil
}

func (s *Server) GetComments(ctx context.Context, req *promo.GetCommentsRequest) (*promo.GetCommentsResponse, error) {
	comments, total, err := s.service.GetComments(ctx, req.PromoId, req.Page, req.PageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	protoComments := make([]*promo.PromoComment, len(comments))
	for i, c := range comments {
		protoComments[i] = &promo.PromoComment{
			Id:          c.ID.String(),
			PromoId:     c.PromoID.String(),
			ClientId:    c.ClientID.String(),
			CommentText: c.CommentText,
			CreatedAt:   timestamppb.New(c.CreatedAt),
			UpdatedAt:   timestamppb.New(c.UpdatedAt),
		}
	}

	return &promo.GetCommentsResponse{
		Comments: protoComments,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}
