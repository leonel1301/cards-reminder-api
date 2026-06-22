package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/leonelortega/cards-reminder-api/internal/i18n"
)

const feedbackAdminTokenHeader = "X-Feedback-Token"

func RequireFeedbackAdminToken(adminToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := extractFeedbackAdminToken(c)
		if !ok || token != adminToken {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": i18n.Error(LanguageFromContext(c), i18n.ErrInvalidFeedbackAdminToken),
			})
			return
		}

		c.Next()
	}
}

func extractFeedbackAdminToken(c *gin.Context) (string, bool) {
	if headerToken := strings.TrimSpace(c.GetHeader(feedbackAdminTokenHeader)); headerToken != "" {
		return headerToken, true
	}

	if bearer, ok := extractBearerToken(c.GetHeader("Authorization")); ok {
		return bearer, true
	}

	return "", false
}
