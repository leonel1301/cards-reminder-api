package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/leonelortega/cards-reminder-api/internal/auth"
	"github.com/leonelortega/cards-reminder-api/internal/config"
	"github.com/leonelortega/cards-reminder-api/internal/database"
	"github.com/leonelortega/cards-reminder-api/internal/handler"
	"github.com/leonelortega/cards-reminder-api/internal/middleware"
	"github.com/leonelortega/cards-reminder-api/internal/repository"
	"github.com/leonelortega/cards-reminder-api/internal/server"
	"github.com/leonelortega/cards-reminder-api/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	firebaseApp, err := auth.NewFirebase(ctx, cfg.FirebaseCredentialsPath)
	if err != nil {
		log.Fatalf("firebase: %v", err)
	}

	userRepo := repository.NewUserRepository(pool)
	ownerRepo := repository.NewOwnerRepository(pool)
	userService := service.NewUserService(userRepo, ownerRepo)
	cardRepo := repository.NewCardRepository(pool)
	cardService := service.NewCardService(cardRepo, ownerRepo)
	ownerService := service.NewOwnerService(ownerRepo)
	paymentRepo := repository.NewPaymentRepository(pool)
	cardStatusService := service.NewCardStatusService(cardRepo, paymentRepo, ownerRepo)
	authMiddleware := middleware.NewAuthMiddleware(firebaseApp.Auth, userService)
	authHandler := handler.NewAuthHandler()
	cardHandler := handler.NewCardHandler(cardService)
	deviceTokenRepo := repository.NewDeviceTokenRepository(pool)
	deviceTokenService := service.NewDeviceTokenService(deviceTokenRepo)
	cardStatusHandler := handler.NewCardStatusHandler(cardStatusService, deviceTokenService)
	ownerHandler := handler.NewOwnerHandler(ownerService)
	deviceHandler := handler.NewDeviceHandler(deviceTokenService)
	notificationService := service.NewNotificationService(deviceTokenRepo, firebaseApp.Messaging)
	notificationHandler := handler.NewNotificationHandler(notificationService)

	router := server.NewRouter(
		authHandler,
		cardHandler,
		cardStatusHandler,
		ownerHandler,
		deviceHandler,
		notificationHandler,
		authMiddleware,
	).Setup()

	go func() {
		if err := router.Run(":" + cfg.Port); err != nil {
			log.Fatalf("server: %v", err)
		}
	}()

	log.Printf("server listening on :%s", cfg.Port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	_ = shutdownCtx
}
