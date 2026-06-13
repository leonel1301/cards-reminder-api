package service

import (
	"time"

	"github.com/leonelortega/cards-reminder-api/internal/domain"
)

const defaultOptimalWindowDays = 3

func ComputeBillingCycle(ref time.Time, closingDay, paymentDueDay int) domain.BillingCycle {
	refDate := truncateToDate(ref.UTC())

	year, month, day := refDate.Date()

	var cycleEnd time.Time
	if day > closingDay {
		nextYear, nextMonth := addMonths(year, month, 1)
		cycleEnd = clampDay(nextYear, nextMonth, closingDay)
	} else {
		cycleEnd = clampDay(year, month, closingDay)
	}

	prevYear, prevMonth := addMonths(cycleEnd.Year(), cycleEnd.Month(), -1)
	prevClosing := clampDay(prevYear, prevMonth, closingDay)
	cycleStart := prevClosing.AddDate(0, 0, 1)

	payYear, payMonth := cycleEnd.Year(), cycleEnd.Month()
	if paymentDueDay <= closingDay {
		payYear, payMonth = addMonths(payYear, payMonth, 1)
	}
	paymentDue := clampDay(payYear, payMonth, paymentDueDay)

	return domain.BillingCycle{
		Start:      cycleStart,
		End:        cycleEnd,
		PaymentDue: paymentDue,
	}
}

func OptimalPurchaseDays(cycle domain.BillingCycle, count int) []time.Time {
	if count <= 0 {
		count = defaultOptimalWindowDays
	}

	days := make([]time.Time, 0, count)
	for i := range count {
		days = append(days, cycle.Start.AddDate(0, 0, i))
	}
	return days
}

func IsOptimalPurchaseDay(ref time.Time, cycle domain.BillingCycle, windowDays int) bool {
	if windowDays <= 0 {
		windowDays = defaultOptimalWindowDays
	}

	refDate := truncateToDate(ref.UTC())
	for i := range windowDays {
		if refDate.Equal(truncateToDate(cycle.Start.AddDate(0, 0, i))) {
			return true
		}
	}
	return false
}

func DaysUntilPayment(ref time.Time, paymentDue time.Time) int {
	return daysBetween(truncateToDate(ref.UTC()), truncateToDate(paymentDue))
}

func DetermineCardStatus(paid bool, daysUntilPayment int, isOptimal bool) domain.CardStatusValue {
	if paid {
		return domain.CardStatusPaid
	}
	if daysUntilPayment <= 2 {
		return domain.CardStatusUrgent
	}
	if daysUntilPayment <= 7 {
		return domain.CardStatusDueSoon
	}
	if isOptimal {
		return domain.CardStatusOptimalDay
	}
	return domain.CardStatusOnTrack
}

func BuildCardStatusInfo(ref time.Time, cycle domain.BillingCycle, paid bool) domain.CardStatusInfo {
	daysUntilPayment := DaysUntilPayment(ref, cycle.PaymentDue)
	isOptimal := IsOptimalPurchaseDay(ref, cycle, defaultOptimalWindowDays)

	return domain.CardStatusInfo{
		Status:               DetermineCardStatus(paid, daysUntilPayment, isOptimal),
		CycleStart:           cycle.Start,
		CycleEnd:             cycle.End,
		PaymentDueDate:       cycle.PaymentDue,
		DaysUntilPayment:     daysUntilPayment,
		IsOptimalPurchaseDay: isOptimal,
		IsPaidThisCycle:      paid,
	}
}

func truncateToDate(t time.Time) time.Time {
	t = t.UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func daysBetween(from, to time.Time) int {
	return int(to.Sub(from).Hours() / 24)
}

func clampDay(year int, month time.Month, day int) time.Time {
	lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
	if day > lastDay {
		day = lastDay
	}
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func addMonths(year int, month time.Month, delta int) (int, time.Month) {
	total := int(month)-1 + delta
	newYear := year + total/12
	newMonth := time.Month(total%12 + 1)
	if newMonth <= 0 {
		newMonth += 12
		newYear--
	}
	return newYear, newMonth
}
