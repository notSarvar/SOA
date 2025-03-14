package handlers

import (
	"net/http"
	"strconv"
	"time"

	"promoservice/api-gateway/internal/client"
	"promoservice/proto/promo"

	"github.com/labstack/echo/v4"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PromoHandler struct {
	promoClient *client.PromoClient
}

func NewPromoHandler(promoClient *client.PromoClient) *PromoHandler {
	return &PromoHandler{
		promoClient: promoClient,
	}
}

func (h *PromoHandler) CreatePromo(c echo.Context) error {
	var req struct {
		Code      string    `json:"code"`
		Discount  float64   `json:"discount"`
		ExpiresAt time.Time `json:"expires_at"`
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	userID := c.Get("user_id").(string)
	userRole := promo.UserRole_CUSTOMER // Default role if not set

	if role, ok := c.Get("user_role").(string); ok {
		switch role {
		case "business":
			userRole = promo.UserRole_BUSINESS
		case "user":
			userRole = promo.UserRole_CUSTOMER
		default:
			userRole = promo.UserRole_UNKNOWN
		}
	}

	protoReq := &promo.CreatePromoRequest{
		Code:           req.Code,
		DiscountAmount: req.Discount,
		ValidUntil:     timestamppb.New(req.ExpiresAt),
		MaxUses:        100, // Default value
	}

	resp, err := h.promoClient.CreatePromo(c.Request().Context(), protoReq, userID, userRole)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, resp)
}

func (h *PromoHandler) GetPromo(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "id is required")
	}

	resp, err := h.promoClient.GetPromo(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *PromoHandler) ListPromos(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(c.QueryParam("page_size"))
	if pageSize < 1 {
		pageSize = 10
	}

	creatorID := c.QueryParam("creator_id")
	onlyActive := c.QueryParam("only_active") == "true"

	resp, err := h.promoClient.ListPromos(c.Request().Context(), creatorID, int32(page), int32(pageSize), onlyActive)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *PromoHandler) UpdatePromo(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "id is required")
	}

	var req struct {
		Name           string    `json:"name"`
		Description    string    `json:"description"`
		DiscountAmount float64   `json:"discount_amount"`
		Code           string    `json:"code"`
		ValidUntil     time.Time `json:"valid_until"`
		MaxUses        int32     `json:"max_uses"`
		IsActive       bool      `json:"is_active"`
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	userID := c.Get("user_id").(string)
	userRole := promo.UserRole_CUSTOMER // Default role if not set

	if role, ok := c.Get("user_role").(string); ok {
		switch role {
		case "business":
			userRole = promo.UserRole_BUSINESS
		case "user":
			userRole = promo.UserRole_CUSTOMER
		default:
			userRole = promo.UserRole_UNKNOWN
		}
	}

	protoReq := &promo.UpdatePromoRequest{
		Id:             id,
		Name:           req.Name,
		Description:    req.Description,
		DiscountAmount: req.DiscountAmount,
		Code:           req.Code,
		ValidUntil:     timestamppb.New(req.ValidUntil),
		MaxUses:        req.MaxUses,
		IsActive:       req.IsActive,
	}

	resp, err := h.promoClient.UpdatePromo(c.Request().Context(), protoReq, userID, userRole)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *PromoHandler) DeletePromo(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "id is required")
	}

	userID := c.Get("user_id").(string)
	userRole := promo.UserRole_CUSTOMER // Default role if not set

	if role, ok := c.Get("user_role").(string); ok {
		switch role {
		case "business":
			userRole = promo.UserRole_BUSINESS
		case "user":
			userRole = promo.UserRole_CUSTOMER
		default:
			userRole = promo.UserRole_UNKNOWN
		}
	}

	err := h.promoClient.DeletePromo(c.Request().Context(), id, userID, userRole)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}

func (s *Server) PostPromos(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"message": "create promo"})
}

func (s *Server) GetPromos(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"message": "get promos"})
}

func (s *Server) GetPromosById(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"message": "get promo by id"})
}

func (s *Server) PutPromosById(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"message": "update promo"})
}

func (s *Server) DeletePromosById(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"message": "delete promo"})
}
