package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
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
		INSERT INTO device_tokens (user_id, fcm_token, platform, language)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (fcm_token) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			platform = EXCLUDED.platform,
			language = EXCLUDED.language,
			updated_at = now()
		RETURNING id, user_id, fcm_token, platform, language, created_at, updated_at
	`

	var token domain.DeviceToken
	err := r.pool.QueryRow(ctx, query, userID, input.FCMToken, input.Platform, input.Language).Scan(
		&token.ID,
		&token.UserID,
		&token.FCMToken,
		&token.Platform,
		&token.Language,
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
		SELECT id, user_id, fcm_token, platform, language, created_at, updated_at
		FROM device_tokens
		WHERE user_id = $1
		ORDER BY updated_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list device tokens: %w", err)
	}
	defer rows.Close()

	tokens := make([]domain.DeviceToken, 0)
	for rows.Next() {
		var token domain.DeviceToken
		if err := rows.Scan(
			&token.ID,
			&token.UserID,
			&token.FCMToken,
			&token.Platform,
			&token.Language,
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

func (r *DeviceTokenRepository) DeleteByFCMToken(ctx context.Context, fcmToken string) error {
	const query = `DELETE FROM device_tokens WHERE fcm_token = $1`

	_, err := r.pool.Exec(ctx, query, fcmToken)
	if err != nil {
		return fmt.Errorf("delete device token by fcm token: %w", err)
	}

	return nil
}
