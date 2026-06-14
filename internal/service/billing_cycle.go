package service

import (
	"time"

	"github.com/leonelortega/cards-reminder-api/internal/domain"
)

const defaultOptimalWindowDays = 3

func ResolveLocation(tz string) *time.Location {
	if tz == "" {
		return time.UTC
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return time.UTC
	}
	return loc
}

func ComputeBillingCycle(ref time.Time, closingDay, paymentDueDay int, loc *time.Location) domain.BillingCycle {
	refDate := truncateToDateInLoc(ref, loc)

	year, month, day := refDate.Date()

	var cycleEnd time.Time
	if day > closingDay {
		nextYear, nextMonth := addMonths(year, month, 1)
		cycleEnd = clampDayInLoc(nextYear, nextMonth, closingDay, loc)
	} else {
		cycleEnd = clampDayInLoc(year, month, closingDay, loc)
	}

	prevYear, prevMonth := addMonths(cycleEnd.Year(), cycleEnd.Month(), -1)
	prevClosing := clampDayInLoc(prevYear, prevMonth, closingDay, loc)
	cycleStart := prevClosing.AddDate(0, 0, 1)

	payYear, payMonth := cycleEnd.Year(), cycleEnd.Month()
	if paymentDueDay <= closingDay {
		payYear, payMonth = addMonths(payYear, payMonth, 1)
	}
	paymentDue := clampDayInLoc(payYear, payMonth, paymentDueDay, loc)

	return domain.BillingCycle{
		Start:      cycleStart,
		End:        cycleEnd,
		PaymentDue: paymentDue,
	}
}

func CurrentMonthPaymentDue(ref time.Time, paymentDueDay int, loc *time.Location) time.Time {
	refDate := truncateToDateInLoc(ref, loc)
	year, month, _ := refDate.Date()
	return clampDayInLoc(year, month, paymentDueDay, loc)
}

func DaysUntilCurrentMonthPayment(ref time.Time, paymentDueDay int, loc *time.Location) int {
	refDate := truncateToDateInLoc(ref, loc)
	paymentDue := CurrentMonthPaymentDue(ref, paymentDueDay, loc)

	if refDate.After(paymentDue) {
		return 0
	}
	return daysBetween(refDate, paymentDue)
}

func ComputeOptimalPurchaseDay(billingCycleDay int, salaryDay *int) int {
	optimal := billingCycleDay + 1
	if optimal > 31 {
		optimal = 1
	}
	if salaryDay != nil && *salaryDay > 0 && *salaryDay > optimal {
		optimal = *salaryDay
	}
	return optimal
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

func OptimalPurchaseDaysInMonth(ref time.Time, optimalDay int, count int, loc *time.Location) []time.Time {
	if count <= 0 {
		count = defaultOptimalWindowDays
	}

	refDate := truncateToDateInLoc(ref, loc)
	year, month := refDate.Year(), refDate.Month()
	lastDay := lastDayOfMonth(year, month, loc)

	days := make([]time.Time, 0, count)
	for i := range count {
		day := optimalDay + i
		if day > lastDay {
			break
		}
		days = append(days, time.Date(year, month, day, 0, 0, 0, 0, loc))
	}
	return days
}

func IsOptimalPurchaseDayInMonth(ref time.Time, optimalDay int, windowDays int, loc *time.Location) bool {
	if windowDays <= 0 {
		windowDays = defaultOptimalWindowDays
	}

	refDate := truncateToDateInLoc(ref, loc)
	day := refDate.Day()
	lastDay := lastDayOfMonth(refDate.Year(), refDate.Month(), loc)

	for i := range windowDays {
		candidate := optimalDay + i
		if candidate > lastDay {
			break
		}
		if day == candidate {
			return true
		}
	}
	return false
}

func DaysUntilPayment(ref time.Time, paymentDue time.Time) int {
	return daysBetween(truncateToDateInLoc(ref, time.UTC), truncateToDateInLoc(paymentDue, time.UTC))
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

func BuildCardStatusInfo(
	ref time.Time,
	cycle domain.BillingCycle,
	paymentDueDay int,
	billingCycleDay int,
	salaryDay *int,
	paid bool,
	loc *time.Location,
) domain.CardStatusInfo {
	paymentDueDate := CurrentMonthPaymentDue(ref, paymentDueDay, loc)
	daysUntilPayment := DaysUntilCurrentMonthPayment(ref, paymentDueDay, loc)
	optimalPurchaseDay := ComputeOptimalPurchaseDay(billingCycleDay, salaryDay)
	isOptimal := IsOptimalPurchaseDayInMonth(ref, optimalPurchaseDay, defaultOptimalWindowDays, loc)

	return domain.CardStatusInfo{
		Status:               DetermineCardStatus(paid, daysUntilPayment, isOptimal),
		CycleStart:           cycle.Start,
		CycleEnd:             cycle.End,
		PaymentDueDate:       paymentDueDate,
		DaysUntilPayment:     daysUntilPayment,
		OptimalPurchaseDay:   optimalPurchaseDay,
		IsOptimalPurchaseDay: isOptimal,
		IsPaidThisCycle:      paid,
	}
}

func truncateToDateInLoc(t time.Time, loc *time.Location) time.Time {
	t = t.In(loc)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
}

func daysBetween(from, to time.Time) int {
	return int(to.Sub(from).Hours() / 24)
}

func clampDayInLoc(year int, month time.Month, day int, loc *time.Location) time.Time {
	lastDay := lastDayOfMonth(year, month, loc)
	if day > lastDay {
		day = lastDay
	}
	return time.Date(year, month, day, 0, 0, 0, 0, loc)
}

func lastDayOfMonth(year int, month time.Month, loc *time.Location) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, loc).Day()
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
