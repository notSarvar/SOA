package service

import (
	"context"
	"time"

	"github.com/sarvar/PromoService/event-service/internal/database"
	"github.com/sarvar/PromoService/event-service/internal/kafka"
	"github.com/sarvar/PromoService/proto/event"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service struct {
	event.UnimplementedEventServiceServer
	db       *database.Postgres
	producer *kafka.Producer
	logger   *logrus.Logger
	config   *Config
}

type Config struct {
	KafkaTopics struct {
		PromoCreated       string
		PromoUpdated       string
		PromoDeleted       string
		PromoViewed        string
		PromoClicked       string
		PromoLiked         string
		UserRegistered     string
		UserProfileUpdated string
	}
}

func NewService(db *database.Postgres, producer *kafka.Producer, logger *logrus.Logger, config *Config) *Service {
	return &Service{
		db:       db,
		producer: producer,
		logger:   logger,
		config:   config,
	}
}

func (s *Service) TrackPromoView(ctx context.Context, req *event.TrackPromoViewRequest) (*event.TrackPromoViewResponse, error) {
	if err := s.db.SavePromoView(ctx, req.PromoId, req.ClientId); err != nil {
		s.logger.WithError(err).Error("Failed to save promo view")
		return nil, err
	}

	// Отправляем событие в Kafka
	viewEvent := struct {
		PromoID  string    `json:"promo_id"`
		UserID   string    `json:"user_id"`
		ViewedAt time.Time `json:"viewed_at"`
	}{
		PromoID:  req.PromoId,
		UserID:   req.ClientId,
		ViewedAt: req.ViewedAt.AsTime(),
	}

	if err := s.producer.SendMessage(ctx, s.config.KafkaTopics.PromoViewed, req.PromoId, viewEvent); err != nil {
		s.logger.WithError(err).Error("Failed to send promo view event to Kafka")
		// Не возвращаем ошибку, так как событие уже сохранено в БД
	}

	return &event.TrackPromoViewResponse{Success: true}, nil
}

func (s *Service) TrackPromoClick(ctx context.Context, req *event.TrackPromoClickRequest) (*event.TrackPromoClickResponse, error) {
	if err := s.db.SavePromoClick(ctx, req.PromoId, req.ClientId); err != nil {
		s.logger.WithError(err).Error("Failed to save promo click")
		return nil, err
	}

	// Отправляем событие в Kafka
	clickEvent := struct {
		PromoID   string    `json:"promo_id"`
		UserID    string    `json:"user_id"`
		ClickedAt time.Time `json:"clicked_at"`
	}{
		PromoID:   req.PromoId,
		UserID:    req.ClientId,
		ClickedAt: req.ClickedAt.AsTime(),
	}

	if err := s.producer.SendMessage(ctx, s.config.KafkaTopics.PromoClicked, req.PromoId, clickEvent); err != nil {
		s.logger.WithError(err).Error("Failed to send promo click event to Kafka")
		// Не возвращаем ошибку, так как событие уже сохранено в БД
	}

	return &event.TrackPromoClickResponse{Success: true}, nil
}

func (s *Service) TrackPromoComment(ctx context.Context, req *event.TrackPromoCommentRequest) (*event.TrackPromoCommentResponse, error) {
	if err := s.db.SavePromoComment(ctx, req.PromoId, req.ClientId, req.CommentText); err != nil {
		s.logger.WithError(err).Error("Failed to save promo comment")
		return nil, err
	}

	// Отправляем событие в Kafka
	commentEvent := struct {
		PromoID     string    `json:"promo_id"`
		UserID      string    `json:"user_id"`
		CommentText string    `json:"comment_text"`
		CommentedAt time.Time `json:"commented_at"`
	}{
		PromoID:     req.PromoId,
		UserID:      req.ClientId,
		CommentText: req.CommentText,
		CommentedAt: req.CommentedAt.AsTime(),
	}

	if err := s.producer.SendMessage(ctx, s.config.KafkaTopics.PromoCreated, req.PromoId, commentEvent); err != nil {
		s.logger.WithError(err).Error("Failed to send promo comment event to Kafka")
		// Не возвращаем ошибку, так как событие уже сохранено в БД
	}

	return &event.TrackPromoCommentResponse{Success: true}, nil
}

func (s *Service) GetPromoComments(ctx context.Context, req *event.GetPromoCommentsRequest) (*event.GetPromoCommentsResponse, error) {
	comments, total, err := s.db.GetPromoComments(ctx, req.PromoId, int(req.Page), int(req.PageSize))
	if err != nil {
		s.logger.WithError(err).Error("Failed to get promo comments")
		return nil, err
	}

	protoComments := make([]*event.PromoComment, len(comments))
	for i, comment := range comments {
		protoComments[i] = &event.PromoComment{
			Id:          comment.ID,
			ClientId:    comment.UserID,
			PromoId:     comment.PromoID,
			CommentText: comment.CommentText,
			CommentedAt: timestamppb.New(time.Now()), // TODO: Parse the actual timestamp
		}
	}

	return &event.GetPromoCommentsResponse{
		Comments:   protoComments,
		TotalCount: int32(total),
	}, nil
}

func (s *Service) TrackClientRegistration(ctx context.Context, req *event.TrackClientRegistrationRequest) (*event.TrackClientRegistrationResponse, error) {
	// Отправляем событие в Kafka
	registrationEvent := struct {
		ClientID     string    `json:"client_id"`
		Email        string    `json:"email"`
		Name         string    `json:"name"`
		RegisteredAt time.Time `json:"registered_at"`
	}{
		ClientID:     req.ClientId,
		Email:        req.Email,
		Name:         req.Name,
		RegisteredAt: req.RegisteredAt.AsTime(),
	}

	if err := s.producer.SendMessage(ctx, s.config.KafkaTopics.UserRegistered, req.ClientId, registrationEvent); err != nil {
		s.logger.WithError(err).Error("Failed to send client registration event to Kafka")
		return nil, err
	}

	return &event.TrackClientRegistrationResponse{Success: true}, nil
}

func (s *Service) TrackPromoCreated(ctx context.Context, req *event.TrackPromoCreatedRequest) (*event.TrackPromoCreatedResponse, error) {
	// Отправляем событие в Kafka
	createdEvent := struct {
		PromoID        string    `json:"promo_id"`
		CreatorID      string    `json:"creator_id"`
		Code           string    `json:"code"`
		DiscountAmount float64   `json:"discount_amount"`
		CreatedAt      time.Time `json:"created_at"`
	}{
		PromoID:        req.PromoId,
		CreatorID:      req.CreatorId,
		Code:           req.Code,
		DiscountAmount: req.DiscountAmount,
		CreatedAt:      req.CreatedAt.AsTime(),
	}

	if err := s.producer.SendMessage(ctx, s.config.KafkaTopics.PromoCreated, req.PromoId, createdEvent); err != nil {
		s.logger.WithError(err).Error("Failed to send promo created event to Kafka")
		return nil, err
	}

	return &event.TrackPromoCreatedResponse{Success: true}, nil
}

func (s *Service) TrackPromoUpdated(ctx context.Context, req *event.TrackPromoUpdatedRequest) (*event.TrackPromoUpdatedResponse, error) {
	// Отправляем событие в Kafka
	updatedEvent := struct {
		PromoID       string    `json:"promo_id"`
		UpdatedBy     string    `json:"updated_by"`
		UpdatedFields []string  `json:"updated_fields"`
		UpdatedAt     time.Time `json:"updated_at"`
	}{
		PromoID:       req.PromoId,
		UpdatedBy:     req.UpdatedBy,
		UpdatedFields: req.UpdatedFields,
		UpdatedAt:     req.UpdatedAt.AsTime(),
	}

	if err := s.producer.SendMessage(ctx, s.config.KafkaTopics.PromoUpdated, req.PromoId, updatedEvent); err != nil {
		s.logger.WithError(err).Error("Failed to send promo updated event to Kafka")
		return nil, err
	}

	return &event.TrackPromoUpdatedResponse{Success: true}, nil
}

func (s *Service) TrackPromoDeleted(ctx context.Context, req *event.TrackPromoDeletedRequest) (*event.TrackPromoDeletedResponse, error) {
	// Отправляем событие в Kafka
	deletedEvent := struct {
		PromoID   string    `json:"promo_id"`
		DeletedBy string    `json:"deleted_by"`
		DeletedAt time.Time `json:"deleted_at"`
	}{
		PromoID:   req.PromoId,
		DeletedBy: req.DeletedBy,
		DeletedAt: req.DeletedAt.AsTime(),
	}

	if err := s.producer.SendMessage(ctx, s.config.KafkaTopics.PromoDeleted, req.PromoId, deletedEvent); err != nil {
		s.logger.WithError(err).Error("Failed to send promo deleted event to Kafka")
		return nil, err
	}

	return &event.TrackPromoDeletedResponse{Success: true}, nil
}

func (s *Service) TrackUserProfileUpdate(ctx context.Context, req *event.TrackUserProfileUpdateRequest) (*event.TrackUserProfileUpdateResponse, error) {
	// Отправляем событие в Kafka
	updateEvent := struct {
		UserID        string    `json:"user_id"`
		UpdatedFields []string  `json:"updated_fields"`
		UpdatedAt     time.Time `json:"updated_at"`
	}{
		UserID:        req.UserId,
		UpdatedFields: req.UpdatedFields,
		UpdatedAt:     req.UpdatedAt.AsTime(),
	}

	if err := s.producer.SendMessage(ctx, s.config.KafkaTopics.UserProfileUpdated, req.UserId, updateEvent); err != nil {
		s.logger.WithError(err).Error("Failed to send user profile update event to Kafka")
		return nil, err
	}

	return &event.TrackUserProfileUpdateResponse{Success: true}, nil
}
