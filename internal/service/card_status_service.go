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
	optimalDay := ComputeOptimalPurchaseDay(card.BillingCycleDay)
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

func (s *CardStatusService) GetDashboard(ctx context.Context, userID uuid.UUID, timezone, language string) (*domain.DashboardResponse, error) {
	cards, err := s.cardRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	ownersByID, err := s.loadOwnersByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	loc := ResolveLocation(timezone)
	now := s.now()
	items := make([]domain.DashboardItem, 0)
	purchaseCandidates := make([]purchaseCandidate, 0)
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

		var salaryDay *int
		if owner, ok := ownersByID[card.OwnerID]; ok {
			salaryDay = owner.SalaryDay
		}
		purchaseCandidates = append(purchaseCandidates, buildPurchaseCandidate(card, statusInfo, now, salaryDay, loc))

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
		Cards:           items,
		Summary:         summary,
		BestForPurchase: RecommendBestForPurchase(purchaseCandidates, now, loc, language),
	}, nil
}

func (s *CardStatusService) MarkPaid(ctx context.Context, userID, cardID uuid.UUID, notes *string, timezone string) (*domain.CardStatusResponse, error) {
	card, err := s.cardRepo.GetByIDAndUserID(ctx, cardID, userID)
	if err != nil {
		return nil, err
	}

	loc := ResolveLocation(timezone)
	obligationCycle, _, _, err := s.resolvePaymentObligation(ctx, cardID, s.now(), card.BillingCycleDay, card.PaymentDueDay, loc)
	if err != nil {
		return nil, err
	}
	if err := s.paymentRepo.Create(ctx, cardID, obligationCycle.End, notes); err != nil {
		paid, checkErr := s.paymentRepo.HasPaymentForCycle(ctx, cardID, obligationCycle.End)
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
	obligationCycle, paymentDue, paid, err := s.resolvePaymentObligation(ctx, card.ID, ref, card.BillingCycleDay, card.PaymentDueDay, loc)
	if err != nil {
		return domain.CardStatusInfo{}, nil, err
	}

	statusInfo := BuildCardStatusInfo(ref, obligationCycle, paymentDue, card.BillingCycleDay, paid, loc)
	optimalDays := OptimalPurchaseDaysInMonth(ref, statusInfo.OptimalPurchaseDay, defaultOptimalWindowDays, loc)
	return statusInfo, optimalDays, nil
}

func (s *CardStatusService) resolvePaymentObligation(
	ctx context.Context,
	cardID uuid.UUID,
	ref time.Time,
	closingDay, paymentDueDay int,
	loc *time.Location,
) (domain.BillingCycle, time.Time, bool, error) {
	currentCycle := ComputeBillingCycle(ref, closingDay, paymentDueDay, loc)
	prevCycle := PreviousBillingCycle(currentCycle, closingDay, loc)

	paidPrev, err := s.paymentRepo.HasPaymentForCycle(ctx, cardID, prevCycle.End)
	if err != nil {
		return domain.BillingCycle{}, time.Time{}, false, err
	}

	if !paidPrev {
		return prevCycle, PaymentDueForCycleEnd(prevCycle.End, closingDay, paymentDueDay, loc), false, nil
	}

	paidCurrent, err := s.paymentRepo.HasPaymentForCycle(ctx, cardID, currentCycle.End)
	if err != nil {
		return domain.BillingCycle{}, time.Time{}, false, err
	}

	return currentCycle, currentCycle.PaymentDue, paidCurrent, nil
}

func (s *CardStatusService) ownerSalaryDay(ctx context.Context, ownerID, userID uuid.UUID) *int {
	owner, err := s.ownerRepo.GetByIDAndUserID(ctx, ownerID, userID)
	if err != nil {
		return nil
	}
	return owner.SalaryDay
}

func (s *CardStatusService) loadOwnersByID(ctx context.Context, userID uuid.UUID) (map[uuid.UUID]domain.Owner, error) {
	owners, err := s.ownerRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	ownersByID := make(map[uuid.UUID]domain.Owner, len(owners))
	for _, owner := range owners {
		ownersByID[owner.ID] = owner
	}

	return ownersByID, nil
}
