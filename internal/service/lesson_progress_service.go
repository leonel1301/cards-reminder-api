package service

import (
	"context"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/leonelortega/cards-reminder-api/internal/domain"
	"github.com/leonelortega/cards-reminder-api/internal/repository"
)

var lessonIDPattern = regexp.MustCompile(`^[a-z0-9_]{1,64}$`)

type LessonProgressService struct {
	repo *repository.LessonProgressRepository
}

func NewLessonProgressService(repo *repository.LessonProgressRepository) *LessonProgressService {
	return &LessonProgressService{repo: repo}
}

func (s *LessonProgressService) List(ctx context.Context, userID uuid.UUID) (*domain.LessonProgressListResponse, error) {
	ids, err := s.repo.ListLessonIDsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if ids == nil {
		ids = []string{}
	}
	return &domain.LessonProgressListResponse{
		LessonIDs:      ids,
		CompletedCount: len(ids),
	}, nil
}

func (s *LessonProgressService) Mark(ctx context.Context, userID uuid.UUID, lessonID string) (*domain.LessonProgress, error) {
	normalized, err := normalizeLessonID(lessonID)
	if err != nil {
		return nil, err
	}
	return s.repo.MarkCompleted(ctx, userID, normalized)
}

func (s *LessonProgressService) Unmark(ctx context.Context, userID uuid.UUID, lessonID string) error {
	normalized, err := normalizeLessonID(lessonID)
	if err != nil {
		return err
	}
	return s.repo.UnmarkCompleted(ctx, userID, normalized)
}

func normalizeLessonID(lessonID string) (string, error) {
	id := strings.TrimSpace(strings.ToLower(lessonID))
	if id == "" {
		return "", ValidationError{Field: "lesson_id", Message: "is required"}
	}
	if !lessonIDPattern.MatchString(id) {
		return "", ValidationError{Field: "lesson_id", Message: "has an invalid format"}
	}
	return id, nil
}
