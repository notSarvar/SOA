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
	repo repository.PromoRepository
}

func NewPromoService(repo repository.PromoRepository) *PromoService {
	return &PromoService{
		repo: repo,
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
