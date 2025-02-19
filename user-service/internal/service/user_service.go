package service

import (
	"context"
	"errors"
	"time"

	"promoservice/user-service/internal/models"
	"promoservice/user-service/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo      repository.UserRepository
	jwtSecret string
}

func NewUserService(repo repository.UserRepository, jwtSecret string) *UserService {
	return &UserService{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

func (s *UserService) Register(ctx context.Context, login, password, email string) (*models.User, string, error) {
	existingUser, err := s.repo.GetUserByLogin(ctx, login)
	if err != nil {
		return nil, "", err
	}
	if existingUser != nil {
		return nil, "", errors.New("user already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	user := &models.User{
		ID:           uuid.New(),
		Login:        login,
		PasswordHash: string(hashedPassword),
		Email:        email,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, "", err
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (s *UserService) Login(ctx context.Context, login, password, email string) (*models.User, string, error) {
	var user *models.User
	var err error

	if login != "" {
		user, err = s.repo.GetUserByLogin(ctx, login)
		if err != nil {
			return nil, "", err
		}
		if user == nil {
			return nil, "", errors.New("user not found")
		}
	} else if email != "" {
		user, err = s.repo.GetUserByEmail(ctx, email)
		if err != nil {
			return nil, "", err
		}
		if user == nil {
			return nil, "", errors.New("user not found")
		}
	} else {
		return nil, "", errors.New("either login or email is required")
	}

	// Проверяем пароль
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", errors.New("invalid password")
	}

	// Генерируем токен
	token, err := s.generateToken(user.ID)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (s *UserService) GetProfile(ctx context.Context, userID string) (*models.User, error) {
	return s.repo.GetUserByID(ctx, userID)
}

func (s *UserService) UpdateProfile(ctx context.Context, userID string, email, firstName, lastName, phone string, birthDate time.Time) (*models.User, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Обновляем поля
	user.Email = email
	user.FirstName = firstName
	user.LastName = lastName
	user.Phone = phone
	user.BirthDate = birthDate
	user.UpdatedAt = time.Now()

	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) generateToken(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}
