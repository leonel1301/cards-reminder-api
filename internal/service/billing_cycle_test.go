package service

import (
	"testing"
	"time"

	"github.com/leonelortega/cards-reminder-api/internal/domain"
)

func TestComputeBillingCycle_beforeClosing(t *testing.T) {
	ref := time.Date(2026, 2, 10, 12, 0, 0, 0, time.UTC)
	cycle := ComputeBillingCycle(ref, 15, 5, time.UTC)

	assertDate(t, cycle.Start, 2026, 1, 16)
	assertDate(t, cycle.End, 2026, 2, 15)
	assertDate(t, cycle.PaymentDue, 2026, 3, 5)
}

func TestComputeBillingCycle_afterClosing(t *testing.T) {
	ref := time.Date(2026, 2, 20, 12, 0, 0, 0, time.UTC)
	cycle := ComputeBillingCycle(ref, 15, 5, time.UTC)

	assertDate(t, cycle.Start, 2026, 2, 16)
	assertDate(t, cycle.End, 2026, 3, 15)
	assertDate(t, cycle.PaymentDue, 2026, 4, 5)
}

func TestComputeBillingCycle_onClosingDay(t *testing.T) {
	ref := time.Date(2026, 2, 15, 12, 0, 0, 0, time.UTC)
	cycle := ComputeBillingCycle(ref, 15, 5, time.UTC)

	assertDate(t, cycle.End, 2026, 2, 15)
	assertDate(t, cycle.Start, 2026, 1, 16)
}

func TestComputeBillingCycle_paymentSameMonth(t *testing.T) {
	ref := time.Date(2026, 1, 3, 12, 0, 0, 0, time.UTC)
	cycle := ComputeBillingCycle(ref, 5, 25, time.UTC)

	assertDate(t, cycle.End, 2026, 1, 5)
	assertDate(t, cycle.PaymentDue, 2026, 1, 25)
}

func TestComputeBillingCycle_closingDay31InFebruary(t *testing.T) {
	ref := time.Date(2026, 2, 10, 12, 0, 0, 0, time.UTC)
	cycle := ComputeBillingCycle(ref, 31, 10, time.UTC)

	assertDate(t, cycle.End, 2026, 2, 28)
}

func TestDaysUntilCurrentMonthPayment_beforeDueDate(t *testing.T) {
	ref := time.Date(2026, 6, 14, 12, 0, 0, 0, time.UTC)
	days := DaysUntilCurrentMonthPayment(ref, 20, time.UTC)

	if days != 6 {
		t.Fatalf("got %d days, want 6", days)
	}
}

func TestDaysUntilCurrentMonthPayment_ignoresBillingCyclePayment(t *testing.T) {
	ref := time.Date(2026, 6, 14, 12, 0, 0, 0, time.UTC)
	cycle := ComputeBillingCycle(ref, 10, 20, time.UTC)

	if cycle.PaymentDue.Month() != time.July {
		t.Fatalf("billing cycle payment due should be July, got %s", cycle.PaymentDue.Format("2006-01-02"))
	}

	days := DaysUntilCurrentMonthPayment(ref, 20, time.UTC)
	if days != 6 {
		t.Fatalf("got %d days, want 6 based on current month", days)
	}
}

func TestDaysUntilCurrentMonthPayment_afterDueDate(t *testing.T) {
	ref := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)
	days := DaysUntilCurrentMonthPayment(ref, 20, time.UTC)

	if days != 0 {
		t.Fatalf("got %d days, want 0 when overdue in current month", days)
	}
}

func TestCurrentMonthPaymentDue(t *testing.T) {
	ref := time.Date(2026, 6, 14, 12, 0, 0, 0, time.UTC)
	due := CurrentMonthPaymentDue(ref, 20, time.UTC)

	assertDate(t, due, 2026, 6, 20)
}

func TestComputeOptimalPurchaseDay_afterClosing(t *testing.T) {
	got := ComputeOptimalPurchaseDay(10, nil)
	if got != 11 {
		t.Fatalf("got %d, want 11", got)
	}
}

func TestComputeOptimalPurchaseDay_considersSalaryDay(t *testing.T) {
	salaryDay := 15
	got := ComputeOptimalPurchaseDay(10, &salaryDay)
	if got != 15 {
		t.Fatalf("got %d, want 15", got)
	}
}

func TestComputeOptimalPurchaseDay_salaryBeforeCycleStart(t *testing.T) {
	salaryDay := 5
	got := ComputeOptimalPurchaseDay(10, &salaryDay)
	if got != 11 {
		t.Fatalf("got %d, want 11", got)
	}
}

func TestIsOptimalPurchaseDayInMonth(t *testing.T) {
	optimalDay := 11
	ref := time.Date(2026, 6, 12, 12, 0, 0, 0, time.UTC)

	if !IsOptimalPurchaseDayInMonth(ref, optimalDay, 3, time.UTC) {
		t.Fatal("expected day 12 to be optimal when window starts on day 11")
	}

	ref = time.Date(2026, 6, 20, 12, 0, 0, 0, time.UTC)
	if IsOptimalPurchaseDayInMonth(ref, optimalDay, 3, time.UTC) {
		t.Fatal("expected day 20 not to be optimal")
	}
}

func TestBuildCardStatusInfo_usesCurrentMonthPayment(t *testing.T) {
	ref := time.Date(2026, 6, 14, 12, 0, 0, 0, time.UTC)
	cycle := ComputeBillingCycle(ref, 10, 20, time.UTC)
	salaryDay := 15

	status := BuildCardStatusInfo(ref, cycle, 20, 10, &salaryDay, false, time.UTC)

	assertDate(t, status.PaymentDueDate, 2026, 6, 20)
	if status.DaysUntilPayment != 6 {
		t.Fatalf("got %d days_until_payment, want 6", status.DaysUntilPayment)
	}
	if status.OptimalPurchaseDay != 15 {
		t.Fatalf("got optimal_purchase_day %d, want 15", status.OptimalPurchaseDay)
	}
	if status.Status != domain.CardStatusDueSoon {
		t.Fatalf("got status %q, want due_soon", status.Status)
	}
}

func TestDetermineCardStatus_priority(t *testing.T) {
	tests := []struct {
		name    string
		paid    bool
		days    int
		optimal bool
		want    string
	}{
		{"paid wins", true, 1, true, "paid"},
		{"urgent", false, 2, true, "urgent"},
		{"due soon", false, 5, true, "due_soon"},
		{"optimal", false, 10, true, "optimal_day"},
		{"on track", false, 10, false, "on_track"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetermineCardStatus(tt.paid, tt.days, tt.optimal)
			if string(got) != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolveLocation_invalidFallsBackToUTC(t *testing.T) {
	loc := ResolveLocation("Invalid/Timezone")
	if loc != time.UTC {
		t.Fatalf("expected UTC fallback, got %s", loc.String())
	}
}

func assertDate(t *testing.T, got time.Time, year int, month time.Month, day int) {
	t.Helper()
	if got.Year() != year || got.Month() != month || got.Day() != day {
		t.Fatalf("got %s, want %d-%02d-%02d", got.Format("2006-01-02"), year, month, day)
	}
}
