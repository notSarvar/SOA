package service_test

import (
	"context"
	"testing"
	"time"

	"promoservice/promo-service/internal/models"
	"promoservice/promo-service/internal/service"
	"promoservice/proto/promo"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// MockPromoRepository - мок репозитория для тестов
type MockPromoRepository struct {
	mock.Mock
}

func (m *MockPromoRepository) CreatePromo(ctx context.Context, promo *models.Promo) error {
	args := m.Called(ctx, promo)
	return args.Error(0)
}

func (m *MockPromoRepository) GetPromoByID(ctx context.Context, id string) (*models.Promo, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Promo), args.Error(1)
}

func (m *MockPromoRepository) GetPromoByCode(ctx context.Context, code string) (*models.Promo, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Promo), args.Error(1)
}

func (m *MockPromoRepository) UpdatePromo(ctx context.Context, promo *models.Promo) error {
	args := m.Called(ctx, promo)
	return args.Error(0)
}

func (m *MockPromoRepository) DeletePromo(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPromoRepository) ListPromos(ctx context.Context, creatorID string, page, pageSize int, onlyActive bool) ([]*models.Promo, int, error) {
	args := m.Called(ctx, creatorID, page, pageSize, onlyActive)
	return args.Get(0).([]*models.Promo), args.Int(1), args.Error(2)
}

func TestCreatePromo(t *testing.T) {
	// Подготовка
	mockRepo := new(MockPromoRepository)
	svc := service.NewPromoService(mockRepo)
	ctx := context.Background()
	userID := uuid.New().String()

	validUntil := time.Now().Add(24 * time.Hour)
	req := &promo.CreatePromoRequest{
		Name:           "Test Promo",
		Description:    "Test Description",
		DiscountAmount: 10.0,
		Code:           "TEST123",
		ValidUntil:     timestamppb.New(validUntil),
		MaxUses:        100,
	}

	t.Run("successful creation by business user", func(t *testing.T) {
		// Сбрасываем состояние мока
		mockRepo.ExpectedCalls = nil
		mockRepo.Calls = nil

		// Настройка мока
		mockRepo.On("GetPromoByCode", ctx, req.Code).Return(nil, nil)
		mockRepo.On("CreatePromo", ctx, mock.AnythingOfType("*models.Promo")).Return(nil)

		// Выполнение
		result, err := svc.CreatePromo(ctx, req, userID, promo.UserRole_BUSINESS)

		// Проверки
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, req.Name, result.Name)
		assert.Equal(t, req.Description, result.Description)
		assert.Equal(t, req.DiscountAmount, result.DiscountAmount)
		assert.Equal(t, req.Code, result.Code)
		assert.Equal(t, userID, result.CreatorID)
		assert.True(t, result.IsActive)
		assert.Equal(t, req.MaxUses, result.MaxUses)
		assert.Equal(t, int32(0), result.CurrentUses)
	})

	t.Run("creation by non-business user should fail", func(t *testing.T) {
		// Сбрасываем состояние мока
		mockRepo.ExpectedCalls = nil
		mockRepo.Calls = nil

		result, err := svc.CreatePromo(ctx, req, userID, promo.UserRole_CUSTOMER)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "only business users can create promos", err.Error())
	})

	t.Run("creation with existing code should fail", func(t *testing.T) {
		// Сбрасываем состояние мока
		mockRepo.ExpectedCalls = nil
		mockRepo.Calls = nil

		existingPromo := &models.Promo{Code: req.Code}
		mockRepo.On("GetPromoByCode", ctx, req.Code).Return(existingPromo, nil)

		result, err := svc.CreatePromo(ctx, req, userID, promo.UserRole_BUSINESS)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "promo code already exists", err.Error())
	})
}

func TestUpdatePromo(t *testing.T) {
	// Подготовка
	mockRepo := new(MockPromoRepository)
	svc := service.NewPromoService(mockRepo)
	ctx := context.Background()
	userID := uuid.New().String()
	promoID := uuid.New().String()

	existingPromo := &models.Promo{
		ID:             promoID,
		Name:           "Old Name",
		Description:    "Old Description",
		CreatorID:      userID,
		DiscountAmount: 5.0,
		Code:           "OLD123",
		IsActive:       true,
		MaxUses:        50,
		CurrentUses:    0,
	}

	validUntil := time.Now().Add(24 * time.Hour)
	req := &promo.UpdatePromoRequest{
		Id:             promoID,
		Name:           "New Name",
		Description:    "New Description",
		DiscountAmount: 15.0,
		Code:           "NEW123",
		ValidUntil:     timestamppb.New(validUntil),
		MaxUses:        200,
		IsActive:       true,
	}

	t.Run("successful update by creator", func(t *testing.T) {
		// Сбрасываем состояние мока
		mockRepo.ExpectedCalls = nil
		mockRepo.Calls = nil

		// Настройка моков
		mockRepo.On("GetPromoByID", ctx, promoID).Return(existingPromo, nil)
		mockRepo.On("GetPromoByCode", ctx, req.Code).Return(nil, nil)
		mockRepo.On("UpdatePromo", ctx, mock.AnythingOfType("*models.Promo")).Return(nil)

		// Выполнение
		result, err := svc.UpdatePromo(ctx, req, userID, promo.UserRole_CUSTOMER)

		// Проверки
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, req.Name, result.Name)
		assert.Equal(t, req.Description, result.Description)
		assert.Equal(t, req.DiscountAmount, result.DiscountAmount)
		assert.Equal(t, req.Code, result.Code)
		assert.Equal(t, req.MaxUses, result.MaxUses)
	})

	t.Run("update by non-creator should fail", func(t *testing.T) {
		// Сбрасываем состояние мока
		mockRepo.ExpectedCalls = nil
		mockRepo.Calls = nil

		otherUserID := uuid.New().String()
		mockRepo.On("GetPromoByID", ctx, promoID).Return(existingPromo, nil)

		result, err := svc.UpdatePromo(ctx, req, otherUserID, promo.UserRole_CUSTOMER)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "unauthorized to update this promo", err.Error())
	})
}

func TestDeletePromo(t *testing.T) {
	// Подготовка
	mockRepo := new(MockPromoRepository)
	svc := service.NewPromoService(mockRepo)
	ctx := context.Background()
	userID := uuid.New().String()
	promoID := uuid.New().String()

	existingPromo := &models.Promo{
		ID:        promoID,
		CreatorID: userID,
	}

	t.Run("successful deletion by creator", func(t *testing.T) {
		// Сбрасываем состояние мока
		mockRepo.ExpectedCalls = nil
		mockRepo.Calls = nil

		// Настройка моков
		mockRepo.On("GetPromoByID", ctx, promoID).Return(existingPromo, nil)
		mockRepo.On("DeletePromo", ctx, promoID).Return(nil)

		// Выполнение
		err := svc.DeletePromo(ctx, promoID, userID, promo.UserRole_CUSTOMER)

		// Проверки
		assert.NoError(t, err)
	})

	t.Run("deletion by non-creator should fail", func(t *testing.T) {
		// Сбрасываем состояние мока
		mockRepo.ExpectedCalls = nil
		mockRepo.Calls = nil

		otherUserID := uuid.New().String()
		mockRepo.On("GetPromoByID", ctx, promoID).Return(existingPromo, nil)

		err := svc.DeletePromo(ctx, promoID, otherUserID, promo.UserRole_CUSTOMER)
		assert.Error(t, err)
		assert.Equal(t, "unauthorized to delete this promo", err.Error())
	})
}
