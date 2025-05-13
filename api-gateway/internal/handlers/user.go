package handlers

import (
	"net/http"

	"promoservice/api-gateway/internal/client"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userClient  *client.UserClient
	eventClient *client.EventClient
}

func NewUserHandler(userClient *client.UserClient, eventClient *client.EventClient) *UserHandler {
	return &UserHandler{
		userClient:  userClient,
		eventClient: eventClient,
	}
}

func (h *UserHandler) Register(c *gin.Context) {
	var req struct {
		Login     string `json:"login" binding:"required"`
		Password  string `json:"password" binding:"required"`
		Email     string `json:"email" binding:"required,email"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Phone     string `json:"phone"`
		BirthDate string `json:"birth_date"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userClient.Register(c.Request.Context(), req.Login, req.Password, req.Email, req.FirstName, req.LastName, req.Phone, req.BirthDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Отправляем событие о регистрации
	err = h.eventClient.TrackClientRegistration(c.Request.Context(), user.Id, user.Email, user.FirstName+" "+user.LastName)
	if err != nil {
		// Логируем ошибку, но не прерываем регистрацию
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to track registration event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":  user,
		"token": user.Token,
	})
}

// ... rest of the file ...
