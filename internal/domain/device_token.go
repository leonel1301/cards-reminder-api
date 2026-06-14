package domain

import (
	"time"

	"github.com/google/uuid"
)

type DeviceToken struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	FCMToken  string    `json:"fcm_token"`
	Platform  string    `json:"platform"`
	Language  string    `json:"language"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RegisterDeviceInput struct {
	FCMToken string
	Platform string
	Language string
}
