package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/leonelortega/cards-reminder-api/internal/i18n"
	"github.com/leonelortega/cards-reminder-api/internal/middleware"
	"github.com/leonelortega/cards-reminder-api/internal/repository"
	"github.com/leonelortega/cards-reminder-api/internal/service"
)

type CardStatusHandler struct {
	cardStatusService  *service.CardStatusService
	deviceTokenService *service.DeviceTokenService
}

func NewCardStatusHandler(cardStatusService *service.CardStatusService, deviceTokenService *service.DeviceTokenService) *CardStatusHandler {
	return &CardStatusHandler{
		cardStatusService:  cardStatusService,
		deviceTokenService: deviceTokenService,
	}
}

type markPaidRequest struct {
	Notes *string `json:"notes"`
}

func (h *CardStatusHandler) GetStatus(c *gin.Context) {
	user, ok := middleware.UserFromContext(c)
	if !ok {
		respondUnauthenticated(c)
		return
	}

	cardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, i18n.ErrInvalidCardID)
		return
	}

	timezone, ok := h.userTimezone(c, user.ID)
	if !ok {
		return
	}

	response, err := h.cardStatusService.GetStatus(c.Request.Context(), user.ID, cardID, timezone)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *CardStatusHandler) GetOptimalPurchaseDays(c *gin.Context) {
	user, ok := middleware.UserFromContext(c)
	if !ok {
		respondUnauthenticated(c)
		return
	}

	cardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, i18n.ErrInvalidCardID)
		return
	}

	timezone, ok := h.userTimezone(c, user.ID)
	if !ok {
		return
	}

	response, err := h.cardStatusService.GetOptimalPurchaseDays(c.Request.Context(), user.ID, cardID, timezone)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *CardStatusHandler) GetDashboard(c *gin.Context) {
	user, ok := middleware.UserFromContext(c)
	if !ok {
		respondUnauthenticated(c)
		return
	}

	timezone, ok := h.userTimezone(c, user.ID)
	if !ok {
		return
	}

	language := middleware.LanguageFromContext(c)

	response, err := h.cardStatusService.GetDashboard(c.Request.Context(), user.ID, timezone, language)
	if err != nil {
		respondError(c, http.StatusInternalServerError, i18n.ErrFailedToBuildDashboard)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *CardStatusHandler) GetCurrentCycle(c *gin.Context) {
	user, ok := middleware.UserFromContext(c)
	if !ok {
		respondUnauthenticated(c)
		return
	}

	cardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, i18n.ErrInvalidCardID)
		return
	}

	timezone, ok := h.userTimezone(c, user.ID)
	if !ok {
		return
	}

	response, err := h.cardStatusService.GetCurrentCycle(c.Request.Context(), user.ID, cardID, timezone)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *CardStatusHandler) ListPayments(c *gin.Context) {
	user, ok := middleware.UserFromContext(c)
	if !ok {
		respondUnauthenticated(c)
		return
	}

	cardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, i18n.ErrInvalidCardID)
		return
	}

	response, err := h.cardStatusService.ListPayments(c.Request.Context(), user.ID, cardID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *CardStatusHandler) MarkPaid(c *gin.Context) {
	user, ok := middleware.UserFromContext(c)
	if !ok {
		respondUnauthenticated(c)
		return
	}

	cardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, i18n.ErrInvalidCardID)
		return
	}

	var req markPaidRequest
	if err := c.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	timezone, ok := h.userTimezone(c, user.ID)
	if !ok {
		return
	}

	response, err := h.cardStatusService.MarkPaid(c.Request.Context(), user.ID, cardID, req.Notes, timezone)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *CardStatusHandler) userTimezone(c *gin.Context, userID uuid.UUID) (string, bool) {
	timezone, err := h.deviceTokenService.GetTimezoneForUser(c.Request.Context(), userID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, i18n.ErrFailedToResolveTimezone)
		return "", false
	}
	return timezone, true
}

func (h *CardStatusHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		respondError(c, http.StatusNotFound, i18n.ErrCardNotFound)
	default:
		respondError(c, http.StatusInternalServerError, i18n.ErrInternalServerError)
	}
}
