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
