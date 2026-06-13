package domain

import (
	"time"

	"github.com/google/uuid"
)

type Owner struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Name      string    `json:"name"`
	SalaryDay *int      `json:"salary_day"`
	IsSelf    bool      `json:"is_self"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateOwnerInput struct {
	Name      string
	SalaryDay *int
}

type UpdateOwnerInput struct {
	Name      *string
	SalaryDay *int
}
