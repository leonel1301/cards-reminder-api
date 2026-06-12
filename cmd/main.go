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

	authClient, err := auth.NewFirebaseAuthClient(ctx, cfg.FirebaseCredentialsPath)
	if err != nil {
		log.Fatalf("firebase: %v", err)
	}

	userRepo := repository.NewUserRepository(pool)
	userService := service.NewUserService(userRepo)
	authMiddleware := middleware.NewAuthMiddleware(authClient, userService)
	authHandler := handler.NewAuthHandler()

	router := server.NewRouter(authHandler, authMiddleware).Setup()

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
