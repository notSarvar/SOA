package service

import (
	"context"
	"errors"
	"time"

	"promoservice/promo-service/internal/models"
	"promoservice/promo-service/internal/repository"
	"promoservice/proto/promo"

	"github.com/google/uuid"
)

type PromoService struct {
	repo        repository.PromoRepository
	commentRepo *repository.CommentRepository
}

func NewPromoService(repo repository.PromoRepository, commentRepo *repository.CommentRepository) *PromoService {
	return &PromoService{
		repo:        repo,
		commentRepo: commentRepo,
	}
}

func (s *PromoService) CreatePromo(ctx context.Context, req *promo.CreatePromoRequest, userID string, userRole promo.UserRole) (*models.Promo, error) {
	if userRole != promo.UserRole_BUSINESS {
		return nil, errors.New("only business users can create promos")
	}

	// Проверяем, не существует ли уже промокод с таким кодом
	existingPromo, err := s.repo.GetPromoByCode(ctx, req.Code)
	if err == nil && existingPromo != nil {
		return nil, errors.New("promo code already exists")
	}

	now := time.Now()
	promo := &models.Promo{
		ID:             uuid.New().String(),
		Name:           req.Name,
		Description:    req.Description,
		CreatorID:      userID,
		DiscountAmount: req.DiscountAmount,
		Code:           req.Code,
		CreatedAt:      now,
		UpdatedAt:      now,
		IsActive:       true,
		ValidUntil:     req.ValidUntil.AsTime(),
		MaxUses:        req.MaxUses,
		CurrentUses:    0,
	}

	if err := s.repo.CreatePromo(ctx, promo); err != nil {
		return nil, err
	}

	return promo, nil
}

func (s *PromoService) UpdatePromo(ctx context.Context, req *promo.UpdatePromoRequest, userID string, userRole promo.UserRole) (*models.Promo, error) {
	existingPromo, err := s.repo.GetPromoByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	// Проверяем права доступа
	if userRole != promo.UserRole_BUSINESS && existingPromo.CreatorID != userID {
		return nil, errors.New("unauthorized to update this promo")
	}

	// Если код изменился, проверяем его уникальность
	if req.Code != existingPromo.Code {
		codeExists, err := s.repo.GetPromoByCode(ctx, req.Code)
		if err == nil && codeExists != nil {
			return nil, errors.New("promo code already exists")
		}
	}

	existingPromo.Name = req.Name
	existingPromo.Description = req.Description
	existingPromo.DiscountAmount = req.DiscountAmount
	existingPromo.Code = req.Code
	existingPromo.UpdatedAt = time.Now()
	existingPromo.IsActive = req.IsActive
	existingPromo.ValidUntil = req.ValidUntil.AsTime()
	existingPromo.MaxUses = req.MaxUses

	if err := s.repo.UpdatePromo(ctx, existingPromo); err != nil {
		return nil, err
	}

	return existingPromo, nil
}

func (s *PromoService) DeletePromo(ctx context.Context, id string, userID string, userRole promo.UserRole) error {
	existingPromo, err := s.repo.GetPromoByID(ctx, id)
	if err != nil {
		return err
	}

	// Проверяем права доступа
	if userRole != promo.UserRole_BUSINESS && existingPromo.CreatorID != userID {
		return errors.New("unauthorized to delete this promo")
	}

	return s.repo.DeletePromo(ctx, id)
}

func (s *PromoService) GetPromo(ctx context.Context, id string) (*models.Promo, error) {
	return s.repo.GetPromoByID(ctx, id)
}

func (s *PromoService) ListPromos(ctx context.Context, creatorID string, page, pageSize int, onlyActive bool) ([]*models.Promo, int, error) {
	return s.repo.ListPromos(ctx, creatorID, page, pageSize, onlyActive)
}

func (s *PromoService) AddComment(ctx context.Context, promoID, clientID, commentText string) (*models.Comment, error) {
	// Проверяем существование промо
	promo, err := s.repo.GetPromoByID(ctx, promoID)
	if err != nil {
		return nil, err
	}
	if promo == nil {
		return nil, errors.New("promo not found")
	}

	// Парсим UUID
	promoUUID, err := uuid.Parse(promoID)
	if err != nil {
		return nil, errors.New("invalid promo ID")
	}

	clientUUID, err := uuid.Parse(clientID)
	if err != nil {
		return nil, errors.New("invalid client ID")
	}

	comment := &models.Comment{
		ID:          uuid.New(),
		PromoID:     promoUUID,
		ClientID:    clientUUID,
		CommentText: commentText,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	return s.commentRepo.Create(ctx, comment)
}

func (s *PromoService) GetComments(ctx context.Context, promoID string, page, pageSize int32) ([]*models.Comment, int32, error) {
	// Проверяем существование промо
	promo, err := s.repo.GetPromoByID(ctx, promoID)
	if err != nil {
		return nil, 0, err
	}
	if promo == nil {
		return nil, 0, errors.New("promo not found")
	}

	// Парсим UUID
	promoUUID, err := uuid.Parse(promoID)
	if err != nil {
		return nil, 0, errors.New("invalid promo ID")
	}

	comments, total, err := s.commentRepo.GetByPromoID(ctx, promoUUID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	return comments, int32(total), nil
}
