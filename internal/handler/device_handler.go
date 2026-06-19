package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leonelortega/cards-reminder-api/internal/domain"
	"github.com/leonelortega/cards-reminder-api/internal/i18n"
	"github.com/leonelortega/cards-reminder-api/internal/middleware"
	"github.com/leonelortega/cards-reminder-api/internal/repository"
	"github.com/leonelortega/cards-reminder-api/internal/service"
)

type DeviceHandler struct {
	deviceService *service.DeviceTokenService
}

func NewDeviceHandler(deviceService *service.DeviceTokenService) *DeviceHandler {
	return &DeviceHandler{deviceService: deviceService}
}

type registerDeviceRequest struct {
	FCMToken string `json:"fcm_token" binding:"required"`
	Platform string `json:"platform"`
	Language string `json:"language"`
	Timezone string `json:"timezone"`
}

type unregisterDeviceRequest struct {
	FCMToken string `json:"fcm_token" binding:"required"`
}

func (h *DeviceHandler) Register(c *gin.Context) {
	user, ok := middleware.UserFromContext(c)
	if !ok {
		respondUnauthenticated(c)
		return
	}

	var req registerDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.deviceService.Register(c.Request.Context(), user.ID, domain.RegisterDeviceInput{
		FCMToken: req.FCMToken,
		Platform: req.Platform,
		Language: req.Language,
		Timezone: req.Timezone,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, token)
}

func (h *DeviceHandler) Unregister(c *gin.Context) {
	user, ok := middleware.UserFromContext(c)
	if !ok {
		respondUnauthenticated(c)
		return
	}

	var req unregisterDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.deviceService.Unregister(c.Request.Context(), user.ID, req.FCMToken); err != nil {
		h.handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *DeviceHandler) handleError(c *gin.Context, err error) {
	var validationErr service.ValidationError
	switch {
	case errors.As(err, &validationErr):
		respondValidationError(c, validationErr)
	case errors.Is(err, repository.ErrNotFound):
		respondError(c, http.StatusNotFound, i18n.ErrDeviceTokenNotFound)
	default:
		log.Printf("device handler error: %v", err)
		respondError(c, http.StatusInternalServerError, i18n.ErrInternalServerError)
	}
}
