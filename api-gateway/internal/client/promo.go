package client

import (
	"context"
	"promoservice/proto/promo"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type PromoClient struct {
	client promo.PromoServiceClient
}

func NewPromoClient(address string) (*PromoClient, error) {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := promo.NewPromoServiceClient(conn)
	return &PromoClient{
		client: client,
	}, nil
}

func (c *PromoClient) CreatePromo(ctx context.Context, req *promo.CreatePromoRequest, userID string, userRole promo.UserRole) (*promo.CreatePromoResponse, error) {
	// Добавляем информацию о пользователе в метаданные gRPC
	md := metadata.New(map[string]string{
		"user_id":   userID,
		"user_role": userRole.String(),
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	return c.client.CreatePromo(ctx, req)
}

func (c *PromoClient) GetPromo(ctx context.Context, id string) (*promo.GetPromoResponse, error) {
	return c.client.GetPromo(ctx, &promo.GetPromoRequest{Id: id})
}

func (c *PromoClient) ListPromos(ctx context.Context, creatorID string, page, pageSize int32, onlyActive bool) (*promo.ListPromosResponse, error) {
	return c.client.ListPromos(ctx, &promo.ListPromosRequest{
		CreatorId:  creatorID,
		Page:       page,
		PageSize:   pageSize,
		OnlyActive: onlyActive,
	})
}

func (c *PromoClient) UpdatePromo(ctx context.Context, req *promo.UpdatePromoRequest, userID string, userRole promo.UserRole) (*promo.UpdatePromoResponse, error) {
	ctx = metadata.AppendToOutgoingContext(ctx, "user_id", userID, "user_role", userRole.String())
	return c.client.UpdatePromo(ctx, req)
}

func (c *PromoClient) DeletePromo(ctx context.Context, id string, userID string, userRole promo.UserRole) error {
	ctx = metadata.AppendToOutgoingContext(ctx, "user_id", userID, "user_role", userRole.String())
	_, err := c.client.DeletePromo(ctx, &promo.DeletePromoRequest{Id: id})
	return err
}
