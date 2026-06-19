package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leonelortega/cards-reminder-api/internal/domain"
	"github.com/leonelortega/cards-reminder-api/internal/i18n"
	"github.com/leonelortega/cards-reminder-api/internal/middleware"
	"github.com/leonelortega/cards-reminder-api/internal/service"
)

type NotificationHandler struct {
	notificationService *service.NotificationService
}

func NewNotificationHandler(notificationService *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{notificationService: notificationService}
}

type sendTestNotificationRequest struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

func (h *NotificationHandler) SendTest(c *gin.Context) {
	user, ok := middleware.UserFromContext(c)
	if !ok {
		respondUnauthenticated(c)
		return
	}

	var req sendTestNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	title := req.Title
	if title == "" {
		title = "Cards Reminder"
	}

	body := req.Body
	if body == "" {
		if middleware.LanguageFromContext(c) == "en" {
			body = "Test notification"
		} else {
			body = "Notificación de prueba"
		}
	}

	result, err := h.notificationService.SendToUser(c.Request.Context(), user.ID, domain.PushNotification{
		Title: title,
		Body:  body,
		Data: map[string]string{
			"type": "test",
		},
	})
	if err != nil {
		if errors.Is(err, service.ErrNoDeviceTokens) {
			respondError(c, http.StatusNotFound, i18n.ErrNoDeviceTokensRegistered)
			return
		}
		respondError(c, http.StatusInternalServerError, i18n.ErrFailedToSendNotification)
		return
	}

	c.JSON(http.StatusOK, result)
}
