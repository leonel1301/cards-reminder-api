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
	now         func() time.Time
}

func NewCardStatusService(cardRepo *repository.CardRepository, paymentRepo *repository.PaymentRepository) *CardStatusService {
	return &CardStatusService{
		cardRepo:    cardRepo,
		paymentRepo: paymentRepo,
		now:         time.Now,
	}
}

func (s *CardStatusService) GetStatus(ctx context.Context, userID, cardID uuid.UUID) (*domain.CardStatusResponse, error) {
	card, err := s.cardRepo.GetByIDAndUserID(ctx, cardID, userID)
	if err != nil {
		return nil, err
	}

	statusInfo, optimalDays, err := s.buildStatusForCard(ctx, *card, s.now())
	if err != nil {
		return nil, err
	}

	return &domain.CardStatusResponse{
		Card:                *card,
		Status:              statusInfo,
		OptimalPurchaseDays: optimalDays,
	}, nil
}

func (s *CardStatusService) GetOptimalPurchaseDays(ctx context.Context, userID, cardID uuid.UUID) (*domain.OptimalPurchaseDaysResponse, error) {
	card, err := s.cardRepo.GetByIDAndUserID(ctx, cardID, userID)
	if err != nil {
		return nil, err
	}

	cycle := ComputeBillingCycle(s.now(), card.BillingCycleDay, card.PaymentDueDay)
	optimalDays := OptimalPurchaseDays(cycle, defaultOptimalWindowDays)

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

func (s *CardStatusService) GetDashboard(ctx context.Context, userID uuid.UUID) (*domain.DashboardResponse, error) {
	cards, err := s.cardRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	now := s.now()
	items := make([]domain.DashboardItem, 0)
	summary := domain.DashboardSummary{}

	for _, card := range cards {
		if !card.IsActive {
			continue
		}

		statusInfo, _, err := s.buildStatusForCard(ctx, card, now)
		if err != nil {
			return nil, err
		}

		items = append(items, domain.DashboardItem{
			Card:   card,
			Status: statusInfo,
		})

		summary.Total++
		switch statusInfo.Status {
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

func (s *CardStatusService) MarkPaid(ctx context.Context, userID, cardID uuid.UUID, notes *string) (*domain.CardStatusResponse, error) {
	card, err := s.cardRepo.GetByIDAndUserID(ctx, cardID, userID)
	if err != nil {
		return nil, err
	}

	cycle := ComputeBillingCycle(s.now(), card.BillingCycleDay, card.PaymentDueDay)
	if err := s.paymentRepo.Create(ctx, cardID, cycle.End, notes); err != nil {
		paid, checkErr := s.paymentRepo.HasPaymentForCycle(ctx, cardID, cycle.End)
		if checkErr == nil && paid {
			return s.GetStatus(ctx, userID, cardID)
		}
		return nil, err
	}

	return s.GetStatus(ctx, userID, cardID)
}

func (s *CardStatusService) GetCurrentCycle(ctx context.Context, userID, cardID uuid.UUID) (*domain.CurrentCycleResponse, error) {
	card, err := s.cardRepo.GetByIDAndUserID(ctx, cardID, userID)
	if err != nil {
		return nil, err
	}

	now := s.now()
	cycle := ComputeBillingCycle(now, card.BillingCycleDay, card.PaymentDueDay)
	statusInfo, _, err := s.buildStatusForCard(ctx, *card, now)
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

func (s *CardStatusService) buildStatusForCard(ctx context.Context, card domain.Card, ref time.Time) (domain.CardStatusInfo, []time.Time, error) {
	cycle := ComputeBillingCycle(ref, card.BillingCycleDay, card.PaymentDueDay)
	paid, err := s.paymentRepo.HasPaymentForCycle(ctx, card.ID, cycle.End)
	if err != nil {
		return domain.CardStatusInfo{}, nil, err
	}

	statusInfo := BuildCardStatusInfo(ref, cycle, paid)
	optimalDays := OptimalPurchaseDays(cycle, defaultOptimalWindowDays)
	return statusInfo, optimalDays, nil
}
