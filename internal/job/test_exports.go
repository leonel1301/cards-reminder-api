//go:build test

package job

import (
	"github.com/google/uuid"
	"github.com/leonelortega/cards-reminder-api/internal/domain"
	ni18n "github.com/leonelortega/cards-reminder-api/internal/notification/i18n"
)

type ReminderBatch = reminderBatch

func PickReminderBatches(items []domain.DashboardItem, ownersByID map[uuid.UUID]domain.Owner) []ReminderBatch {
	return pickReminderBatches(items, ownersByID)
}

func (b ReminderBatch) Kind() ni18n.ReminderKind {
	return b.kind
}

func (b ReminderBatch) CardCount() int {
	return len(b.cards)
}
