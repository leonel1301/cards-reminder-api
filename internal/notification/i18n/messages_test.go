package i18n

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/leonelortega/cards-reminder-api/internal/domain"
)

func TestBuildReminderNotification_urgentSpanish(t *testing.T) {
	card := domain.Card{
		ID:             uuid.New(),
		Name:           "Visa Banco X",
		LastFourDigits: "4532",
	}

	notification := BuildReminderNotification(ReminderKindUrgent, []CardReminder{
		{
			Card: card,
			Status: domain.CardStatusInfo{
				DaysUntilPayment: 1,
			},
		},
	}, "es")

	if notification.Title != "Pago urgente" {
		t.Fatalf("unexpected title: %s", notification.Title)
	}
	if notification.Data["kind"] != "urgent" {
		t.Fatalf("unexpected kind: %s", notification.Data["kind"])
	}
}

func TestBuildReminderNotification_dueSoonEnglish(t *testing.T) {
	card := domain.Card{
		ID:             uuid.New(),
		Name:           "Amex",
		LastFourDigits: "9911",
	}

	notification := BuildReminderNotification(ReminderKindDueSoon, []CardReminder{
		{
			Card: card,
			Status: domain.CardStatusInfo{
				DaysUntilPayment: 5,
			},
		},
	}, "en-US")

	if notification.Title != "Upcoming payment" {
		t.Fatalf("unexpected title: %s", notification.Title)
	}
}

func TestBuildReminderNotification_includesOwnerName(t *testing.T) {
	ownerID := uuid.New()
	card := domain.Card{
		ID:             uuid.New(),
		OwnerID:        ownerID,
		Name:           "Visa",
		LastFourDigits: "4532",
	}
	owner := domain.Owner{
		ID:     ownerID,
		Name:   "María",
		IsSelf: false,
	}

	notification := BuildReminderNotification(ReminderKindUrgent, []CardReminder{
		{
			Card:   card,
			Status: domain.CardStatusInfo{DaysUntilPayment: 1},
			Owner:  &owner,
		},
	}, "es")

	if !strings.Contains(notification.Body, "Visa de María") {
		t.Fatalf("expected owner in body, got: %s", notification.Body)
	}
	if notification.Data["owner_name"] != "María" {
		t.Fatalf("expected owner_name in data, got: %s", notification.Data["owner_name"])
	}
}

func TestBuildReminderNotification_overdueIncludesOwner(t *testing.T) {
	ownerID := uuid.New()
	card := domain.Card{
		ID:             uuid.New(),
		OwnerID:        ownerID,
		Name:           "Visa",
		LastFourDigits: "4532",
	}
	owner := domain.Owner{
		ID:     ownerID,
		Name:   "María",
		IsSelf: false,
	}

	notification := BuildReminderNotification(ReminderKindOverdue, []CardReminder{
		{
			Card:   card,
			Status: domain.CardStatusInfo{DaysOverdue: 5},
			Owner:  &owner,
		},
	}, "es")

	if !strings.Contains(notification.Body, "Visa de María") {
		t.Fatalf("expected owner in overdue body, got: %s", notification.Body)
	}
	if notification.Data["kind"] != "overdue" {
		t.Fatalf("unexpected kind: %s", notification.Data["kind"])
	}
}

func TestBuildReminderNotification_selfOwnerOmitsOwnerName(t *testing.T) {
	ownerID := uuid.New()
	card := domain.Card{
		ID:             uuid.New(),
		OwnerID:        ownerID,
		Name:           "Visa",
		LastFourDigits: "4532",
	}
	owner := domain.Owner{
		ID:     ownerID,
		Name:   "Yo",
		IsSelf: true,
	}

	notification := BuildReminderNotification(ReminderKindDueSoon, []CardReminder{
		{
			Card:   card,
			Status: domain.CardStatusInfo{DaysUntilPayment: 4},
			Owner:  &owner,
		},
	}, "es")

	if strings.Contains(notification.Body, " de Yo") {
		t.Fatalf("self owner should not appear in body: %s", notification.Body)
	}
	if _, ok := notification.Data["owner_name"]; ok {
		t.Fatalf("owner_name should not be set for self owner")
	}
}
