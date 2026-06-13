package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leonelortega/cards-reminder-api/internal/domain"
)

type OwnerRepository struct {
	pool *pgxpool.Pool
}

func NewOwnerRepository(pool *pgxpool.Pool) *OwnerRepository {
	return &OwnerRepository{pool: pool}
}

func (r *OwnerRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Owner, error) {
	const query = `
		SELECT id, user_id, name, salary_day, is_self, created_at, updated_at
		FROM owners
		WHERE user_id = $1
		ORDER BY is_self DESC, name ASC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list owners: %w", err)
	}
	defer rows.Close()

	owners := make([]domain.Owner, 0)
	for rows.Next() {
		owner, err := scanOwner(rows)
		if err != nil {
			return nil, fmt.Errorf("scan owner: %w", err)
		}
		owners = append(owners, owner)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate owners: %w", err)
	}

	return owners, nil
}

func (r *OwnerRepository) GetByIDAndUserID(ctx context.Context, ownerID, userID uuid.UUID) (*domain.Owner, error) {
	const query = `
		SELECT id, user_id, name, salary_day, is_self, created_at, updated_at
		FROM owners
		WHERE id = $1 AND user_id = $2
	`

	row := r.pool.QueryRow(ctx, query, ownerID, userID)
	owner, err := scanOwner(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get owner: %w", err)
	}

	return &owner, nil
}

func (r *OwnerRepository) GetSelfByUserID(ctx context.Context, userID uuid.UUID) (*domain.Owner, error) {
	const query = `
		SELECT id, user_id, name, salary_day, is_self, created_at, updated_at
		FROM owners
		WHERE user_id = $1 AND is_self = true
	`

	row := r.pool.QueryRow(ctx, query, userID)
	owner, err := scanOwner(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get self owner: %w", err)
	}

	return &owner, nil
}

func (r *OwnerRepository) EnsureSelfOwner(ctx context.Context, userID uuid.UUID, displayName *string) error {
	name := "Yo"
	if displayName != nil && *displayName != "" {
		name = *displayName
	}

	const insertQuery = `
		INSERT INTO owners (user_id, name, is_self)
		SELECT $1, $2, true
		WHERE NOT EXISTS (
			SELECT 1 FROM owners WHERE user_id = $1 AND is_self = true
		)
	`
	if _, err := r.pool.Exec(ctx, insertQuery, userID, name); err != nil {
		return fmt.Errorf("ensure self owner insert: %w", err)
	}

	const updateQuery = `
		UPDATE owners
		SET name = $2, updated_at = now()
		WHERE user_id = $1 AND is_self = true
	`
	if _, err := r.pool.Exec(ctx, updateQuery, userID, name); err != nil {
		return fmt.Errorf("ensure self owner update: %w", err)
	}

	return nil
}

func (r *OwnerRepository) Create(ctx context.Context, userID uuid.UUID, input domain.CreateOwnerInput) (*domain.Owner, error) {
	const query = `
		INSERT INTO owners (user_id, name, salary_day, is_self)
		VALUES ($1, $2, $3, false)
		RETURNING id, user_id, name, salary_day, is_self, created_at, updated_at
	`

	row := r.pool.QueryRow(ctx, query, userID, input.Name, input.SalaryDay)
	owner, err := scanOwner(row)
	if err != nil {
		return nil, fmt.Errorf("create owner: %w", err)
	}

	return &owner, nil
}

func (r *OwnerRepository) Update(ctx context.Context, ownerID, userID uuid.UUID, input domain.UpdateOwnerInput) (*domain.Owner, error) {
	setClauses := make([]string, 0, 3)
	args := make([]any, 0, 5)
	argIndex := 1

	addField := func(column string, value any) {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", column, argIndex))
		args = append(args, value)
		argIndex++
	}

	if input.Name != nil {
		addField("name", *input.Name)
	}
	if input.SalaryDay != nil {
		addField("salary_day", *input.SalaryDay)
	}

	if len(setClauses) == 0 {
		return r.GetByIDAndUserID(ctx, ownerID, userID)
	}

	setClauses = append(setClauses, "updated_at = now()")
	args = append(args, ownerID, userID)

	query := fmt.Sprintf(`
		UPDATE owners
		SET %s
		WHERE id = $%d AND user_id = $%d
		RETURNING id, user_id, name, salary_day, is_self, created_at, updated_at
	`, strings.Join(setClauses, ", "), argIndex, argIndex+1)

	row := r.pool.QueryRow(ctx, query, args...)
	owner, err := scanOwner(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update owner: %w", err)
	}

	return &owner, nil
}

func (r *OwnerRepository) Delete(ctx context.Context, ownerID, userID uuid.UUID) error {
	const query = `
		DELETE FROM owners
		WHERE id = $1 AND user_id = $2 AND is_self = false
	`

	result, err := r.pool.Exec(ctx, query, ownerID, userID)
	if err != nil {
		return fmt.Errorf("delete owner: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *OwnerRepository) CountCards(ctx context.Context, ownerID uuid.UUID) (int, error) {
	const query = `SELECT COUNT(*) FROM cards WHERE owner_id = $1`

	var count int
	if err := r.pool.QueryRow(ctx, query, ownerID).Scan(&count); err != nil {
		return 0, fmt.Errorf("count cards for owner: %w", err)
	}

	return count, nil
}

func scanOwner(row scannable) (domain.Owner, error) {
	var owner domain.Owner
	err := row.Scan(
		&owner.ID,
		&owner.UserID,
		&owner.Name,
		&owner.SalaryDay,
		&owner.IsSelf,
		&owner.CreatedAt,
		&owner.UpdatedAt,
	)
	return owner, err
}
