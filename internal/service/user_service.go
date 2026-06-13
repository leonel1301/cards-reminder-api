package service

import (
	"context"

	"github.com/leonelortega/cards-reminder-api/internal/domain"
	"github.com/leonelortega/cards-reminder-api/internal/repository"
)

type UserService struct {
	userRepo  *repository.UserRepository
	ownerRepo *repository.OwnerRepository
}

func NewUserService(userRepo *repository.UserRepository, ownerRepo *repository.OwnerRepository) *UserService {
	return &UserService{
		userRepo:  userRepo,
		ownerRepo: ownerRepo,
	}
}

func (s *UserService) GetOrCreate(ctx context.Context, firebaseUID string, email, displayName *string) (*domain.User, error) {
	user, err := s.userRepo.Upsert(ctx, firebaseUID, email, displayName)
	if err != nil {
		return nil, err
	}

	if err := s.ownerRepo.EnsureSelfOwner(ctx, user.ID, displayName); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetByFirebaseUID(ctx context.Context, firebaseUID string) (*domain.User, error) {
	return s.userRepo.GetByFirebaseUID(ctx, firebaseUID)
}
