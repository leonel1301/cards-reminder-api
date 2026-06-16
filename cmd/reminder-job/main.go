package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/leonelortega/cards-reminder-api/internal/auth"
	"github.com/leonelortega/cards-reminder-api/internal/config"
	"github.com/leonelortega/cards-reminder-api/internal/database"
	"github.com/leonelortega/cards-reminder-api/internal/job"
	"github.com/leonelortega/cards-reminder-api/internal/repository"
	"github.com/leonelortega/cards-reminder-api/internal/service"
)

func main() {
	cfg, err := config.LoadJob()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx := context.Background()

	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	firebaseApp, err := auth.NewFirebase(ctx, cfg.FirebaseCredentialsPath)
	if err != nil {
		log.Fatalf("firebase: %v", err)
	}

	deviceRepo := repository.NewDeviceTokenRepository(pool)
	cardRepo := repository.NewCardRepository(pool)
	paymentRepo := repository.NewPaymentRepository(pool)
	ownerRepo := repository.NewOwnerRepository(pool)
	cardStatusSvc := service.NewCardStatusService(cardRepo, paymentRepo, ownerRepo)
	notificationSvc := service.NewNotificationService(deviceRepo, firebaseApp.Messaging)

	reminderJob := job.NewReminderJob(deviceRepo, ownerRepo, cardStatusSvc, notificationSvc, cfg.ReminderTimezone, cfg.ReminderHour)

	log.Printf("reminder job started default_timezone=%s reminder_hour=%d", cfg.ReminderTimezone, cfg.ReminderHour)

	result, err := reminderJob.Run(ctx)
	if err != nil {
		log.Fatalf("reminder job failed: %v", err)
	}

	output, _ := json.MarshalIndent(result, "", "  ")
	log.Printf("reminder job finished:\n%s", output)

	if result.SendFailures > 0 {
		os.Exit(1)
	}
}
