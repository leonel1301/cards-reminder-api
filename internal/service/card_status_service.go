package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/leonelortega/cards-reminder-api/internal/domain"
	"github.com/leonelortega/cards-reminder-api/internal/repository"
)

type CardStatusService struct {
	cardRepo    *repository.CardRepository
	paymentRepo *repository.PaymentRepository
	ownerRepo   *repository.OwnerRepository
	now         func() time.Time
}

func NewCardStatusService(
	cardRepo *repository.CardRepository,
	paymentRepo *repository.PaymentRepository,
	ownerRepo *repository.OwnerRepository,
) *CardStatusService {
	return &CardStatusService{
		cardRepo:    cardRepo,
		paymentRepo: paymentRepo,
		ownerRepo:   ownerRepo,
		now:         time.Now,
	}
}

func (s *CardStatusService) GetStatus(ctx context.Context, userID, cardID uuid.UUID, timezone string) (*domain.CardStatusResponse, error) {
	card, err := s.cardRepo.GetByIDAndUserID(ctx, cardID, userID)
	if err != nil {
		return nil, err
	}

	loc := ResolveLocation(timezone)
	statusInfo, optimalDays, err := s.buildStatusForCard(ctx, *card, s.now(), loc)
	if err != nil {
		return nil, err
	}

	return &domain.CardStatusResponse{
		Card:                *card,
		Status:              statusInfo,
		OptimalPurchaseDays: optimalDays,
	}, nil
}

func (s *CardStatusService) GetOptimalPurchaseDays(ctx context.Context, userID, cardID uuid.UUID, timezone string) (*domain.OptimalPurchaseDaysResponse, error) {
	card, err := s.cardRepo.GetByIDAndUserID(ctx, cardID, userID)
	if err != nil {
		return nil, err
	}

	loc := ResolveLocation(timezone)
	now := s.now()
	cycle := ComputeBillingCycle(now, card.BillingCycleDay, card.PaymentDueDay, loc)
	salaryDay := s.ownerSalaryDay(ctx, card.OwnerID, card.UserID)
	optimalDay := ComputeOptimalPurchaseDay(card.BillingCycleDay, salaryDay)
	optimalDays := OptimalPurchaseDaysInMonth(now, optimalDay, defaultOptimalWindowDays, loc)

	return &domain.OptimalPurchaseDaysResponse{
		Card: *card,
		Cycle: domain.BillingCycleDates{
			Start:      cycle.Start,
			End:        cycle.End,
			PaymentDue: cycle.PaymentDue,
		},
		OptimalPurchaseDays: optimalDays,
	}, nil
}

func (s *CardStatusService) GetDashboard(ctx context.Context, userID uuid.UUID, timezone string) (*domain.DashboardResponse, error) {
	cards, err := s.cardRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	loc := ResolveLocation(timezone)
	now := s.now()
	items := make([]domain.DashboardItem, 0)
	summary := domain.DashboardSummary{}

	for _, card := range cards {
		if !card.IsActive {
			continue
		}

		statusInfo, _, err := s.buildStatusForCard(ctx, card, now, loc)
		if err != nil {
			return nil, err
		}

		items = append(items, domain.DashboardItem{
			Card:   card,
			Status: statusInfo,
		})

		summary.Total++
		switch statusInfo.Status {
		case domain.CardStatusOverdue:
			summary.Overdue++
		case domain.CardStatusUrgent:
			summary.Urgent++
		case domain.CardStatusDueSoon:
			summary.DueSoon++
		case domain.CardStatusPaid:
			summary.Paid++
		case domain.CardStatusOptimalDay:
			summary.OptimalDay++
		case domain.CardStatusOnTrack:
			summary.OnTrack++
		}
	}

	if items == nil {
		items = []domain.DashboardItem{}
	}

	return &domain.DashboardResponse{
		Cards:   items,
		Summary: summary,
	}, nil
}

func (s *CardStatusService) MarkPaid(ctx context.Context, userID, cardID uuid.UUID, notes *string, timezone string) (*domain.CardStatusResponse, error) {
	card, err := s.cardRepo.GetByIDAndUserID(ctx, cardID, userID)
	if err != nil {
		return nil, err
	}

	loc := ResolveLocation(timezone)
	cycle := ComputeBillingCycle(s.now(), card.BillingCycleDay, card.PaymentDueDay, loc)
	if err := s.paymentRepo.Create(ctx, cardID, cycle.End, notes); err != nil {
		paid, checkErr := s.paymentRepo.HasPaymentForCycle(ctx, cardID, cycle.End)
		if checkErr == nil && paid {
			return s.GetStatus(ctx, userID, cardID, timezone)
		}
		return nil, err
	}

	return s.GetStatus(ctx, userID, cardID, timezone)
}

func (s *CardStatusService) GetCurrentCycle(ctx context.Context, userID, cardID uuid.UUID, timezone string) (*domain.CurrentCycleResponse, error) {
	card, err := s.cardRepo.GetByIDAndUserID(ctx, cardID, userID)
	if err != nil {
		return nil, err
	}

	loc := ResolveLocation(timezone)
	now := s.now()
	cycle := ComputeBillingCycle(now, card.BillingCycleDay, card.PaymentDueDay, loc)
	statusInfo, _, err := s.buildStatusForCard(ctx, *card, now, loc)
	if err != nil {
		return nil, err
	}

	return &domain.CurrentCycleResponse{
		Card: *card,
		Cycle: domain.BillingCycleDates{
			Start:      cycle.Start,
			End:        cycle.End,
			PaymentDue: cycle.PaymentDue,
		},
		Status: statusInfo,
	}, nil
}

func (s *CardStatusService) ListPayments(ctx context.Context, userID, cardID uuid.UUID) (*domain.PaymentsResponse, error) {
	card, err := s.cardRepo.GetByIDAndUserID(ctx, cardID, userID)
	if err != nil {
		return nil, err
	}

	records, err := s.paymentRepo.ListByCardID(ctx, cardID)
	if err != nil {
		return nil, err
	}

	payments := make([]domain.Payment, 0, len(records))
	for _, record := range records {
		payments = append(payments, domain.Payment{
			ID:       record.ID,
			CardID:   record.CardID,
			CycleEnd: record.CycleEnd,
			PaidAt:   record.PaidAt,
			Notes:    record.Notes,
		})
	}
	if payments == nil {
		payments = []domain.Payment{}
	}

	return &domain.PaymentsResponse{
		Card:     *card,
		Payments: payments,
	}, nil
}

func (s *CardStatusService) buildStatusForCard(ctx context.Context, card domain.Card, ref time.Time, loc *time.Location) (domain.CardStatusInfo, []time.Time, error) {
	cycle := ComputeBillingCycle(ref, card.BillingCycleDay, card.PaymentDueDay, loc)
	paid, err := s.paymentRepo.HasPaymentForCycle(ctx, card.ID, cycle.End)
	if err != nil {
		return domain.CardStatusInfo{}, nil, err
	}

	salaryDay := s.ownerSalaryDay(ctx, card.OwnerID, card.UserID)
	statusInfo := BuildCardStatusInfo(ref, cycle, card.PaymentDueDay, card.BillingCycleDay, salaryDay, paid, loc)
	optimalDays := OptimalPurchaseDaysInMonth(ref, statusInfo.OptimalPurchaseDay, defaultOptimalWindowDays, loc)
	return statusInfo, optimalDays, nil
}

func (s *CardStatusService) ownerSalaryDay(ctx context.Context, ownerID, userID uuid.UUID) *int {
	owner, err := s.ownerRepo.GetByIDAndUserID(ctx, ownerID, userID)
	if err != nil {
		return nil
	}
	return owner.SalaryDay
}
