package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/leonelortega/cards-reminder-api/internal/domain"
)

func TestRecommendBestForPurchase_salary30_prefersLongestFinancing(t *testing.T) {
	loc := time.FixedZone("PET", -5*3600)
	ref := time.Date(2026, 6, 18, 12, 0, 0, 0, loc)
	salaryDay := 30

	candidates := []purchaseCandidate{
		buildTestPurchaseCandidate("CMR", 9, 5, salaryDay, ref, loc, true),
		buildTestPurchaseCandidate("BCP", 25, 23, salaryDay, ref, loc, true),
		buildTestPurchaseCandidate("IO", 25, 12, salaryDay, ref, loc, true),
	}

	got := RecommendBestForPurchase(candidates, ref, loc, "es")
	if got == nil {
		t.Fatal("expected recommendation")
	}
	if got.CardID != candidates[0].Card.ID {
		t.Fatalf("expected CMR card id %s, got %s", candidates[0].Card.ID, got.CardID)
	}
	if candidates[0].FinancingDays != 48 {
		t.Fatalf("expected CMR financing days 48, got %d", candidates[0].FinancingDays)
	}
	if candidates[1].FinancingDays != 35 {
		t.Fatalf("expected BCP financing days 35, got %d", candidates[1].FinancingDays)
	}
	if candidates[2].FinancingDays != 24 {
		t.Fatalf("expected IO financing days 24, got %d", candidates[2].FinancingDays)
	}
}

func TestRecommendBestForPurchase_salary15_penalizesPaymentBeforeSalary(t *testing.T) {
	loc := time.FixedZone("PET", -5*3600)
	ref := time.Date(2026, 6, 18, 12, 0, 0, 0, loc)
	salaryDay := 15

	candidates := []purchaseCandidate{
		buildTestPurchaseCandidate("CMR", 9, 5, salaryDay, ref, loc, true),
		buildTestPurchaseCandidate("BCP", 25, 23, salaryDay, ref, loc, true),
		buildTestPurchaseCandidate("IO", 25, 12, salaryDay, ref, loc, true),
	}

	if candidates[2].AlignsWithSalary {
		t.Fatal("expected IO not to align with salary day 15")
	}
	if !candidates[1].AlignsWithSalary {
		t.Fatal("expected BCP to align with salary day 15")
	}
	if !candidates[0].AlignsWithSalary {
		t.Fatal("expected CMR to align with salary day 15")
	}

	got := RecommendBestForPurchase(candidates, ref, loc, "es")
	if got == nil {
		t.Fatal("expected recommendation")
	}
	if got.CardID != candidates[0].Card.ID {
		t.Fatalf("expected CMR card id %s, got %s", candidates[0].Card.ID, got.CardID)
	}
}

func TestNextSalaryDate(t *testing.T) {
	loc := time.UTC

	assertDate(t, NextSalaryDate(time.Date(2026, 6, 18, 0, 0, 0, 0, loc), 30, loc), 2026, 6, 30)
	assertDate(t, NextSalaryDate(time.Date(2026, 6, 18, 0, 0, 0, 0, loc), 15, loc), 2026, 7, 15)
	assertDate(t, NextSalaryDate(time.Date(2026, 6, 18, 0, 0, 0, 0, loc), 10, loc), 2026, 7, 10)
}

func TestBuildBestForPurchaseWhy_pendingDebtMentioned(t *testing.T) {
	loc := time.FixedZone("PET", -5*3600)
	ref := time.Date(2026, 6, 18, 12, 0, 0, 0, loc)
	candidate := buildTestPurchaseCandidate("CMR", 9, 5, 30, ref, loc, false)

	why := buildBestForPurchaseWhy(candidate, truncateToDateInLoc(ref, loc), loc, "es")
	if why == "" {
		t.Fatal("expected why message")
	}
	if !containsAll(why, "pendiente", "ciclo anterior") {
		t.Fatalf("expected pending debt mention, got: %s", why)
	}
}

func buildTestPurchaseCandidate(
	name string,
	closingDay, paymentDay, salaryDay int,
	ref time.Time,
	loc *time.Location,
	paid bool,
) purchaseCandidate {
	card := domain.Card{
		ID:              uuid.New(),
		Name:            name,
		LastFourDigits:  "1234",
		BillingCycleDay: closingDay,
		PaymentDueDay:   paymentDay,
	}
	status := domain.CardStatusInfo{
		IsPaidThisCycle: paid,
	}
	if !paid {
		prevCycle := PreviousBillingCycle(ComputeBillingCycle(ref, closingDay, paymentDay, loc), closingDay, loc)
		status.PaymentDueDate = PaymentDueForCycleEnd(prevCycle.End, closingDay, paymentDay, loc)
	}

	salary := salaryDay
	return buildPurchaseCandidate(card, status, ref, &salary, loc)
}

func containsAll(text string, parts ...string) bool {
	for _, part := range parts {
		if !contains(text, part) {
			return false
		}
	}
	return true
}

func contains(text, part string) bool {
	return len(text) >= len(part) && (text == part || len(part) == 0 || indexOf(text, part) >= 0)
}

func indexOf(text, part string) int {
	for i := 0; i+len(part) <= len(text); i++ {
		if text[i:i+len(part)] == part {
			return i
		}
	}
	return -1
}
