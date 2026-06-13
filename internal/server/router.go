package server

import (
	"github.com/gin-gonic/gin"
	"github.com/leonelortega/cards-reminder-api/internal/handler"
	"github.com/leonelortega/cards-reminder-api/internal/middleware"
)

type Router struct {
	authHandler       *handler.AuthHandler
	cardHandler       *handler.CardHandler
	cardStatusHandler *handler.CardStatusHandler
	ownerHandler      *handler.OwnerHandler
	auth              *middleware.AuthMiddleware
}

func NewRouter(
	authHandler *handler.AuthHandler,
	cardHandler *handler.CardHandler,
	cardStatusHandler *handler.CardStatusHandler,
	ownerHandler *handler.OwnerHandler,
	auth *middleware.AuthMiddleware,
) *Router {
	return &Router{
		authHandler:       authHandler,
		cardHandler:       cardHandler,
		cardStatusHandler: cardStatusHandler,
		ownerHandler:      ownerHandler,
		auth:              auth,
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

		authGroup.GET("/owners", r.ownerHandler.List)
		authGroup.POST("/owners", r.ownerHandler.Create)
		authGroup.GET("/owners/:id", r.ownerHandler.Get)
		authGroup.PATCH("/owners/:id", r.ownerHandler.Update)
		authGroup.DELETE("/owners/:id", r.ownerHandler.Delete)

		authGroup.GET("/dashboard", r.cardStatusHandler.GetDashboard)

		authGroup.GET("/cards", r.cardHandler.List)
		authGroup.POST("/cards", r.cardHandler.Create)
		authGroup.GET("/cards/:id/status", r.cardStatusHandler.GetStatus)
		authGroup.GET("/cards/:id/optimal-purchase-days", r.cardStatusHandler.GetOptimalPurchaseDays)
		authGroup.GET("/cards/:id/current-cycle", r.cardStatusHandler.GetCurrentCycle)
		authGroup.GET("/cards/:id/payments", r.cardStatusHandler.ListPayments)
		authGroup.POST("/cards/:id/payments", r.cardStatusHandler.MarkPaid)
		authGroup.GET("/cards/:id", r.cardHandler.Get)
		authGroup.PATCH("/cards/:id", r.cardHandler.Update)
		authGroup.DELETE("/cards/:id", r.cardHandler.Delete)
	}

	return router
}
