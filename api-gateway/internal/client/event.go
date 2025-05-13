package client

import (
	"context"
	"time"

	"promoservice/proto/event"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type EventClient struct {
	client event.EventServiceClient
}

func NewEventClient(address string) (*EventClient, error) {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := event.NewEventServiceClient(conn)
	return &EventClient{client: client}, nil
}

func (c *EventClient) TrackPromoView(ctx context.Context, clientID, promoID string) error {
	_, err := c.client.TrackPromoView(ctx, &event.TrackPromoViewRequest{
		ClientId: clientID,
		PromoId:  promoID,
		ViewedAt: timestamppb.New(time.Now()),
	})
	return err
}

func (c *EventClient) TrackPromoClick(ctx context.Context, clientID, promoID string) error {
	_, err := c.client.TrackPromoClick(ctx, &event.TrackPromoClickRequest{
		ClientId:  clientID,
		PromoId:   promoID,
		ClickedAt: timestamppb.New(time.Now()),
	})
	return err
}

func (c *EventClient) TrackPromoComment(ctx context.Context, clientID, promoID, commentText string) error {
	_, err := c.client.TrackPromoComment(ctx, &event.TrackPromoCommentRequest{
		ClientId:    clientID,
		PromoId:     promoID,
		CommentText: commentText,
		CommentedAt: timestamppb.New(time.Now()),
	})
	return err
}

func (c *EventClient) GetPromoComments(ctx context.Context, promoID string, page, pageSize int32) (*event.GetPromoCommentsResponse, error) {
	return c.client.GetPromoComments(ctx, &event.GetPromoCommentsRequest{
		PromoId:  promoID,
		Page:     page,
		PageSize: pageSize,
	})
}

func (c *EventClient) TrackClientRegistration(ctx context.Context, clientID, email, name string) error {
	_, err := c.client.TrackClientRegistration(ctx, &event.TrackClientRegistrationRequest{
		ClientId:     clientID,
		Email:        email,
		Name:         name,
		RegisteredAt: timestamppb.New(time.Now()),
	})
	return err
}
