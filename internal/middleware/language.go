package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/leonelortega/cards-reminder-api/internal/i18n"
)

const ContextKeyLanguage = "language"

func ResolveLanguage() gin.HandlerFunc {
	return func(c *gin.Context) {
		language := i18n.ParseAcceptLanguage(c.GetHeader("Accept-Language"))
		c.Set(ContextKeyLanguage, language)
		c.Next()
	}
}

func LanguageFromContext(c *gin.Context) string {
	value, ok := c.Get(ContextKeyLanguage)
	if !ok {
		return "es"
	}

	language, ok := value.(string)
	if !ok || language == "" {
		return "es"
	}

	return language
}
