package middleware

import (
	"net/http"
	"strings"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/leonelortega/cards-reminder-api/internal/domain"
	"github.com/leonelortega/cards-reminder-api/internal/i18n"
	"github.com/leonelortega/cards-reminder-api/internal/service"
)

const (
	ContextKeyFirebaseUID = "firebase_uid"
	ContextKeyEmail       = "email"
	ContextKeyDisplayName = "display_name"
	ContextKeyUser        = "user"
)

type AuthMiddleware struct {
	authClient  *auth.Client
	userService *service.UserService
}

func NewAuthMiddleware(authClient *auth.Client, userService *service.UserService) *AuthMiddleware {
	return &AuthMiddleware{
		authClient:  authClient,
		userService: userService,
	}
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := extractBearerToken(c.GetHeader("Authorization"))
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": i18n.Error(LanguageFromContext(c), i18n.ErrMissingAuthorizationHeader),
			})
			return
		}

		firebaseToken, err := m.authClient.VerifyIDToken(c.Request.Context(), token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": i18n.Error(LanguageFromContext(c), i18n.ErrInvalidToken),
			})
			return
		}

		email := stringClaim(firebaseToken.Claims, "email")
		displayName := stringClaim(firebaseToken.Claims, "name")

		c.Set(ContextKeyFirebaseUID, firebaseToken.UID)
		c.Set(ContextKeyEmail, email)
		c.Set(ContextKeyDisplayName, displayName)
		c.Next()
	}
}

func (m *AuthMiddleware) RequireUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		firebaseUID, ok := c.Get(ContextKeyFirebaseUID)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": i18n.Error(LanguageFromContext(c), i18n.ErrUnauthenticated),
			})
			return
		}

		email, _ := c.Get(ContextKeyEmail)
		displayName, _ := c.Get(ContextKeyDisplayName)

		user, err := m.userService.GetOrCreate(
			c.Request.Context(),
			firebaseUID.(string),
			email.(*string),
			displayName.(*string),
		)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": i18n.Error(LanguageFromContext(c), i18n.ErrFailedToResolveUser),
			})
			return
		}

		c.Set(ContextKeyUser, user)
		c.Next()
	}
}

func UserFromContext(c *gin.Context) (*domain.User, bool) {
	value, ok := c.Get(ContextKeyUser)
	if !ok {
		return nil, false
	}

	user, ok := value.(*domain.User)
	return user, ok
}

func extractBearerToken(header string) (string, bool) {
	if header == "" {
		return "", false
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", false
	}

	return parts[1], true
}

func stringClaim(claims map[string]interface{}, key string) *string {
	value, ok := claims[key].(string)
	if !ok || value == "" {
		return nil
	}
	return &value
}
