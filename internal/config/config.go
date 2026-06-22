package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                    string
	DatabaseURL             string
	FirebaseCredentialsPath string
	FeedbackAdminToken      string
}

type JobConfig struct {
	DatabaseURL             string
	FirebaseCredentialsPath string
	ReminderTimezone        string
	ReminderHour            int
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Port:                    envOrDefault("PORT", "8080"),
		DatabaseURL:             os.Getenv("DATABASE_URL"),
		FirebaseCredentialsPath: os.Getenv("FIREBASE_CREDENTIALS_PATH"),
		FeedbackAdminToken:      os.Getenv("FEEDBACK_ADMIN_TOKEN"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.FirebaseCredentialsPath == "" {
		return nil, fmt.Errorf("FIREBASE_CREDENTIALS_PATH is required")
	}
	if cfg.FeedbackAdminToken == "" {
		return nil, fmt.Errorf("FEEDBACK_ADMIN_TOKEN is required")
	}

	return cfg, nil
}

func LoadJob() (*JobConfig, error) {
	_ = godotenv.Load()

	cfg := &JobConfig{
		DatabaseURL:             os.Getenv("DATABASE_URL"),
		FirebaseCredentialsPath: os.Getenv("FIREBASE_CREDENTIALS_PATH"),
		ReminderTimezone:        envOrDefault("REMINDER_TIMEZONE", "America/Lima"),
		ReminderHour:            envOrDefaultInt("REMINDER_HOUR", 8),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.FirebaseCredentialsPath == "" {
		return nil, fmt.Errorf("FIREBASE_CREDENTIALS_PATH is required")
	}

	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envOrDefaultInt(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 || value > 23 {
		return fallback
	}

	return value
}
