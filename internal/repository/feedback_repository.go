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

type FeedbackRepository struct {
	pool *pgxpool.Pool
}

func NewFeedbackRepository(pool *pgxpool.Pool) *FeedbackRepository {
	return &FeedbackRepository{pool: pool}
}

func (r *FeedbackRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Feedback, error) {
	const query = `
		SELECT id, user_id, title, device, content, created_at, updated_at
		FROM feedback
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list feedback: %w", err)
	}
	defer rows.Close()

	items := make([]domain.Feedback, 0)
	for rows.Next() {
		item, err := scanFeedback(rows)
		if err != nil {
			return nil, fmt.Errorf("scan feedback: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate feedback: %w", err)
	}

	return items, nil
}

func (r *FeedbackRepository) ListAllWithUserName(ctx context.Context) ([]domain.FeedbackAdminItem, error) {
	const query = `
		SELECT
			f.id,
			COALESCE(NULLIF(u.display_name, ''), NULLIF(u.email, ''), 'Sin nombre') AS user_name,
			f.title,
			f.device,
			f.content,
			f.created_at,
			f.updated_at
		FROM feedback f
		INNER JOIN users u ON u.id = f.user_id
		ORDER BY f.created_at DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list feedback for admin: %w", err)
	}
	defer rows.Close()

	items := make([]domain.FeedbackAdminItem, 0)
	for rows.Next() {
		var item domain.FeedbackAdminItem
		if err := rows.Scan(
			&item.ID,
			&item.UserName,
			&item.Title,
			&item.Device,
			&item.Content,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan feedback for admin: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate feedback for admin: %w", err)
	}

	return items, nil
}

func (r *FeedbackRepository) GetByIDAndUserID(ctx context.Context, feedbackID, userID uuid.UUID) (*domain.Feedback, error) {
	const query = `
		SELECT id, user_id, title, device, content, created_at, updated_at
		FROM feedback
		WHERE id = $1 AND user_id = $2
	`

	row := r.pool.QueryRow(ctx, query, feedbackID, userID)
	item, err := scanFeedback(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get feedback: %w", err)
	}

	return &item, nil
}

func (r *FeedbackRepository) Create(ctx context.Context, userID uuid.UUID, input domain.CreateFeedbackInput) (*domain.Feedback, error) {
	const query = `
		INSERT INTO feedback (user_id, title, device, content)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, title, device, content, created_at, updated_at
	`

	row := r.pool.QueryRow(ctx, query, userID, input.Title, input.Device, input.Content)
	item, err := scanFeedback(row)
	if err != nil {
		return nil, fmt.Errorf("create feedback: %w", err)
	}

	return &item, nil
}

func (r *FeedbackRepository) Update(ctx context.Context, feedbackID, userID uuid.UUID, input domain.UpdateFeedbackInput) (*domain.Feedback, error) {
	setClauses := make([]string, 0, 4)
	args := make([]any, 0, 6)
	argIndex := 1

	addField := func(column string, value any) {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", column, argIndex))
		args = append(args, value)
		argIndex++
	}

	if input.Title != nil {
		addField("title", *input.Title)
	}
	if input.Device != nil {
		addField("device", *input.Device)
	}
	if input.Content != nil {
		addField("content", *input.Content)
	}

	if len(setClauses) == 0 {
		return r.GetByIDAndUserID(ctx, feedbackID, userID)
	}

	setClauses = append(setClauses, "updated_at = now()")
	args = append(args, feedbackID, userID)

	query := fmt.Sprintf(`
		UPDATE feedback
		SET %s
		WHERE id = $%d AND user_id = $%d
		RETURNING id, user_id, title, device, content, created_at, updated_at
	`, strings.Join(setClauses, ", "), argIndex, argIndex+1)

	row := r.pool.QueryRow(ctx, query, args...)
	item, err := scanFeedback(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update feedback: %w", err)
	}

	return &item, nil
}

func (r *FeedbackRepository) Delete(ctx context.Context, feedbackID, userID uuid.UUID) error {
	const query = `DELETE FROM feedback WHERE id = $1 AND user_id = $2`

	result, err := r.pool.Exec(ctx, query, feedbackID, userID)
	if err != nil {
		return fmt.Errorf("delete feedback: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func scanFeedback(row scannable) (domain.Feedback, error) {
	var item domain.Feedback
	err := row.Scan(
		&item.ID,
		&item.UserID,
		&item.Title,
		&item.Device,
		&item.Content,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	return item, err
}
