package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrContractAnalyzeLimitReached = errors.New("contract analyze limit reached")

type ContractUsageRepository struct {
	pool *pgxpool.Pool
}

func NewContractUsageRepository(pool *pgxpool.Pool) *ContractUsageRepository {
	return &ContractUsageRepository{pool: pool}
}

func (r *ContractUsageRepository) GetAnalyzeCount(ctx context.Context, userID uuid.UUID) (int, error) {
	const query = `
		SELECT analyze_count
		FROM user_contract_usage
		WHERE user_id = $1
	`

	var count int
	err := r.pool.QueryRow(ctx, query, userID).Scan(&count)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("get contract usage: %w", err)
	}
	return count, nil
}

// TryConsume increments usage when under the limit. Returns the new count.
func (r *ContractUsageRepository) TryConsume(ctx context.Context, userID uuid.UUID, limit int) (int, error) {
	const query = `
		INSERT INTO user_contract_usage (user_id, analyze_count)
		VALUES ($1, 1)
		ON CONFLICT (user_id) DO UPDATE
			SET analyze_count = user_contract_usage.analyze_count + 1,
			    updated_at = now()
		WHERE user_contract_usage.analyze_count < $2
		RETURNING analyze_count
	`

	var count int
	err := r.pool.QueryRow(ctx, query, userID, limit).Scan(&count)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrContractAnalyzeLimitReached
	}
	if err != nil {
		return 0, fmt.Errorf("consume contract usage: %w", err)
	}
	return count, nil
}

func (r *ContractUsageRepository) Release(ctx context.Context, userID uuid.UUID) error {
	const query = `
		UPDATE user_contract_usage
		SET analyze_count = GREATEST(analyze_count - 1, 0),
		    updated_at = now()
		WHERE user_id = $1
	`

	_, err := r.pool.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("release contract usage: %w", err)
	}
	return nil
}
