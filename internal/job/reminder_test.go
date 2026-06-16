package job

import (
	"testing"

	"github.com/google/uuid"
	"github.com/leonelortega/cards-reminder-api/internal/domain"
	"github.com/leonelortega/cards-reminder-api/internal/notification/i18n"
)

func TestPickReminderCards_priority(t *testing.T) {
	card := domain.Card{ID: uuid.New(), Name: "Visa", LastFourDigits: "1234"}

	items := []domain.DashboardItem{
		{Card: card, Status: domain.CardStatusInfo{Status: domain.CardStatusOptimalDay}},
		{Card: card, Status: domain.CardStatusInfo{Status: domain.CardStatusDueSoon}},
		{Card: card, Status: domain.CardStatusInfo{Status: domain.CardStatusUrgent}},
	}

	kind, cards := pickReminderCards(items, nil)
	if kind != i18n.ReminderKindUrgent {
		t.Fatalf("got kind %q, want urgent", kind)
	}
	if len(cards) != 1 {
		t.Fatalf("got %d urgent cards, want 1", len(cards))
	}
}

func TestPickReminderCards_skipsPaidAndOnTrack(t *testing.T) {
	card := domain.Card{ID: uuid.New(), Name: "Visa", LastFourDigits: "1234"}

	items := []domain.DashboardItem{
		{Card: card, Status: domain.CardStatusInfo{Status: domain.CardStatusPaid}},
		{Card: card, Status: domain.CardStatusInfo{Status: domain.CardStatusOnTrack}},
	}

	kind, cards := pickReminderCards(items, nil)
	if kind != "" || cards != nil {
		t.Fatalf("expected no reminders, got kind=%q cards=%v", kind, cards)
	}
}
