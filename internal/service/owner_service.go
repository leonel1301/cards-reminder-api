package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/leonelortega/cards-reminder-api/internal/domain"
	"github.com/leonelortega/cards-reminder-api/internal/repository"
)

var ErrOwnerHasCards = errors.New("owner has assigned cards")

type OwnerService struct {
	repo *repository.OwnerRepository
}

func NewOwnerService(repo *repository.OwnerRepository) *OwnerService {
	return &OwnerService{repo: repo}
}

func (s *OwnerService) List(ctx context.Context, userID uuid.UUID) ([]domain.Owner, error) {
	return s.repo.ListByUserID(ctx, userID)
}

func (s *OwnerService) Get(ctx context.Context, userID, ownerID uuid.UUID) (*domain.Owner, error) {
	return s.repo.GetByIDAndUserID(ctx, ownerID, userID)
}

func (s *OwnerService) Create(ctx context.Context, userID uuid.UUID, input domain.CreateOwnerInput) (*domain.Owner, error) {
	if err := validateOwnerName(input.Name); err != nil {
		return nil, err
	}
	if err := validateOptionalDay("salary_day", input.SalaryDay); err != nil {
		return nil, err
	}
	return s.repo.Create(ctx, userID, input)
}

func (s *OwnerService) Update(ctx context.Context, userID, ownerID uuid.UUID, input domain.UpdateOwnerInput) (*domain.Owner, error) {
	if input.Name != nil {
		if err := validateOwnerName(*input.Name); err != nil {
			return nil, err
		}
	}
	if err := validateOptionalDay("salary_day", input.SalaryDay); err != nil {
		return nil, err
	}
	return s.repo.Update(ctx, ownerID, userID, input)
}

func (s *OwnerService) Delete(ctx context.Context, userID, ownerID uuid.UUID) error {
	owner, err := s.repo.GetByIDAndUserID(ctx, ownerID, userID)
	if err != nil {
		return err
	}
	if owner.IsSelf {
		return ValidationError{Field: "owner", Message: "cannot delete self owner"}
	}

	count, err := s.repo.CountCards(ctx, ownerID)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrOwnerHasCards
	}

	return s.repo.Delete(ctx, ownerID, userID)
}

func validateOwnerName(name string) error {
	if name == "" {
		return ValidationError{Field: "name", Message: "is required"}
	}
	return nil
}

func validateOptionalDay(field string, day *int) error {
	if day == nil {
		return nil
	}
	return validateDay(field, *day)
}
