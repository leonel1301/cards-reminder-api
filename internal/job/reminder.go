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

		kind, cards := pickReminderCards(dashboard.Cards, ownersByID)
		if kind == "" || len(cards) == 0 {
			result.UsersSkipped++
			continue
		}

		result.DevicesNotified++
		notification := i18n.BuildReminderNotification(kind, cards, device.Language)

		if err := j.notificationSvc.SendToDeviceWithCleanup(ctx, device.FCMToken, notification); err != nil {
			log.Printf("reminder send failed user=%s device=%s: %v", device.UserID, device.ID, err)
			result.SendFailures++
			continue
		}

		result.NotificationsSent++
		log.Printf("reminder sent user=%s device=%s kind=%s timezone=%s language=%s", device.UserID, device.ID, kind, timezone, device.Language)
	}

	return result, nil
}

// pickReminderCards selects which cards to notify and the highest-priority kind.
// Priority: urgent > due_soon > optimal_day.
// paid and on_track cards are never included.
func pickReminderCards(items []domain.DashboardItem, ownersByID map[uuid.UUID]domain.Owner) (i18n.ReminderKind, []i18n.CardReminder) {
	urgent := filterCards(items, domain.CardStatusUrgent, ownersByID)
	if len(urgent) > 0 {
		return i18n.ReminderKindUrgent, urgent
	}

	dueSoon := filterCards(items, domain.CardStatusDueSoon, ownersByID)
	if len(dueSoon) > 0 {
		return i18n.ReminderKindDueSoon, dueSoon
	}

	optimal := filterCards(items, domain.CardStatusOptimalDay, ownersByID)
	if len(optimal) > 0 {
		return i18n.ReminderKindOptimalDay, optimal
	}

	return "", nil
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
