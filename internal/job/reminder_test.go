package job

import (
	"testing"

	"github.com/google/uuid"
	"github.com/leonelortega/cards-reminder-api/internal/domain"
	"github.com/leonelortega/cards-reminder-api/internal/notification/i18n"
)

func TestPickReminderBatches_separateOverdueAndUrgent(t *testing.T) {
	cardA := domain.Card{ID: uuid.New(), Name: "Visa", LastFourDigits: "1234"}
	cardB := domain.Card{ID: uuid.New(), Name: "Amex", LastFourDigits: "5678"}

	items := []domain.DashboardItem{
		{Card: cardA, Status: domain.CardStatusInfo{Status: domain.CardStatusOverdue, DaysOverdue: 5}},
		{Card: cardB, Status: domain.CardStatusInfo{Status: domain.CardStatusUrgent, DaysUntilPayment: 1}},
	}

	batches := pickReminderBatches(items, nil)
	if len(batches) != 2 {
		t.Fatalf("got %d batches, want 2", len(batches))
	}
	if batches[0].kind != i18n.ReminderKindOverdue || len(batches[0].cards) != 1 {
		t.Fatalf("expected first batch overdue with 1 card, got %+v", batches[0])
	}
	if batches[1].kind != i18n.ReminderKindUrgent || len(batches[1].cards) != 1 {
		t.Fatalf("expected second batch urgent with 1 card, got %+v", batches[1])
	}
}

func TestPickReminderBatches_skipsPaidAndOnTrack(t *testing.T) {
	card := domain.Card{ID: uuid.New(), Name: "Visa", LastFourDigits: "1234"}

	items := []domain.DashboardItem{
		{Card: card, Status: domain.CardStatusInfo{Status: domain.CardStatusPaid}},
		{Card: card, Status: domain.CardStatusInfo{Status: domain.CardStatusOnTrack}},
	}

	batches := pickReminderBatches(items, nil)
	if len(batches) != 0 {
		t.Fatalf("expected no reminders, got %v", batches)
	}
}
