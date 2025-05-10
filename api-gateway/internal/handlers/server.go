package handlers

import (
	"context"
	"net/http"
	"time"

	"promoservice/proto/promo"
	"promoservice/proto/user"

	"github.com/labstack/echo/v4"
)

type Server struct {
	userClient  UserClient
	promoClient PromoClient
}

type UserClient interface {
	Register(ctx context.Context, login, password, email, role string) (*user.User, string, error)
	Login(ctx context.Context, login, password, email string) (*user.User, string, error)
	GetProfile(ctx context.Context, token string) (*user.User, error)
	UpdateProfile(ctx context.Context, token string, email, firstName, lastName, phone string, birthDate *time.Time) (*user.User, error)
}

type PromoClient interface {
	CreatePromo(ctx context.Context, req *promo.CreatePromoRequest, userID string, userRole promo.UserRole) (*promo.CreatePromoResponse, error)
	GetPromo(ctx context.Context, id string) (*promo.GetPromoResponse, error)
	ListPromos(ctx context.Context, creatorID string, page, pageSize int32, onlyActive bool) (*promo.ListPromosResponse, error)
	UpdatePromo(ctx context.Context, req *promo.UpdatePromoRequest, userID string, userRole promo.UserRole) (*promo.UpdatePromoResponse, error)
	DeletePromo(ctx context.Context, id string, userID string, userRole promo.UserRole) error
}

func NewServer(userClient UserClient, promoClient PromoClient) *Server {
	return &Server{
		userClient:  userClient,
		promoClient: promoClient,
	}
}

func (s *Server) PostUsersRegister(c echo.Context) error {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
		Email    string `json:"email"`
		Role     string `json:"role" validate:"required,oneof=user business admin"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	user, token, err := s.userClient.Register(c.Request().Context(), req.Login, req.Password, req.Email, req.Role)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"user":  user,
		"token": token,
	})
}

func (s *Server) PostUsersLogin(c echo.Context) error {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if req.Login == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Login is required"})
	}

	user, token, err := s.userClient.Login(c.Request().Context(), req.Login, req.Password, "")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"user":  user,
		"token": token,
	})
}

func (s *Server) GetUsersProfile(c echo.Context) error {
	token := c.Request().Header.Get("Authorization")
	if token == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}

	user, err := s.userClient.GetProfile(c.Request().Context(), token)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, user)
}

func (s *Server) PutUsersProfile(c echo.Context) error {
	token := c.Request().Header.Get("Authorization")
	if token == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}

	var req struct {
		Email     *string    `json:"email,omitempty"`
		FirstName *string    `json:"first_name,omitempty"`
		LastName  *string    `json:"last_name,omitempty"`
		Phone     *string    `json:"phone,omitempty"`
		BirthDate *time.Time `json:"birth_date,omitempty"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	var email, firstName, lastName, phone string
	if req.Email != nil {
		email = *req.Email
	}
	if req.FirstName != nil {
		firstName = *req.FirstName
	}
	if req.LastName != nil {
		lastName = *req.LastName
	}
	if req.Phone != nil {
		phone = *req.Phone
	}

	user, err := s.userClient.UpdateProfile(c.Request().Context(), token, email, firstName, lastName, phone, req.BirthDate)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, user)
}
