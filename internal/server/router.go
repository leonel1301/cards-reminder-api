package server

import (
	"github.com/gin-gonic/gin"
	"github.com/leonelortega/cards-reminder-api/internal/handler"
	"github.com/leonelortega/cards-reminder-api/internal/middleware"
)

type Router struct {
	authHandler *handler.AuthHandler
	auth        *middleware.AuthMiddleware
}

func NewRouter(authHandler *handler.AuthHandler, auth *middleware.AuthMiddleware) *Router {
	return &Router{
		authHandler: authHandler,
		auth:        auth,
	}
}

func (r *Router) Setup() *gin.Engine {
	router := gin.Default()

	router.GET("/health", r.authHandler.Health)

	authGroup := router.Group("/")
	authGroup.Use(r.auth.RequireAuth(), r.auth.RequireUser())
	{
		authGroup.POST("/auth/session", r.authHandler.CreateSession)
		authGroup.GET("/me", r.authHandler.GetMe)
	}

	return router
}
