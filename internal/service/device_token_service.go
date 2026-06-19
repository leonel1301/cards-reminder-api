package service

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/leonelortega/cards-reminder-api/internal/domain"
	"github.com/leonelortega/cards-reminder-api/internal/repository"
)

type DeviceTokenService struct {
	repo *repository.DeviceTokenRepository
}

func NewDeviceTokenService(repo *repository.DeviceTokenRepository) *DeviceTokenService {
	return &DeviceTokenService{repo: repo}
}

func (s *DeviceTokenService) Register(ctx context.Context, userID uuid.UUID, input domain.RegisterDeviceInput) (*domain.DeviceToken, error) {
	input.FCMToken = strings.TrimSpace(input.FCMToken)
	if input.FCMToken == "" {
		return nil, ValidationError{Field: "fcm_token", Message: "is required"}
	}

	if input.Platform = strings.TrimSpace(strings.ToLower(input.Platform)); input.Platform == "" {
		input.Platform = "ios"
	}

	if input.Language = normalizeLanguage(input.Language); input.Language == "" {
		input.Language = "es"
	}

	input.Timezone = NormalizeTimezone(input.Timezone)

	return s.repo.Upsert(ctx, userID, input)
}

func (s *DeviceTokenService) GetTimezoneForUser(ctx context.Context, userID uuid.UUID) (string, error) {
	timezone, err := s.repo.GetLatestTimezoneByUserID(ctx, userID)
	if err != nil {
		return "", err
	}
	return ResolveTimezone(timezone, DefaultTimezone), nil
}

func (s *DeviceTokenService) GetLanguageForUser(ctx context.Context, userID uuid.UUID) (string, error) {
	language, err := s.repo.GetLatestLanguageByUserID(ctx, userID)
	if err != nil {
		return "", err
	}
	language = normalizeLanguage(language)
	if language == "" {
		return "es", nil
	}
	return language, nil
}

func (s *DeviceTokenService) Unregister(ctx context.Context, userID uuid.UUID, fcmToken string) error {
	fcmToken = strings.TrimSpace(fcmToken)
	if fcmToken == "" {
		return ValidationError{Field: "fcm_token", Message: "is required"}
	}

	return s.repo.DeleteByTokenAndUserID(ctx, userID, fcmToken)
}

func normalizeLanguage(language string) string {
	language = strings.TrimSpace(strings.ToLower(language))
	if language == "" {
		return ""
	}

	if idx := strings.Index(language, "-"); idx > 0 {
		language = language[:idx]
	}

	return language
}
