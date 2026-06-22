package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leonelortega/cards-reminder-api/internal/i18n"
	"github.com/leonelortega/cards-reminder-api/internal/middleware"
	"github.com/leonelortega/cards-reminder-api/internal/repository"
	"github.com/leonelortega/cards-reminder-api/internal/service"
)

type AuthHandler struct {
	userService  *service.UserService
	termsVersion string
}

func NewAuthHandler(userService *service.UserService, termsVersion string) *AuthHandler {
	return &AuthHandler{
		userService:  userService,
		termsVersion: termsVersion,
	}
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

func (h *AuthHandler) DeleteAccount(c *gin.Context) {
	user, ok := middleware.UserFromContext(c)
	if !ok {
		respondUnauthenticated(c)
		return
	}

	if err := h.userService.DeleteAccount(c.Request.Context(), user); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			respondError(c, http.StatusNotFound, i18n.ErrUserNotFound)
			return
		}
		respondError(c, http.StatusInternalServerError, i18n.ErrFailedToDeleteAccount)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *AuthHandler) AcceptTerms(c *gin.Context) {
	user, ok := middleware.UserFromContext(c)
	if !ok {
		respondUnauthenticated(c)
		return
	}

	updated, err := h.userService.AcceptTerms(c.Request.Context(), user.ID, h.termsVersion)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			respondError(c, http.StatusNotFound, i18n.ErrUserNotFound)
			return
		}
		respondError(c, http.StatusInternalServerError, i18n.ErrFailedToAcceptTerms)
		return
	}

	c.JSON(http.StatusOK, updated)
}

func (h *AuthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "API funcionando"})
}
