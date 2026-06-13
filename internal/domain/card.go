package domain

import (
	"time"

	"github.com/google/uuid"
)

type Card struct {
	ID              uuid.UUID `json:"id"`
	UserID          uuid.UUID `json:"user_id"`
	OwnerID         uuid.UUID `json:"owner_id"`
	Name            string    `json:"name"`
	LastFourDigits  string    `json:"last_four_digits"`
	Issuer          *string   `json:"issuer"`
	BillingCycleDay int       `json:"billing_cycle_day"`
	PaymentDueDay   int       `json:"payment_due_day"`
	ColorHex        *string   `json:"color_hex"`
	Notes           *string   `json:"notes"`
	IsActive        bool      `json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type CreateCardInput struct {
	Name            string
	LastFourDigits  string
	Issuer          *string
	BillingCycleDay int
	PaymentDueDay   int
	ColorHex        *string
	Notes           *string
	OwnerID         *uuid.UUID
}

type UpdateCardInput struct {
	Name            *string
	LastFourDigits  *string
	Issuer          *string
	BillingCycleDay *int
	PaymentDueDay   *int
	ColorHex        *string
	Notes           *string
	IsActive        *bool
	OwnerID         *uuid.UUID
}
