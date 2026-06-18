package job

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/leonelortega/cards-reminder-api/internal/domain"
	"github.com/leonelortega/cards-reminder-api/internal/notification/i18n"
	"github.com/leonelortega/cards-reminder-api/internal/repository"
	"github.com/leonelortega/cards-reminder-api/internal/service"
)

type ReminderJob struct {
	deviceRepo          *repository.DeviceTokenRepository
	ownerRepo           *repository.OwnerRepository
	cardStatusSvc       *service.CardStatusService
	notificationSvc     *service.NotificationService
	defaultTimezone     string
	reminderHour        int
	now                 func() time.Time
}

type reminderBatch struct {
	kind  i18n.ReminderKind
	cards []i18n.CardReminder
}

func NewReminderJob(
	deviceRepo *repository.DeviceTokenRepository,
	ownerRepo *repository.OwnerRepository,
	cardStatusSvc *service.CardStatusService,
	notificationSvc *service.NotificationService,
	defaultTimezone string,
	reminderHour int,
) *ReminderJob {
	if reminderHour < 0 || reminderHour > 23 {
		reminderHour = 8
	}

	return &ReminderJob{
		deviceRepo:      deviceRepo,
		ownerRepo:       ownerRepo,
		cardStatusSvc:   cardStatusSvc,
		notificationSvc: notificationSvc,
		defaultTimezone: service.ResolveTimezone(defaultTimezone, service.DefaultTimezone),
		reminderHour:    reminderHour,
		now:             time.Now,
	}
}

func (j *ReminderJob) Run(ctx context.Context) (*domain.ReminderJobResult, error) {
	result := &domain.ReminderJobResult{}
	now := j.now()

	devices, err := j.deviceRepo.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	dashboardCache := make(map[string]*domain.DashboardResponse)
	ownersCache := make(map[uuid.UUID]map[uuid.UUID]domain.Owner)
	processedUsers := make(map[uuid.UUID]struct{})

	for _, device := range devices {
		timezone := service.ResolveTimezone(device.Timezone, j.defaultTimezone)
		if !service.IsLocalReminderHour(now, timezone, j.reminderHour, j.defaultTimezone) {
			result.DevicesSkippedOutsideHour++
			continue
		}

		if _, ok := processedUsers[device.UserID]; !ok {
			processedUsers[device.UserID] = struct{}{}
			result.UsersProcessed++
		}

		cacheKey := device.UserID.String() + "|" + timezone
		dashboard, ok := dashboardCache[cacheKey]
		if !ok {
			dashboard, err = j.cardStatusSvc.GetDashboard(ctx, device.UserID, timezone)
			if err != nil {
				return nil, err
			}
			dashboardCache[cacheKey] = dashboard
		}

		ownersByID, ok := ownersCache[device.UserID]
		if !ok {
			ownersByID, err = j.loadOwnersByID(ctx, device.UserID)
			if err != nil {
				return nil, err
			}
			ownersCache[device.UserID] = ownersByID
		}

		batches := pickReminderBatches(dashboard.Cards, ownersByID)
		if len(batches) == 0 {
			result.UsersSkipped++
			continue
		}

		result.DevicesNotified++
		for _, batch := range batches {
			notification := i18n.BuildReminderNotification(batch.kind, batch.cards, device.Language)

			if err := j.notificationSvc.SendToDeviceWithCleanup(ctx, device.FCMToken, notification); err != nil {
				log.Printf("reminder send failed user=%s device=%s kind=%s: %v", device.UserID, device.ID, batch.kind, err)
				result.SendFailures++
				continue
			}

			result.NotificationsSent++
			log.Printf("reminder sent user=%s device=%s kind=%s timezone=%s language=%s", device.UserID, device.ID, batch.kind, timezone, device.Language)
		}
	}

	return result, nil
}

// pickReminderBatches returns separate notification batches.
// Order: overdue, urgent, due_soon, optimal_day.
func pickReminderBatches(items []domain.DashboardItem, ownersByID map[uuid.UUID]domain.Owner) []reminderBatch {
	batches := make([]reminderBatch, 0, 4)

	if overdue := filterCards(items, domain.CardStatusOverdue, ownersByID); len(overdue) > 0 {
		batches = append(batches, reminderBatch{kind: i18n.ReminderKindOverdue, cards: overdue})
	}
	if urgent := filterCards(items, domain.CardStatusUrgent, ownersByID); len(urgent) > 0 {
		batches = append(batches, reminderBatch{kind: i18n.ReminderKindUrgent, cards: urgent})
	}
	if dueSoon := filterCards(items, domain.CardStatusDueSoon, ownersByID); len(dueSoon) > 0 {
		batches = append(batches, reminderBatch{kind: i18n.ReminderKindDueSoon, cards: dueSoon})
	}
	if len(batches) == 0 {
		if optimal := filterCards(items, domain.CardStatusOptimalDay, ownersByID); len(optimal) > 0 {
			batches = append(batches, reminderBatch{kind: i18n.ReminderKindOptimalDay, cards: optimal})
		}
	}

	return batches
}

func (j *ReminderJob) loadOwnersByID(ctx context.Context, userID uuid.UUID) (map[uuid.UUID]domain.Owner, error) {
	owners, err := j.ownerRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	ownersByID := make(map[uuid.UUID]domain.Owner, len(owners))
	for _, owner := range owners {
		ownersByID[owner.ID] = owner
	}

	return ownersByID, nil
}

func filterCards(items []domain.DashboardItem, status domain.CardStatusValue, ownersByID map[uuid.UUID]domain.Owner) []i18n.CardReminder {
	result := make([]i18n.CardReminder, 0)
	for _, item := range items {
		if item.Status.Status != status {
			continue
		}

		reminder := i18n.CardReminder{
			Card:   item.Card,
			Status: item.Status,
		}
		if owner, ok := ownersByID[item.Card.OwnerID]; ok {
			ownerCopy := owner
			reminder.Owner = &ownerCopy
		}

		result = append(result, reminder)
	}
	return result
}
