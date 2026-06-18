package domain

import (
	"time"

	"github.com/google/uuid"
)

type CardStatusValue string

const (
	CardStatusPaid       CardStatusValue = "paid"
	CardStatusOverdue    CardStatusValue = "overdue"
	CardStatusUrgent     CardStatusValue = "urgent"
	CardStatusDueSoon    CardStatusValue = "due_soon"
	CardStatusOptimalDay CardStatusValue = "optimal_day"
	CardStatusOnTrack    CardStatusValue = "on_track"
)

type BillingCycle struct {
	Start      time.Time
	End        time.Time
	PaymentDue time.Time
}

type CardStatusInfo struct {
	Status               CardStatusValue `json:"status"`
	CycleStart           time.Time       `json:"cycle_start"`
	CycleEnd             time.Time       `json:"cycle_end"`
	PaymentDueDate       time.Time       `json:"payment_due_date"`
	DaysUntilPayment     int             `json:"days_until_payment"`
	DaysOverdue          int             `json:"days_overdue"`
	OptimalPurchaseDay   int             `json:"optimal_purchase_day"`
	IsOptimalPurchaseDay bool            `json:"is_optimal_purchase_day"`
	IsPaidThisCycle      bool            `json:"is_paid_this_cycle"`
}

type CardStatusResponse struct {
	Card                Card        `json:"card"`
	Status              CardStatusInfo `json:"status"`
	OptimalPurchaseDays []time.Time `json:"optimal_purchase_days"`
}

type OptimalPurchaseDaysResponse struct {
	Card                Card        `json:"card"`
	Cycle               BillingCycleDates `json:"cycle"`
	OptimalPurchaseDays []time.Time `json:"optimal_purchase_days"`
}

type BillingCycleDates struct {
	Start      time.Time `json:"start"`
	End        time.Time `json:"end"`
	PaymentDue time.Time `json:"payment_due"`
}

type DashboardItem struct {
	Card   Card           `json:"card"`
	Status CardStatusInfo `json:"status"`
}

type DashboardSummary struct {
	Total      int `json:"total"`
	Overdue    int `json:"overdue"`
	Urgent     int `json:"urgent"`
	DueSoon    int `json:"due_soon"`
	Paid       int `json:"paid"`
	OptimalDay int `json:"optimal_day"`
	OnTrack    int `json:"on_track"`
}

type DashboardResponse struct {
	Cards   []DashboardItem  `json:"cards"`
	Summary DashboardSummary `json:"summary"`
}

type Payment struct {
	ID       uuid.UUID `json:"id"`
	CardID   uuid.UUID `json:"card_id"`
	CycleEnd time.Time `json:"cycle_end"`
	PaidAt   time.Time `json:"paid_at"`
	Notes    *string   `json:"notes"`
}

type CurrentCycleResponse struct {
	Card   Card              `json:"card"`
	Cycle  BillingCycleDates `json:"cycle"`
	Status CardStatusInfo    `json:"status"`
}

type PaymentsResponse struct {
	Card     Card      `json:"card"`
	Payments []Payment `json:"payments"`
}
