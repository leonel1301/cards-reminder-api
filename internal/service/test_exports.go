//go:build test

package service

import (
	"time"

	"github.com/leonelortega/cards-reminder-api/internal/domain"
)

type PurchaseCandidate = purchaseCandidate

func BuildPurchaseCandidate(
	card domain.Card,
	status domain.CardStatusInfo,
	now time.Time,
	salaryDay *int,
	loc *time.Location,
) PurchaseCandidate {
	return buildPurchaseCandidate(card, status, now, salaryDay, loc)
}

func BuildBestForPurchaseWhy(
	candidate PurchaseCandidate,
	refDate time.Time,
	loc *time.Location,
	language string,
) string {
	return buildBestForPurchaseWhy(candidate, refDate, loc, language)
}

func TruncateToDateInLoc(t time.Time, loc *time.Location) time.Time {
	return truncateToDateInLoc(t, loc)
}
