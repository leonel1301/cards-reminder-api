package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leonelortega/cards-reminder-api/internal/i18n"
	"github.com/leonelortega/cards-reminder-api/internal/middleware"
	"github.com/leonelortega/cards-reminder-api/internal/service"
)

func respondError(c *gin.Context, status int, key i18n.ErrorKey) {
	c.JSON(status, gin.H{
		"error": i18n.Error(middleware.LanguageFromContext(c), key),
	})
}

func respondValidationError(c *gin.Context, err service.ValidationError) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error": i18n.ValidationErrorMessage(middleware.LanguageFromContext(c), err.Field, err.Message),
	})
}

func respondUnauthenticated(c *gin.Context) {
	respondError(c, http.StatusUnauthorized, i18n.ErrUnauthenticated)
}
