package service_test

import (
	"context"
	"testing"
	"time"

	"promoservice/user-service/internal/models"
	"promoservice/user-service/internal/service"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// MockUserRepository - мок репозитория для тестов
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	args := m.Called(ctx, login)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) UpdateUser(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func TestRegister(t *testing.T) {
	// Подготовка
	mockRepo := new(MockUserRepository)
	svc := service.NewUserService(mockRepo, "test-secret")
	ctx := context.Background()

	t.Run("successful registration", func(t *testing.T) {
		// Сбрасываем состояние мока
		mockRepo.ExpectedCalls = nil
		mockRepo.Calls = nil

		// Настройка мока
		mockRepo.On("GetUserByLogin", ctx, "testuser").Return(nil, nil)
		mockRepo.On("CreateUser", ctx, mock.AnythingOfType("*models.User")).Return(nil)

		// Выполнение
		user, token, err := svc.Register(ctx, "testuser", "password123", "test@example.com", "CUSTOMER")

		// Проверки
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.NotEmpty(t, token)
		assert.Equal(t, "testuser", user.Login)
		assert.Equal(t, "test@example.com", user.Email)
		assert.Equal(t, "CUSTOMER", user.Role)
		assert.NotEmpty(t, user.PasswordHash)
		assert.NotEqual(t, "password123", user.PasswordHash) // пароль должен быть захэширован
	})

	t.Run("registration with existing login should fail", func(t *testing.T) {
		// Сбрасываем состояние мока
		mockRepo.ExpectedCalls = nil
		mockRepo.Calls = nil

		// Настройка мока
		existingUser := &models.User{Login: "testuser"}
		mockRepo.On("GetUserByLogin", ctx, "testuser").Return(existingUser, nil)

		// Выполнение
		user, token, err := svc.Register(ctx, "testuser", "password123", "test@example.com", "CUSTOMER")

		// Проверки
		assert.Error(t, err)
		assert.Equal(t, "user already exists", err.Error())
		assert.Nil(t, user)
		assert.Empty(t, token)
	})
}

func TestLogin(t *testing.T) {
	// Подготовка
	mockRepo := new(MockUserRepository)
	svc := service.NewUserService(mockRepo, "test-secret")
	ctx := context.Background()

	// Создаем тестового пользователя с захэшированным паролем
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	testUser := &models.User{
		ID:           uuid.New(),
		Login:        "testuser",
		PasswordHash: string(hashedPassword),
		Email:        "test@example.com",
		Role:         "CUSTOMER",
	}

	t.Run("successful login by login", func(t *testing.T) {
		// Сбрасываем состояние мока
		mockRepo.ExpectedCalls = nil
		mockRepo.Calls = nil

		// Настройка мока
		mockRepo.On("GetUserByLogin", ctx, "testuser").Return(testUser, nil)

		// Выполнение
		user, token, err := svc.Login(ctx, "testuser", "password123", "")

		// Проверки
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.NotEmpty(t, token)
		assert.Equal(t, testUser.ID, user.ID)
		assert.Equal(t, testUser.Login, user.Login)
		assert.Equal(t, testUser.Email, user.Email)
	})

	t.Run("successful login by email", func(t *testing.T) {
		// Сбрасываем состояние мока
		mockRepo.ExpectedCalls = nil
		mockRepo.Calls = nil

		// Настройка мока
		mockRepo.On("GetUserByEmail", ctx, "test@example.com").Return(testUser, nil)

		// Выполнение
		user, token, err := svc.Login(ctx, "", "password123", "test@example.com")

		// Проверки
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.NotEmpty(t, token)
		assert.Equal(t, testUser.ID, user.ID)
		assert.Equal(t, testUser.Login, user.Login)
		assert.Equal(t, testUser.Email, user.Email)
	})

	t.Run("login with invalid password should fail", func(t *testing.T) {
		// Сбрасываем состояние мока
		mockRepo.ExpectedCalls = nil
		mockRepo.Calls = nil

		// Настройка мока
		mockRepo.On("GetUserByLogin", ctx, "testuser").Return(testUser, nil)

		// Выполнение
		user, token, err := svc.Login(ctx, "testuser", "wrongpassword", "")

		// Проверки
		assert.Error(t, err)
		assert.Equal(t, "invalid password", err.Error())
		assert.Nil(t, user)
		assert.Empty(t, token)
	})
}

func TestUpdateProfile(t *testing.T) {
	// Подготовка
	mockRepo := new(MockUserRepository)
	svc := service.NewUserService(mockRepo, "test-secret")
	ctx := context.Background()
	userID := uuid.New()

	existingUser := &models.User{
		ID:        userID,
		Login:     "testuser",
		Email:     "old@example.com",
		FirstName: "Old",
		LastName:  "Name",
		Phone:     "1234567890",
		Role:      "CUSTOMER",
	}

	t.Run("successful profile update", func(t *testing.T) {
		// Сбрасываем состояние мока
		mockRepo.ExpectedCalls = nil
		mockRepo.Calls = nil

		// Настройка моков
		mockRepo.On("GetUserByID", ctx, userID.String()).Return(existingUser, nil)
		mockRepo.On("UpdateUser", ctx, mock.AnythingOfType("*models.User")).Return(nil)

		// Выполнение
		updatedUser, err := svc.UpdateProfile(
			ctx,
			userID.String(),
			"new@example.com",
			"New",
			"Name",
			"9876543210",
			time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		)

		// Проверки
		assert.NoError(t, err)
		assert.NotNil(t, updatedUser)
		assert.Equal(t, "new@example.com", updatedUser.Email)
		assert.Equal(t, "New", updatedUser.FirstName)
		assert.Equal(t, "Name", updatedUser.LastName)
		assert.Equal(t, "9876543210", updatedUser.Phone)
		assert.Equal(t, time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC), updatedUser.BirthDate)
	})

	t.Run("update non-existent user should fail", func(t *testing.T) {
		// Сбрасываем состояние мока
		mockRepo.ExpectedCalls = nil
		mockRepo.Calls = nil

		// Настройка мока
		mockRepo.On("GetUserByID", ctx, userID.String()).Return(nil, nil)

		// Выполнение
		updatedUser, err := svc.UpdateProfile(
			ctx,
			userID.String(),
			"new@example.com",
			"New",
			"Name",
			"9876543210",
			time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		)

		// Проверки
		assert.Error(t, err)
		assert.Equal(t, "user not found", err.Error())
		assert.Nil(t, updatedUser)
	})
}
