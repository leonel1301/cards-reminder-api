package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leonelortega/cards-reminder-api/internal/domain"
)

const userSelectColumns = `
	id, firebase_uid, email, display_name,
	terms_accepted_at, privacy_accepted_at, terms_version,
	created_at, updated_at
`

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func scanUser(row interface {
	Scan(dest ...any) error
}) (*domain.User, error) {
	var user domain.User
	err := row.Scan(
		&user.ID,
		&user.FirebaseUID,
		&user.Email,
		&user.DisplayName,
		&user.TermsAcceptedAt,
		&user.PrivacyAcceptedAt,
		&user.TermsVersion,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByFirebaseUID(ctx context.Context, firebaseUID string) (*domain.User, error) {
	query := `
		SELECT ` + userSelectColumns + `
		FROM users
		WHERE firebase_uid = $1
	`

	return scanUser(r.pool.QueryRow(ctx, query, firebaseUID))
}

func (r *UserRepository) Upsert(ctx context.Context, firebaseUID string, email, displayName *string) (*domain.User, error) {
	query := `
		INSERT INTO users (firebase_uid, email, display_name)
		VALUES ($1, $2, $3)
		ON CONFLICT (firebase_uid) DO UPDATE SET
			email = COALESCE(EXCLUDED.email, users.email),
			display_name = COALESCE(EXCLUDED.display_name, users.display_name),
			updated_at = now()
		RETURNING ` + userSelectColumns

	user, err := scanUser(r.pool.QueryRow(ctx, query, firebaseUID, email, displayName))
	if err != nil {
		return nil, fmt.Errorf("upsert user: %w", err)
	}

	return user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT ` + userSelectColumns + `
		FROM users
		WHERE id = $1
	`

	return scanUser(r.pool.QueryRow(ctx, query, id))
}

func (r *UserRepository) AcceptTerms(ctx context.Context, id uuid.UUID, termsVersion string) (*domain.User, error) {
	query := `
		UPDATE users
		SET
			terms_accepted_at = now(),
			privacy_accepted_at = now(),
			terms_version = $2,
			updated_at = now()
		WHERE id = $1
		RETURNING ` + userSelectColumns

	user, err := scanUser(r.pool.QueryRow(ctx, query, id, termsVersion))
	if err != nil {
		return nil, fmt.Errorf("accept terms: %w", err)
	}

	return user, nil
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const query = `DELETE FROM users WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
