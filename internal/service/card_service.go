package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/leonelortega/cards-reminder-api/internal/domain"
	"github.com/leonelortega/cards-reminder-api/internal/repository"
)

type CardService struct {
	cardRepo  *repository.CardRepository
	ownerRepo *repository.OwnerRepository
}

func NewCardService(cardRepo *repository.CardRepository, ownerRepo *repository.OwnerRepository) *CardService {
	return &CardService{
		cardRepo:  cardRepo,
		ownerRepo: ownerRepo,
	}
}

func (s *CardService) List(ctx context.Context, userID uuid.UUID) ([]domain.Card, error) {
	return s.cardRepo.ListByUserID(ctx, userID)
}

func (s *CardService) Get(ctx context.Context, userID, cardID uuid.UUID) (*domain.Card, error) {
	return s.cardRepo.GetByIDAndUserID(ctx, cardID, userID)
}

func (s *CardService) Create(ctx context.Context, userID uuid.UUID, input domain.CreateCardInput) (*domain.Card, error) {
	if err := validateCreateInput(input); err != nil {
		return nil, err
	}

	ownerID, err := s.resolveOwnerID(ctx, userID, input.OwnerID)
	if err != nil {
		return nil, err
	}
	input.OwnerID = &ownerID

	return s.cardRepo.Create(ctx, userID, input)
}

func (s *CardService) Update(ctx context.Context, userID, cardID uuid.UUID, input domain.UpdateCardInput) (*domain.Card, error) {
	if err := validateUpdateInput(input); err != nil {
		return nil, err
	}

	if input.OwnerID != nil {
		if _, err := s.ownerRepo.GetByIDAndUserID(ctx, *input.OwnerID, userID); err != nil {
			return nil, ValidationError{Field: "owner_id", Message: "invalid owner"}
		}
	}

	return s.cardRepo.Update(ctx, cardID, userID, input)
}

func (s *CardService) Delete(ctx context.Context, userID, cardID uuid.UUID) error {
	return s.cardRepo.Delete(ctx, cardID, userID)
}

func (s *CardService) resolveOwnerID(ctx context.Context, userID uuid.UUID, ownerID *uuid.UUID) (uuid.UUID, error) {
	if ownerID != nil {
		owner, err := s.ownerRepo.GetByIDAndUserID(ctx, *ownerID, userID)
		if err != nil {
			return uuid.Nil, ValidationError{Field: "owner_id", Message: "invalid owner"}
		}
		return owner.ID, nil
	}

	selfOwner, err := s.ownerRepo.GetSelfByUserID(ctx, userID)
	if err != nil {
		return uuid.Nil, err
	}
	return selfOwner.ID, nil
}
