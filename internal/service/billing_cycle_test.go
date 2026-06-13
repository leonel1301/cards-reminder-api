package service

import (
	"testing"
	"time"

	"github.com/leonelortega/cards-reminder-api/internal/domain"
)

func TestComputeBillingCycle_beforeClosing(t *testing.T) {
	ref := time.Date(2026, 2, 10, 12, 0, 0, 0, time.UTC)
	cycle := ComputeBillingCycle(ref, 15, 5)

	assertDate(t, cycle.Start, 2026, 1, 16)
	assertDate(t, cycle.End, 2026, 2, 15)
	assertDate(t, cycle.PaymentDue, 2026, 3, 5)
}

func TestComputeBillingCycle_afterClosing(t *testing.T) {
	ref := time.Date(2026, 2, 20, 12, 0, 0, 0, time.UTC)
	cycle := ComputeBillingCycle(ref, 15, 5)

	assertDate(t, cycle.Start, 2026, 2, 16)
	assertDate(t, cycle.End, 2026, 3, 15)
	assertDate(t, cycle.PaymentDue, 2026, 4, 5)
}

func TestComputeBillingCycle_onClosingDay(t *testing.T) {
	ref := time.Date(2026, 2, 15, 12, 0, 0, 0, time.UTC)
	cycle := ComputeBillingCycle(ref, 15, 5)

	assertDate(t, cycle.End, 2026, 2, 15)
	assertDate(t, cycle.Start, 2026, 1, 16)
}

func TestComputeBillingCycle_paymentSameMonth(t *testing.T) {
	ref := time.Date(2026, 1, 3, 12, 0, 0, 0, time.UTC)
	cycle := ComputeBillingCycle(ref, 5, 25)

	assertDate(t, cycle.End, 2026, 1, 5)
	assertDate(t, cycle.PaymentDue, 2026, 1, 25)
}

func TestComputeBillingCycle_closingDay31InFebruary(t *testing.T) {
	ref := time.Date(2026, 2, 10, 12, 0, 0, 0, time.UTC)
	cycle := ComputeBillingCycle(ref, 31, 10)

	assertDate(t, cycle.End, 2026, 2, 28)
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

func TestIsOptimalPurchaseDay_firstDaysOfCycle(t *testing.T) {
	cycle := domain.BillingCycle{
		Start:      time.Date(2026, 2, 16, 0, 0, 0, 0, time.UTC),
		End:        time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC),
		PaymentDue: time.Date(2026, 4, 5, 0, 0, 0, 0, time.UTC),
	}

	if !IsOptimalPurchaseDay(time.Date(2026, 2, 16, 0, 0, 0, 0, time.UTC), cycle, 3) {
		t.Fatal("expected day 1 of cycle to be optimal")
	}
	if IsOptimalPurchaseDay(time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC), cycle, 3) {
		t.Fatal("expected day 5 of cycle not to be optimal")
	}
}

func assertDate(t *testing.T, got time.Time, year int, month time.Month, day int) {
	t.Helper()
	if got.Year() != year || got.Month() != month || got.Day() != day {
		t.Fatalf("got %s, want %d-%02d-%02d", got.Format("2006-01-02"), year, month, day)
	}
}
