package domain

// ContractExtraction is the structured result returned after GPT analyzes
// a credit-card contract PDF or image.
type ContractExtraction struct {
	Name                 *string `json:"name"`
	LastFourDigits       *string `json:"last_four_digits"`
	Issuer               *string `json:"issuer"`
	BillingCycleDay      *int    `json:"billing_cycle_day"`
	PaymentDueDay        *int    `json:"payment_due_day"`
	AnnualFee            *string `json:"annual_fee"`
	InterestRateSummary  *string `json:"interest_rate_summary"`
	Notes                *string `json:"notes"`
	Summary              string  `json:"summary"`
	Confidence           string  `json:"confidence"`
	Warnings             []string `json:"warnings"`
}
