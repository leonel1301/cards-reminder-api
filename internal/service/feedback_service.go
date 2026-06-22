package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/leonelortega/cards-reminder-api/internal/domain"
	"github.com/leonelortega/cards-reminder-api/internal/repository"
)

type FeedbackService struct {
	repo *repository.FeedbackRepository
}

func NewFeedbackService(repo *repository.FeedbackRepository) *FeedbackService {
	return &FeedbackService{repo: repo}
}

func (s *FeedbackService) List(ctx context.Context, userID uuid.UUID) ([]domain.Feedback, error) {
	return s.repo.ListByUserID(ctx, userID)
}

func (s *FeedbackService) ListByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Feedback, error) {
	return s.repo.ListByUserID(ctx, userID)
}

func (s *FeedbackService) ListAllForAdmin(ctx context.Context) ([]domain.FeedbackAdminItem, error) {
	return s.repo.ListAllWithUserName(ctx)
}

func (s *FeedbackService) Get(ctx context.Context, userID, feedbackID uuid.UUID) (*domain.Feedback, error) {
	return s.repo.GetByIDAndUserID(ctx, feedbackID, userID)
}

func (s *FeedbackService) Create(ctx context.Context, userID uuid.UUID, input domain.CreateFeedbackInput) (*domain.Feedback, error) {
	if err := validateFeedbackFields(input.Title, input.Device, input.Content); err != nil {
		return nil, err
	}
	return s.repo.Create(ctx, userID, input)
}

func (s *FeedbackService) Update(ctx context.Context, userID, feedbackID uuid.UUID, input domain.UpdateFeedbackInput) (*domain.Feedback, error) {
	if input.Title != nil && *input.Title == "" {
		return nil, ValidationError{Field: "title", Message: "cannot be empty"}
	}
	if input.Device != nil && *input.Device == "" {
		return nil, ValidationError{Field: "device", Message: "cannot be empty"}
	}
	if input.Content != nil && *input.Content == "" {
		return nil, ValidationError{Field: "content", Message: "cannot be empty"}
	}
	return s.repo.Update(ctx, feedbackID, userID, input)
}

func (s *FeedbackService) Delete(ctx context.Context, userID, feedbackID uuid.UUID) error {
	return s.repo.Delete(ctx, feedbackID, userID)
}

func validateFeedbackFields(title, device, content string) error {
	if title == "" {
		return ValidationError{Field: "title", Message: "is required"}
	}
	if device == "" {
		return ValidationError{Field: "device", Message: "is required"}
	}
	if content == "" {
		return ValidationError{Field: "content", Message: "is required"}
	}
	return nil
}
