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

type CardRepository struct {
	pool *pgxpool.Pool
}

func NewCardRepository(pool *pgxpool.Pool) *CardRepository {
	return &CardRepository{pool: pool}
}

const cardSelectColumns = `
	id, user_id, owner_id, name, last_four_digits, issuer, billing_cycle_day,
	payment_due_day, color_hex, notes, is_active, created_at, updated_at
`

func (r *CardRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Card, error) {
	query := `
		SELECT ` + cardSelectColumns + `
		FROM cards
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list cards: %w", err)
	}
	defer rows.Close()

	cards := make([]domain.Card, 0)
	for rows.Next() {
		card, err := scanCard(rows)
		if err != nil {
			return nil, fmt.Errorf("scan card: %w", err)
		}
		cards = append(cards, card)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate cards: %w", err)
	}

	return cards, nil
}

func (r *CardRepository) GetByIDAndUserID(ctx context.Context, cardID, userID uuid.UUID) (*domain.Card, error) {
	query := `
		SELECT ` + cardSelectColumns + `
		FROM cards
		WHERE id = $1 AND user_id = $2
	`

	row := r.pool.QueryRow(ctx, query, cardID, userID)
	card, err := scanCard(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get card: %w", err)
	}

	return &card, nil
}

func (r *CardRepository) Create(ctx context.Context, userID uuid.UUID, input domain.CreateCardInput) (*domain.Card, error) {
	query := `
		INSERT INTO cards (
			user_id, owner_id, name, last_four_digits, issuer,
			billing_cycle_day, payment_due_day, color_hex, notes
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING ` + cardSelectColumns

	row := r.pool.QueryRow(
		ctx,
		query,
		userID,
		*input.OwnerID,
		input.Name,
		input.LastFourDigits,
		input.Issuer,
		input.BillingCycleDay,
		input.PaymentDueDay,
		input.ColorHex,
		input.Notes,
	)

	card, err := scanCard(row)
	if err != nil {
		return nil, fmt.Errorf("create card: %w", err)
	}

	return &card, nil
}

func (r *CardRepository) Update(ctx context.Context, cardID, userID uuid.UUID, input domain.UpdateCardInput) (*domain.Card, error) {
	setClauses := make([]string, 0, 9)
	args := make([]any, 0, 11)
	argIndex := 1

	addField := func(column string, value any) {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", column, argIndex))
		args = append(args, value)
		argIndex++
	}

	if input.Name != nil {
		addField("name", *input.Name)
	}
	if input.LastFourDigits != nil {
		addField("last_four_digits", *input.LastFourDigits)
	}
	if input.Issuer != nil {
		addField("issuer", *input.Issuer)
	}
	if input.BillingCycleDay != nil {
		addField("billing_cycle_day", *input.BillingCycleDay)
	}
	if input.PaymentDueDay != nil {
		addField("payment_due_day", *input.PaymentDueDay)
	}
	if input.ColorHex != nil {
		addField("color_hex", *input.ColorHex)
	}
	if input.Notes != nil {
		addField("notes", *input.Notes)
	}
	if input.IsActive != nil {
		addField("is_active", *input.IsActive)
	}
	if input.OwnerID != nil {
		addField("owner_id", *input.OwnerID)
	}

	if len(setClauses) == 0 {
		return r.GetByIDAndUserID(ctx, cardID, userID)
	}

	setClauses = append(setClauses, "updated_at = now()")

	args = append(args, cardID, userID)
	query := fmt.Sprintf(`
		UPDATE cards
		SET %s
		WHERE id = $%d AND user_id = $%d
		RETURNING %s
	`, strings.Join(setClauses, ", "), argIndex, argIndex+1, cardSelectColumns)

	row := r.pool.QueryRow(ctx, query, args...)
	card, err := scanCard(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update card: %w", err)
	}

	return &card, nil
}

func (r *CardRepository) Delete(ctx context.Context, cardID, userID uuid.UUID) error {
	const query = `DELETE FROM cards WHERE id = $1 AND user_id = $2`

	result, err := r.pool.Exec(ctx, query, cardID, userID)
	if err != nil {
		return fmt.Errorf("delete card: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func scanCard(row scannable) (domain.Card, error) {
	var card domain.Card
	err := row.Scan(
		&card.ID,
		&card.UserID,
		&card.OwnerID,
		&card.Name,
		&card.LastFourDigits,
		&card.Issuer,
		&card.BillingCycleDay,
		&card.PaymentDueDay,
		&card.ColorHex,
		&card.Notes,
		&card.IsActive,
		&card.CreatedAt,
		&card.UpdatedAt,
	)
	return card, err
}
