package domain

import (
	"time"

	"github.com/google/uuid"
)

type LessonProgress struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	LessonID    string    `json:"lesson_id"`
	CompletedAt time.Time `json:"completed_at"`
}

type LessonProgressListResponse struct {
	LessonIDs      []string `json:"lesson_ids"`
	CompletedCount int      `json:"completed_count"`
}
