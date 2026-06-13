package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PaymentRepository struct {
	pool *pgxpool.Pool
}

func NewPaymentRepository(pool *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{pool: pool}
}

func (r *PaymentRepository) HasPaymentForCycle(ctx context.Context, cardID uuid.UUID, cycleEnd time.Time) (bool, error) {
	const query = `
		SELECT EXISTS(
			SELECT 1 FROM card_payments
			WHERE card_id = $1 AND cycle_end = $2
		)
	`

	var exists bool
	err := r.pool.QueryRow(ctx, query, cardID, truncateDate(cycleEnd)).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check payment: %w", err)
	}

	return exists, nil
}

func (r *PaymentRepository) Create(ctx context.Context, cardID uuid.UUID, cycleEnd time.Time, notes *string) error {
	const query = `
		INSERT INTO card_payments (card_id, cycle_end, notes)
		VALUES ($1, $2, $3)
	`

	_, err := r.pool.Exec(ctx, query, cardID, truncateDate(cycleEnd), notes)
	if err != nil {
		return fmt.Errorf("create payment: %w", err)
	}

	return nil
}

func (r *PaymentRepository) ListByCardID(ctx context.Context, cardID uuid.UUID) ([]PaymentRecord, error) {
	const query = `
		SELECT id, card_id, cycle_end, paid_at, notes
		FROM card_payments
		WHERE card_id = $1
		ORDER BY cycle_end DESC
	`

	rows, err := r.pool.Query(ctx, query, cardID)
	if err != nil {
		return nil, fmt.Errorf("list payments: %w", err)
	}
	defer rows.Close()

	payments := make([]PaymentRecord, 0)
	for rows.Next() {
		var payment PaymentRecord
		if err := rows.Scan(&payment.ID, &payment.CardID, &payment.CycleEnd, &payment.PaidAt, &payment.Notes); err != nil {
			return nil, fmt.Errorf("scan payment: %w", err)
		}
		payments = append(payments, payment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate payments: %w", err)
	}

	return payments, nil
}

func (r *PaymentRepository) GetByCardIDAndCycleEnd(ctx context.Context, cardID uuid.UUID, cycleEnd time.Time) (*PaymentRecord, error) {
	const query = `
		SELECT id, card_id, cycle_end, paid_at, notes
		FROM card_payments
		WHERE card_id = $1 AND cycle_end = $2
	`

	var payment PaymentRecord
	err := r.pool.QueryRow(ctx, query, cardID, truncateDate(cycleEnd)).Scan(
		&payment.ID,
		&payment.CardID,
		&payment.CycleEnd,
		&payment.PaidAt,
		&payment.Notes,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get payment: %w", err)
	}

	return &payment, nil
}

type PaymentRecord struct {
	ID       uuid.UUID
	CardID   uuid.UUID
	CycleEnd time.Time
	PaidAt   time.Time
	Notes    *string
}

func truncateDate(t time.Time) time.Time {
	t = t.UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
