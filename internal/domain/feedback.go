package domain

import (
	"time"

	"github.com/google/uuid"
)

type Feedback struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Title     string    `json:"title"`
	Device    string    `json:"device"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateFeedbackInput struct {
	Title   string
	Device  string
	Content string
}

type UpdateFeedbackInput struct {
	Title   *string
	Device  *string
	Content *string
}

type FeedbackAdminItem struct {
	ID        uuid.UUID `json:"id"`
	UserName  string    `json:"user_name"`
	Title     string    `json:"title"`
	Device    string    `json:"device"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
