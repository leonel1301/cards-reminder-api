package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                uuid.UUID  `json:"id"`
	FirebaseUID       string     `json:"firebase_uid"`
	Email             *string    `json:"email"`
	DisplayName       *string    `json:"display_name"`
	TermsAcceptedAt   *time.Time `json:"terms_accepted_at"`
	PrivacyAcceptedAt *time.Time `json:"privacy_accepted_at"`
	TermsVersion      *string    `json:"terms_version"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}
