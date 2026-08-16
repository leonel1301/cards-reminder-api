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

type LessonProgressRepository struct {
	pool *pgxpool.Pool
}

func NewLessonProgressRepository(pool *pgxpool.Pool) *LessonProgressRepository {
	return &LessonProgressRepository{pool: pool}
}

func (r *LessonProgressRepository) ListLessonIDsByUserID(ctx context.Context, userID uuid.UUID) ([]string, error) {
	const query = `
		SELECT lesson_id
		FROM user_lesson_progress
		WHERE user_id = $1
		ORDER BY completed_at ASC, lesson_id ASC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list lesson progress: %w", err)
	}
	defer rows.Close()

	ids := make([]string, 0)
	for rows.Next() {
		var lessonID string
		if err := rows.Scan(&lessonID); err != nil {
			return nil, fmt.Errorf("scan lesson progress: %w", err)
		}
		ids = append(ids, lessonID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate lesson progress: %w", err)
	}
	return ids, nil
}

func (r *LessonProgressRepository) MarkCompleted(ctx context.Context, userID uuid.UUID, lessonID string) (*domain.LessonProgress, error) {
	const query = `
		INSERT INTO user_lesson_progress (user_id, lesson_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, lesson_id) DO UPDATE
			SET lesson_id = EXCLUDED.lesson_id
		RETURNING id, user_id, lesson_id, completed_at
	`

	var item domain.LessonProgress
	err := r.pool.QueryRow(ctx, query, userID, lessonID).Scan(
		&item.ID,
		&item.UserID,
		&item.LessonID,
		&item.CompletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("mark lesson completed: %w", err)
	}
	return &item, nil
}

func (r *LessonProgressRepository) UnmarkCompleted(ctx context.Context, userID uuid.UUID, lessonID string) error {
	const query = `
		DELETE FROM user_lesson_progress
		WHERE user_id = $1 AND lesson_id = $2
	`

	_, err := r.pool.Exec(ctx, query, userID, lessonID)
	if err != nil {
		return fmt.Errorf("unmark lesson completed: %w", err)
	}
	return nil
}

func (r *LessonProgressRepository) Get(ctx context.Context, userID uuid.UUID, lessonID string) (*domain.LessonProgress, error) {
	const query = `
		SELECT id, user_id, lesson_id, completed_at
		FROM user_lesson_progress
		WHERE user_id = $1 AND lesson_id = $2
	`

	var item domain.LessonProgress
	err := r.pool.QueryRow(ctx, query, userID, lessonID).Scan(
		&item.ID,
		&item.UserID,
		&item.LessonID,
		&item.CompletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get lesson progress: %w", err)
	}
	return &item, nil
}
