package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leonelortega/cards-reminder-api/internal/domain"
)

type DeviceTokenRepository struct {
	pool *pgxpool.Pool
}

func NewDeviceTokenRepository(pool *pgxpool.Pool) *DeviceTokenRepository {
	return &DeviceTokenRepository{pool: pool}
}

func (r *DeviceTokenRepository) Upsert(ctx context.Context, userID uuid.UUID, input domain.RegisterDeviceInput) (*domain.DeviceToken, error) {
	const query = `
		INSERT INTO device_tokens (user_id, fcm_token, platform, language, timezone)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (fcm_token) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			platform = EXCLUDED.platform,
			language = EXCLUDED.language,
			timezone = EXCLUDED.timezone,
			updated_at = now()
		RETURNING id, user_id, fcm_token, platform, language, timezone, created_at, updated_at
	`

	var token domain.DeviceToken
	err := r.pool.QueryRow(ctx, query, userID, input.FCMToken, input.Platform, input.Language, input.Timezone).Scan(
		&token.ID,
		&token.UserID,
		&token.FCMToken,
		&token.Platform,
		&token.Language,
		&token.Timezone,
		&token.CreatedAt,
		&token.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("upsert device token: %w", err)
	}

	return &token, nil
}

func (r *DeviceTokenRepository) DeleteByTokenAndUserID(ctx context.Context, userID uuid.UUID, fcmToken string) error {
	const query = `DELETE FROM device_tokens WHERE fcm_token = $1 AND user_id = $2`

	result, err := r.pool.Exec(ctx, query, fcmToken, userID)
	if err != nil {
		return fmt.Errorf("delete device token: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *DeviceTokenRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]domain.DeviceToken, error) {
	const query = `
		SELECT id, user_id, fcm_token, platform, language, timezone, created_at, updated_at
		FROM device_tokens
		WHERE user_id = $1
		ORDER BY updated_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list device tokens: %w", err)
	}
	defer rows.Close()

	return scanDeviceTokens(rows)
}

func (r *DeviceTokenRepository) ListAll(ctx context.Context) ([]domain.DeviceToken, error) {
	const query = `
		SELECT id, user_id, fcm_token, platform, language, timezone, created_at, updated_at
		FROM device_tokens
		ORDER BY user_id, updated_at DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list all device tokens: %w", err)
	}
	defer rows.Close()

	return scanDeviceTokens(rows)
}

func (r *DeviceTokenRepository) GetLatestTimezoneByUserID(ctx context.Context, userID uuid.UUID) (string, error) {
	const query = `
		SELECT timezone
		FROM device_tokens
		WHERE user_id = $1
		ORDER BY updated_at DESC
		LIMIT 1
	`

	var timezone string
	err := r.pool.QueryRow(ctx, query, userID).Scan(&timezone)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil
		}
		return "", fmt.Errorf("get latest timezone: %w", err)
	}

	return timezone, nil
}

func (r *DeviceTokenRepository) DeleteByFCMToken(ctx context.Context, fcmToken string) error {
	const query = `DELETE FROM device_tokens WHERE fcm_token = $1`

	_, err := r.pool.Exec(ctx, query, fcmToken)
	if err != nil {
		return fmt.Errorf("delete device token by fcm token: %w", err)
	}

	return nil
}

func (r *DeviceTokenRepository) ListDistinctUserIDs(ctx context.Context) ([]uuid.UUID, error) {
	const query = `
		SELECT DISTINCT user_id
		FROM device_tokens
		ORDER BY user_id
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list distinct user ids: %w", err)
	}
	defer rows.Close()

	userIDs := make([]uuid.UUID, 0)
	for rows.Next() {
		var userID uuid.UUID
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("scan user id: %w", err)
		}
		userIDs = append(userIDs, userID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate user ids: %w", err)
	}

	return userIDs, nil
}

type deviceTokenRows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}

func scanDeviceTokens(rows deviceTokenRows) ([]domain.DeviceToken, error) {
	tokens := make([]domain.DeviceToken, 0)
	for rows.Next() {
		var token domain.DeviceToken
		if err := rows.Scan(
			&token.ID,
			&token.UserID,
			&token.FCMToken,
			&token.Platform,
			&token.Language,
			&token.Timezone,
			&token.CreatedAt,
			&token.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan device token: %w", err)
		}
		tokens = append(tokens, token)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate device tokens: %w", err)
	}

	return tokens, nil
}
