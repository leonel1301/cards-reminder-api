package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leonelortega/cards-reminder-api/internal/domain"
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
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
		body = "Notificación de prueba"
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
			c.JSON(http.StatusNotFound, gin.H{"error": "no device tokens registered"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send notification"})
		return
	}

	c.JSON(http.StatusOK, result)
}
