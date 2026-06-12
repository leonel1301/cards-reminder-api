package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leonelortega/cards-reminder-api/internal/domain"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) GetByFirebaseUID(ctx context.Context, firebaseUID string) (*domain.User, error) {
	const query = `
		SELECT id, firebase_uid, email, display_name, created_at, updated_at
		FROM users
		WHERE firebase_uid = $1
	`

	var user domain.User
	err := r.pool.QueryRow(ctx, query, firebaseUID).Scan(
		&user.ID,
		&user.FirebaseUID,
		&user.Email,
		&user.DisplayName,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) Upsert(ctx context.Context, firebaseUID string, email, displayName *string) (*domain.User, error) {
	const query = `
		INSERT INTO users (firebase_uid, email, display_name)
		VALUES ($1, $2, $3)
		ON CONFLICT (firebase_uid) DO UPDATE SET
			email = COALESCE(EXCLUDED.email, users.email),
			display_name = COALESCE(EXCLUDED.display_name, users.display_name),
			updated_at = now()
		RETURNING id, firebase_uid, email, display_name, created_at, updated_at
	`

	var user domain.User
	err := r.pool.QueryRow(ctx, query, firebaseUID, email, displayName).Scan(
		&user.ID,
		&user.FirebaseUID,
		&user.Email,
		&user.DisplayName,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("upsert user: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	const query = `
		SELECT id, firebase_uid, email, display_name, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user domain.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.FirebaseUID,
		&user.Email,
		&user.DisplayName,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
