package job_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/leonelortega/cards-reminder-api/internal/domain"
	"github.com/leonelortega/cards-reminder-api/internal/job"
	ni18n "github.com/leonelortega/cards-reminder-api/internal/notification/i18n"
)

func TestPickReminderBatches_separateOverdueAndUrgent(t *testing.T) {
	cardA := domain.Card{ID: uuid.New(), Name: "Visa", LastFourDigits: "1234"}
	cardB := domain.Card{ID: uuid.New(), Name: "Amex", LastFourDigits: "5678"}

	items := []domain.DashboardItem{
		{Card: cardA, Status: domain.CardStatusInfo{Status: domain.CardStatusOverdue, DaysOverdue: 5}},
		{Card: cardB, Status: domain.CardStatusInfo{Status: domain.CardStatusUrgent, DaysUntilPayment: 1}},
	}

	batches := job.PickReminderBatches(items, nil)
	if len(batches) != 2 {
		t.Fatalf("got %d batches, want 2", len(batches))
	}
	if batches[0].Kind() != ni18n.ReminderKindOverdue || batches[0].CardCount() != 1 {
		t.Fatalf("expected first batch overdue with 1 card, got kind=%v count=%d", batches[0].Kind(), batches[0].CardCount())
	}
	if batches[1].Kind() != ni18n.ReminderKindUrgent || batches[1].CardCount() != 1 {
		t.Fatalf("expected second batch urgent with 1 card, got kind=%v count=%d", batches[1].Kind(), batches[1].CardCount())
	}
}

func TestPickReminderBatches_skipsPaidAndOnTrack(t *testing.T) {
	card := domain.Card{ID: uuid.New(), Name: "Visa", LastFourDigits: "1234"}

	items := []domain.DashboardItem{
		{Card: card, Status: domain.CardStatusInfo{Status: domain.CardStatusPaid}},
		{Card: card, Status: domain.CardStatusInfo{Status: domain.CardStatusOnTrack}},
	}

	batches := job.PickReminderBatches(items, nil)
	if len(batches) != 0 {
		t.Fatalf("expected no reminders, got %v", batches)
	}
}
