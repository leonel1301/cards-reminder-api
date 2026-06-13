package service

import (
	"regexp"

	"github.com/leonelortega/cards-reminder-api/internal/domain"
)

var lastFourDigitsPattern = regexp.MustCompile(`^\d{4}$`)

func validateCreateInput(input domain.CreateCardInput) error {
	if input.Name == "" {
		return ValidationError{Field: "name", Message: "is required"}
	}
	if !lastFourDigitsPattern.MatchString(input.LastFourDigits) {
		return ValidationError{Field: "last_four_digits", Message: "must be exactly 4 digits"}
	}
	if err := validateDay("billing_cycle_day", input.BillingCycleDay); err != nil {
		return err
	}
	if err := validateDay("payment_due_day", input.PaymentDueDay); err != nil {
		return err
	}
	return nil
}

func validateUpdateInput(input domain.UpdateCardInput) error {
	if input.Name != nil && *input.Name == "" {
		return ValidationError{Field: "name", Message: "cannot be empty"}
	}
	if input.LastFourDigits != nil && !lastFourDigitsPattern.MatchString(*input.LastFourDigits) {
		return ValidationError{Field: "last_four_digits", Message: "must be exactly 4 digits"}
	}
	if input.BillingCycleDay != nil {
		if err := validateDay("billing_cycle_day", *input.BillingCycleDay); err != nil {
			return err
		}
	}
	if input.PaymentDueDay != nil {
		if err := validateDay("payment_due_day", *input.PaymentDueDay); err != nil {
			return err
		}
	}
	return nil
}

func validateDay(field string, day int) error {
	if day < 1 || day > 31 {
		return ValidationError{Field: field, Message: "must be between 1 and 31"}
	}
	return nil
}

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return e.Field + " " + e.Message
}
