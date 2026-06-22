package service_test

import (
	"testing"
	"time"

	"github.com/leonelortega/cards-reminder-api/internal/domain"
	"github.com/leonelortega/cards-reminder-api/internal/service"
)

func TestComputeBillingCycle_beforeClosing(t *testing.T) {
	ref := time.Date(2026, 2, 10, 12, 0, 0, 0, time.UTC)
	cycle := service.ComputeBillingCycle(ref, 15, 5, time.UTC)

	assertDate(t, cycle.Start, 2026, 1, 16)
	assertDate(t, cycle.End, 2026, 2, 15)
	assertDate(t, cycle.PaymentDue, 2026, 3, 5)
}

func TestComputeBillingCycle_afterClosing(t *testing.T) {
	ref := time.Date(2026, 2, 20, 12, 0, 0, 0, time.UTC)
	cycle := service.ComputeBillingCycle(ref, 15, 5, time.UTC)

	assertDate(t, cycle.Start, 2026, 2, 16)
	assertDate(t, cycle.End, 2026, 3, 15)
	assertDate(t, cycle.PaymentDue, 2026, 4, 5)
}

func TestComputeBillingCycle_onClosingDay(t *testing.T) {
	ref := time.Date(2026, 2, 15, 12, 0, 0, 0, time.UTC)
	cycle := service.ComputeBillingCycle(ref, 15, 5, time.UTC)

	assertDate(t, cycle.End, 2026, 2, 15)
	assertDate(t, cycle.Start, 2026, 1, 16)
}

func TestComputeBillingCycle_paymentSameMonth(t *testing.T) {
	ref := time.Date(2026, 1, 3, 12, 0, 0, 0, time.UTC)
	cycle := service.ComputeBillingCycle(ref, 5, 25, time.UTC)

	assertDate(t, cycle.End, 2026, 1, 5)
	assertDate(t, cycle.PaymentDue, 2026, 1, 25)
}

func TestComputeBillingCycle_closingDay31InFebruary(t *testing.T) {
	ref := time.Date(2026, 2, 10, 12, 0, 0, 0, time.UTC)
	cycle := service.ComputeBillingCycle(ref, 31, 10, time.UTC)

	assertDate(t, cycle.End, 2026, 2, 28)
}

func TestDaysUntilCurrentMonthPayment_beforeDueDate(t *testing.T) {
	ref := time.Date(2026, 6, 14, 12, 0, 0, 0, time.UTC)
	days := service.DaysUntilCurrentMonthPayment(ref, 20, time.UTC)

	if days != 6 {
		t.Fatalf("got %d days, want 6", days)
	}
}

func TestDaysUntilCurrentMonthPayment_ignoresBillingCyclePayment(t *testing.T) {
	ref := time.Date(2026, 6, 14, 12, 0, 0, 0, time.UTC)
	cycle := service.ComputeBillingCycle(ref, 10, 20, time.UTC)

	if cycle.PaymentDue.Month() != time.July {
		t.Fatalf("billing cycle payment due should be July, got %s", cycle.PaymentDue.Format("2006-01-02"))
	}

	days := service.DaysUntilCurrentMonthPayment(ref, 20, time.UTC)
	if days != 6 {
		t.Fatalf("got %d days, want 6 based on current month", days)
	}
}

func TestDaysUntilCurrentMonthPayment_afterDueDate(t *testing.T) {
	ref := time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)
	days := service.DaysUntilCurrentMonthPayment(ref, 20, time.UTC)

	if days != 25 {
		t.Fatalf("got %d days, want 25 until next month payment", days)
	}
}

func TestIsPaymentOverdue(t *testing.T) {
	ref := time.Date(2026, 6, 17, 12, 0, 0, 0, time.UTC)

	if !service.IsPaymentOverdue(ref, 12, false, time.UTC) {
		t.Fatal("expected overdue when due day passed and unpaid")
	}
	if service.IsPaymentOverdue(ref, 12, true, time.UTC) {
		t.Fatal("paid card should not be overdue")
	}
	if service.IsPaymentOverdue(ref, 18, false, time.UTC) {
		t.Fatal("future due date should not be overdue")
	}
}

func TestDaysOverdue(t *testing.T) {
	ref := time.Date(2026, 6, 17, 12, 0, 0, 0, time.UTC)
	days := service.DaysOverdue(ref, 12, time.UTC)

	if days != 5 {
		t.Fatalf("got %d days overdue, want 5", days)
	}
}

func TestBuildCardStatusInfo_overdue(t *testing.T) {
	ref := time.Date(2026, 6, 17, 12, 0, 0, 0, time.UTC)
	currentCycle := service.ComputeBillingCycle(ref, 10, 12, time.UTC)
	prevCycle := service.PreviousBillingCycle(currentCycle, 10, time.UTC)
	paymentDue := service.PaymentDueForCycleEnd(prevCycle.End, 10, 12, time.UTC)

	status := service.BuildCardStatusInfo(ref, prevCycle, paymentDue, 10, false, time.UTC)

	if status.Status != domain.CardStatusOverdue {
		t.Fatalf("got status %q, want overdue", status.Status)
	}
	if status.DaysOverdue != 5 {
		t.Fatalf("got days_overdue %d, want 5", status.DaysOverdue)
	}
	assertDate(t, status.CycleEnd, prevCycle.End.Year(), prevCycle.End.Month(), prevCycle.End.Day())
	assertDate(t, status.PaymentDueDate, paymentDue.Year(), paymentDue.Month(), paymentDue.Day())
}

func TestCurrentMonthPaymentDue(t *testing.T) {
	ref := time.Date(2026, 6, 14, 12, 0, 0, 0, time.UTC)
	due := service.CurrentMonthPaymentDue(ref, 20, time.UTC)

	assertDate(t, due, 2026, 6, 20)
}

func TestComputeOptimalPurchaseDay_afterClosing(t *testing.T) {
	got := service.ComputeOptimalPurchaseDay(10)
	if got != 11 {
		t.Fatalf("got %d, want 11", got)
	}
}

func TestComputeOptimalPurchaseDay_wrapsAfterMonthEnd(t *testing.T) {
	got := service.ComputeOptimalPurchaseDay(31)
	if got != 1 {
		t.Fatalf("got %d, want 1", got)
	}
}

func TestIsOptimalPurchaseDayInMonth(t *testing.T) {
	optimalDay := 11
	ref := time.Date(2026, 6, 12, 12, 0, 0, 0, time.UTC)

	if !service.IsOptimalPurchaseDayInMonth(ref, optimalDay, 3, time.UTC) {
		t.Fatal("expected day 12 to be optimal when window starts on day 11")
	}

	ref = time.Date(2026, 6, 20, 12, 0, 0, 0, time.UTC)
	if service.IsOptimalPurchaseDayInMonth(ref, optimalDay, 3, time.UTC) {
		t.Fatal("expected day 20 not to be optimal")
	}
}

func TestBuildCardStatusInfo_usesPaymentObligationCycle(t *testing.T) {
	ref := time.Date(2026, 6, 14, 12, 0, 0, 0, time.UTC)
	currentCycle := service.ComputeBillingCycle(ref, 10, 20, time.UTC)
	prevCycle := service.PreviousBillingCycle(currentCycle, 10, time.UTC)
	paymentDue := service.PaymentDueForCycleEnd(prevCycle.End, 10, 20, time.UTC)
	status := service.BuildCardStatusInfo(ref, prevCycle, paymentDue, 10, false, time.UTC)

	assertDate(t, status.CycleStart, prevCycle.Start.Year(), prevCycle.Start.Month(), prevCycle.Start.Day())
	assertDate(t, status.CycleEnd, prevCycle.End.Year(), prevCycle.End.Month(), prevCycle.End.Day())
	assertDate(t, status.PaymentDueDate, 2026, 6, 20)
	if status.DaysUntilPayment != 6 {
		t.Fatalf("got %d days_until_payment, want 6", status.DaysUntilPayment)
	}
	if status.OptimalPurchaseDay != 11 {
		t.Fatalf("got optimal_purchase_day %d, want 11", status.OptimalPurchaseDay)
	}
	if status.Status != domain.CardStatusDueSoon {
		t.Fatalf("got status %q, want due_soon", status.Status)
	}
}

func TestBuildCardStatusInfo_BCP_paymentObligationCycle(t *testing.T) {
	ref := time.Date(2026, 6, 18, 12, 0, 0, 0, time.UTC)
	currentCycle := service.ComputeBillingCycle(ref, 25, 23, time.UTC)
	prevCycle := service.PreviousBillingCycle(currentCycle, 25, time.UTC)
	paymentDue := service.PaymentDueForCycleEnd(prevCycle.End, 25, 23, time.UTC)

	status := service.BuildCardStatusInfo(ref, prevCycle, paymentDue, 25, false, time.UTC)

	assertDate(t, status.CycleStart, 2026, 4, 26)
	assertDate(t, status.CycleEnd, 2026, 5, 25)
	assertDate(t, status.PaymentDueDate, 2026, 6, 23)
	if status.DaysUntilPayment != 5 {
		t.Fatalf("got %d days_until_payment, want 5", status.DaysUntilPayment)
	}
	if status.Status != domain.CardStatusDueSoon {
		t.Fatalf("got status %q, want due_soon", status.Status)
	}
	assertDate(t, currentCycle.Start, 2026, 5, 26)
	assertDate(t, currentCycle.End, 2026, 6, 25)
}

func TestDetermineCardStatus_priority(t *testing.T) {
	tests := []struct {
		name    string
		paid    bool
		overdue bool
		days    int
		optimal bool
		want    string
	}{
		{"paid wins", true, false, 1, true, "paid"},
		{"overdue", false, true, 1, true, "overdue"},
		{"urgent", false, false, 2, true, "urgent"},
		{"due soon", false, false, 5, true, "due_soon"},
		{"optimal", false, false, 10, true, "optimal_day"},
		{"on track", false, false, 10, false, "on_track"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.DetermineCardStatus(tt.paid, tt.overdue, tt.days, tt.optimal)
			if string(got) != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolveLocation_invalidFallsBackToUTC(t *testing.T) {
	loc := service.ResolveLocation("Invalid/Timezone")
	if loc != time.UTC {
		t.Fatalf("expected UTC fallback, got %s", loc.String())
	}
}
