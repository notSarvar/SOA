package client

import (
	"context"
	"promoservice/proto/user"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserClient struct {
	client user.UserServiceClient
}

func NewUserClient(address string) (*UserClient, error) {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := user.NewUserServiceClient(conn)
	return &UserClient{
		client: client,
	}, nil
}

func (c *UserClient) Register(ctx context.Context, login, password, email string) (*user.User, string, error) {
	resp, err := c.client.Register(ctx, &user.RegisterRequest{
		Login:    login,
		Password: password,
		Email:    email,
	})
	if err != nil {
		return nil, "", err
	}
	return resp.User, resp.Token, nil
}

func (c *UserClient) Login(ctx context.Context, login, email, password string) (*user.User, string, error) {
	resp, err := c.client.Login(ctx, &user.LoginRequest{
		Login:    &login,
		Email:    &email,
		Password: password,
	})
	if err != nil {
		return nil, "", err
	}
	return resp.User, resp.Token, nil
}

func (c *UserClient) GetProfile(ctx context.Context, token string) (*user.User, error) {
	token = strings.TrimPrefix(token, "Bearer ")
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
	resp, err := c.client.GetProfile(ctx, &user.GetProfileRequest{})
	if err != nil {
		return nil, err
	}
	return resp.User, nil
}

func (c *UserClient) UpdateProfile(ctx context.Context, token string, email, firstName, lastName, phone string, birthDate *time.Time) (*user.User, error) {
	token = strings.TrimPrefix(token, "Bearer ")
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
	var birthDateProto *timestamppb.Timestamp
	if birthDate != nil {
		birthDateProto = timestamppb.New(*birthDate)
	}

	resp, err := c.client.UpdateProfile(ctx, &user.UpdateProfileRequest{
		Email:     &email,
		FirstName: &firstName,
		LastName:  &lastName,
		Phone:     &phone,
		BirthDate: birthDateProto,
	})
	if err != nil {
		return nil, err
	}
	return resp.User, nil
}
