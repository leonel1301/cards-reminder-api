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
	deviceHandler     *handler.DeviceHandler
	notificationHandler *handler.NotificationHandler
	feedbackHandler     *handler.FeedbackHandler
	auth                *middleware.AuthMiddleware
	feedbackAdminToken  string
}

func NewRouter(
	authHandler *handler.AuthHandler,
	cardHandler *handler.CardHandler,
	cardStatusHandler *handler.CardStatusHandler,
	ownerHandler *handler.OwnerHandler,
	deviceHandler *handler.DeviceHandler,
	notificationHandler *handler.NotificationHandler,
	feedbackHandler *handler.FeedbackHandler,
	auth *middleware.AuthMiddleware,
	feedbackAdminToken string,
) *Router {
	return &Router{
		authHandler:         authHandler,
		cardHandler:         cardHandler,
		cardStatusHandler:   cardStatusHandler,
		ownerHandler:        ownerHandler,
		deviceHandler:       deviceHandler,
		notificationHandler: notificationHandler,
		feedbackHandler:     feedbackHandler,
		auth:                auth,
		feedbackAdminToken:  feedbackAdminToken,
	}
}

func (r *Router) Setup() *gin.Engine {
	router := gin.Default()

	router.GET("/health", r.authHandler.Health)

	feedbackAdminGroup := router.Group("/")
	feedbackAdminGroup.Use(middleware.ResolveLanguage(), middleware.RequireFeedbackAdminToken(r.feedbackAdminToken))
	{
		feedbackAdminGroup.GET("/feedback", r.feedbackHandler.ListForAdmin)
	}

	authGroup := router.Group("/")
	authGroup.Use(middleware.ResolveLanguage(), r.auth.RequireAuth(), r.auth.RequireUser())
	{
		authGroup.POST("/auth/session", r.authHandler.CreateSession)
		authGroup.GET("/me", r.authHandler.GetMe)
		authGroup.GET("/me/feedback", r.feedbackHandler.ListByUser)
		authGroup.DELETE("/me", r.authHandler.DeleteAccount)

		authGroup.GET("/owners", r.ownerHandler.List)
		authGroup.POST("/owners", r.ownerHandler.Create)
		authGroup.GET("/owners/:id", r.ownerHandler.Get)
		authGroup.PATCH("/owners/:id", r.ownerHandler.Update)
		authGroup.DELETE("/owners/:id", r.ownerHandler.Delete)

		authGroup.PUT("/devices", r.deviceHandler.Register)
		authGroup.DELETE("/devices", r.deviceHandler.Unregister)

		authGroup.POST("/notifications/test", r.notificationHandler.SendTest)

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

		authGroup.POST("/feedback", r.feedbackHandler.Create)
		authGroup.GET("/feedback/:id", r.feedbackHandler.Get)
		authGroup.PATCH("/feedback/:id", r.feedbackHandler.Update)
		authGroup.DELETE("/feedback/:id", r.feedbackHandler.Delete)
	}

	return router
}
