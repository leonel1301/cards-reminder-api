package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/leonelortega/cards-reminder-api/internal/domain"
)

type purchaseCandidate struct {
	Card            domain.Card
	Status          domain.CardStatusInfo
	NewPurchaseDue  time.Time
	FinancingDays   int
	SalaryDay       *int
	AlignsWithSalary bool
}

func RecommendBestForPurchase(
	candidates []purchaseCandidate,
	now time.Time,
	loc *time.Location,
	language string,
) *domain.BestForPurchaseRecommendation {
	if len(candidates) == 0 {
		return nil
	}

	refDate := truncateToDateInLoc(now, loc)
	best := candidates[0]
	bestScore := scorePurchaseCandidate(best, refDate, loc)

	for _, candidate := range candidates[1:] {
		if candidateScore := scorePurchaseCandidate(candidate, refDate, loc); candidateScore > bestScore {
			best = candidate
			bestScore = candidateScore
		}
	}

	return &domain.BestForPurchaseRecommendation{
		CardID: best.Card.ID,
		Why:    buildBestForPurchaseWhy(best, refDate, loc, language),
	}
}

func scorePurchaseCandidate(candidate purchaseCandidate, refDate time.Time, loc *time.Location) int {
	score := candidate.FinancingDays * 100

	if candidate.SalaryDay != nil && *candidate.SalaryDay > 0 {
		if candidate.AlignsWithSalary {
			score += 10000
		} else {
			score -= 100000
		}
	}

	if candidate.Status.IsPaidThisCycle {
		score += 500
	}

	return score
}

func buildPurchaseCandidate(
	card domain.Card,
	status domain.CardStatusInfo,
	now time.Time,
	salaryDay *int,
	loc *time.Location,
) purchaseCandidate {
	currentCycle := ComputeBillingCycle(now, card.BillingCycleDay, card.PaymentDueDay, loc)
	refDate := truncateToDateInLoc(now, loc)
	newPurchaseDue := truncateToDateInLoc(currentCycle.PaymentDue, loc)
	financingDays := daysBetween(refDate, newPurchaseDue)

	candidate := purchaseCandidate{
		Card:           card,
		Status:         status,
		NewPurchaseDue: newPurchaseDue,
		FinancingDays:  financingDays,
		SalaryDay:      salaryDay,
	}

	if salaryDay != nil && *salaryDay > 0 {
		nextSalary := NextSalaryDate(refDate, *salaryDay, loc)
		candidate.AlignsWithSalary = newPurchaseDue.After(nextSalary)
	} else {
		candidate.AlignsWithSalary = true
	}

	return candidate
}

func NextSalaryDate(ref time.Time, salaryDay int, loc *time.Location) time.Time {
	refDate := truncateToDateInLoc(ref, loc)
	year, month, _ := refDate.Date()
	candidate := clampDayInLoc(year, month, salaryDay, loc)
	if !refDate.After(candidate) {
		return candidate
	}

	nextYear, nextMonth := addMonths(year, month, 1)
	return clampDayInLoc(nextYear, nextMonth, salaryDay, loc)
}

func buildBestForPurchaseWhy(candidate purchaseCandidate, refDate time.Time, loc *time.Location, language string) string {
	lang := normalizeRecommendationLanguage(language)
	cardLabel := formatCardLabel(candidate.Card)
	dueDate := formatRecommendationDate(candidate.NewPurchaseDue, loc)

	if lang == "en" {
		return buildBestForPurchaseWhyEN(candidate, cardLabel, dueDate, refDate, loc)
	}
	return buildBestForPurchaseWhyES(candidate, cardLabel, dueDate, refDate, loc)
}

func buildBestForPurchaseWhyES(
	candidate purchaseCandidate,
	cardLabel string,
	dueDate string,
	refDate time.Time,
	loc *time.Location,
) string {
	why := fmt.Sprintf(
		"%s te da %d días de financiamiento: una compra hoy vence el %s.",
		cardLabel,
		candidate.FinancingDays,
		dueDate,
	)

	if candidate.SalaryDay != nil && *candidate.SalaryDay > 0 {
		nextSalary := NextSalaryDate(refDate, *candidate.SalaryDay, loc)
		salaryDate := formatRecommendationDate(nextSalary, loc)
		if candidate.AlignsWithSalary {
			why += fmt.Sprintf(" El pago cae después de tu sueldo del %s.", salaryDate)
		} else {
			why += fmt.Sprintf(" Ojo: el pago cae antes de tu sueldo del %s.", salaryDate)
		}
	}

	if !candidate.Status.IsPaidThisCycle {
		pendingDate := formatRecommendationDate(truncateToDateInLoc(candidate.Status.PaymentDueDate, loc), loc)
		why += fmt.Sprintf(
			" Igual tenés un pago pendiente el %s por el ciclo anterior; la compra nueva no lo reemplaza.",
			pendingDate,
		)
	}

	return why
}

func buildBestForPurchaseWhyEN(
	candidate purchaseCandidate,
	cardLabel string,
	dueDate string,
	refDate time.Time,
	loc *time.Location,
) string {
	why := fmt.Sprintf(
		"%s gives you %d financing days: a purchase today is due on %s.",
		cardLabel,
		candidate.FinancingDays,
		dueDate,
	)

	if candidate.SalaryDay != nil && *candidate.SalaryDay > 0 {
		nextSalary := NextSalaryDate(refDate, *candidate.SalaryDay, loc)
		salaryDate := formatRecommendationDate(nextSalary, loc)
		if candidate.AlignsWithSalary {
			why += fmt.Sprintf(" Payment falls after your salary on %s.", salaryDate)
		} else {
			why += fmt.Sprintf(" Note: payment falls before your salary on %s.", salaryDate)
		}
	}

	if !candidate.Status.IsPaidThisCycle {
		pendingDate := formatRecommendationDate(truncateToDateInLoc(candidate.Status.PaymentDueDate, loc), loc)
		why += fmt.Sprintf(
			" You still have a pending payment on %s from the previous cycle; a new purchase does not replace it.",
			pendingDate,
		)
	}

	return why
}

func formatCardLabel(card domain.Card) string {
	if card.LastFourDigits != "" {
		return fmt.Sprintf("%s •••• %s", card.Name, card.LastFourDigits)
	}
	return card.Name
}

func formatRecommendationDate(date time.Time, loc *time.Location) string {
	return date.In(loc).Format("02/01/2006")
}

func normalizeRecommendationLanguage(language string) string {
	language = strings.TrimSpace(strings.ToLower(language))
	if language == "" {
		return "es"
	}
	if idx := strings.Index(language, "-"); idx > 0 {
		language = language[:idx]
	}
	if language == "en" {
		return "en"
	}
	return "es"
}
