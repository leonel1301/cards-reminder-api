package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leonelortega/cards-reminder-api/internal/middleware"
)

type AuthHandler struct{}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

func (h *AuthHandler) CreateSession(c *gin.Context) {
	user, ok := middleware.UserFromContext(c)
	if !ok {
		respondUnauthenticated(c)
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *AuthHandler) GetMe(c *gin.Context) {
	user, ok := middleware.UserFromContext(c)
	if !ok {
		respondUnauthenticated(c)
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *AuthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "API funcionando"})
}
