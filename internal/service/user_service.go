package service

import (
	"context"
	"fmt"

	"firebase.google.com/go/v4/auth"
	"github.com/leonelortega/cards-reminder-api/internal/domain"
	"github.com/leonelortega/cards-reminder-api/internal/repository"
)

type UserService struct {
	userRepo   *repository.UserRepository
	ownerRepo  *repository.OwnerRepository
	authClient *auth.Client
}

func NewUserService(userRepo *repository.UserRepository, ownerRepo *repository.OwnerRepository, authClient *auth.Client) *UserService {
	return &UserService{
		userRepo:   userRepo,
		ownerRepo:  ownerRepo,
		authClient: authClient,
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

func (s *UserService) DeleteAccount(ctx context.Context, user *domain.User) error {
	if err := s.userRepo.Delete(ctx, user.ID); err != nil {
		return err
	}

	if err := s.authClient.DeleteUser(ctx, user.FirebaseUID); err != nil {
		return fmt.Errorf("delete firebase user: %w", err)
	}

	return nil
}
