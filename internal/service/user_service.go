package service

import (
	"context"

	"github.com/leonelortega/cards-reminder-api/internal/domain"
	"github.com/leonelortega/cards-reminder-api/internal/repository"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetOrCreate(ctx context.Context, firebaseUID string, email, displayName *string) (*domain.User, error) {
	return s.repo.Upsert(ctx, firebaseUID, email, displayName)
}

func (s *UserService) GetByFirebaseUID(ctx context.Context, firebaseUID string) (*domain.User, error) {
	return s.repo.GetByFirebaseUID(ctx, firebaseUID)
}
