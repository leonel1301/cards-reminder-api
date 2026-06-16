package i18n

import (
	"fmt"
	"strings"

	"github.com/leonelortega/cards-reminder-api/internal/domain"
)

type ReminderKind string

const (
	ReminderKindUrgent     ReminderKind = "urgent"
	ReminderKindDueSoon    ReminderKind = "due_soon"
	ReminderKindOptimalDay ReminderKind = "optimal_day"
)

type CardReminder struct {
	Card   domain.Card
	Status domain.CardStatusInfo
	Owner  *domain.Owner
}

func BuildReminderNotification(kind ReminderKind, cards []CardReminder, language string) domain.PushNotification {
	lang := normalizeLanguage(language)

	switch kind {
	case ReminderKindUrgent:
		return buildUrgent(lang, cards)
	case ReminderKindDueSoon:
		return buildDueSoon(lang, cards)
	case ReminderKindOptimalDay:
		return buildOptimalDay(lang, cards)
	default:
		return domain.PushNotification{Title: "Cards Reminder", Body: ""}
	}
}

func buildUrgent(lang string, cards []CardReminder) domain.PushNotification {
	if len(cards) == 1 {
		card := cards[0]
		days := card.Status.DaysUntilPayment
		label := cardLabel(card, lang)
		switch lang {
		case "en":
			return domain.PushNotification{
				Title: "Payment due soon",
				Body:  fmt.Sprintf("%s is due in %d day(s).", label, days),
				Data:  reminderData(ReminderKindUrgent, &card),
			}
		default:
			return domain.PushNotification{
				Title: "Pago urgente",
				Body:  fmt.Sprintf("%s vence en %d día(s).", label, days),
				Data:  reminderData(ReminderKindUrgent, &card),
			}
		}
	}

	switch lang {
	case "en":
		return domain.PushNotification{
			Title: "Urgent payments",
			Body:  fmt.Sprintf("You have %d cards with payments due in 2 days or less.", len(cards)),
			Data:  reminderData(ReminderKindUrgent, nil),
		}
	default:
		return domain.PushNotification{
			Title: "Pagos urgentes",
			Body:  fmt.Sprintf("Tienes %d tarjetas con pago en 2 días o menos.", len(cards)),
			Data:  reminderData(ReminderKindUrgent, nil),
		}
	}
}

func buildDueSoon(lang string, cards []CardReminder) domain.PushNotification {
	if len(cards) == 1 {
		card := cards[0]
		days := card.Status.DaysUntilPayment
		label := cardLabel(card, lang)
		switch lang {
		case "en":
			return domain.PushNotification{
				Title: "Upcoming payment",
				Body:  fmt.Sprintf("%s is due in %d days.", label, days),
				Data:  reminderData(ReminderKindDueSoon, &card),
			}
		default:
			return domain.PushNotification{
				Title: "Pago próximo",
				Body:  fmt.Sprintf("%s vence en %d días.", label, days),
				Data:  reminderData(ReminderKindDueSoon, &card),
			}
		}
	}

	switch lang {
	case "en":
		return domain.PushNotification{
			Title: "Upcoming payments",
			Body:  fmt.Sprintf("You have %d cards with payments due within a week.", len(cards)),
			Data:  reminderData(ReminderKindDueSoon, nil),
		}
	default:
		return domain.PushNotification{
			Title: "Pagos próximos",
			Body:  fmt.Sprintf("Tienes %d tarjetas con pago en los próximos 7 días.", len(cards)),
			Data:  reminderData(ReminderKindDueSoon, nil),
		}
	}
}

func buildOptimalDay(lang string, cards []CardReminder) domain.PushNotification {
	if len(cards) == 1 {
		card := cards[0]
		label := cardLabel(card, lang)
		switch lang {
		case "en":
			return domain.PushNotification{
				Title: "Best day to buy",
				Body:  fmt.Sprintf("Today is a great day to use %s.", label),
				Data:  reminderData(ReminderKindOptimalDay, &card),
			}
		default:
			return domain.PushNotification{
				Title: "Día óptimo de compra",
				Body:  fmt.Sprintf("Hoy es un buen día para usar %s.", label),
				Data:  reminderData(ReminderKindOptimalDay, &card),
			}
		}
	}

	names := make([]string, 0, len(cards))
	for _, item := range cards {
		names = append(names, cardLabel(item, lang))
	}

	switch lang {
	case "en":
		return domain.PushNotification{
			Title: "Best day to buy",
			Body:  fmt.Sprintf("Great day to use: %s.", strings.Join(names, ", ")),
			Data:  reminderData(ReminderKindOptimalDay, nil),
		}
	default:
		return domain.PushNotification{
			Title: "Día óptimo de compra",
			Body:  fmt.Sprintf("Buen día para usar: %s.", strings.Join(names, ", ")),
			Data:  reminderData(ReminderKindOptimalDay, nil),
		}
	}
}

func cardLabel(reminder CardReminder, lang string) string {
	base := fmt.Sprintf("%s •••• %s", reminder.Card.Name, reminder.Card.LastFourDigits)
	if reminder.Owner == nil || reminder.Owner.IsSelf {
		return base
	}

	switch lang {
	case "en":
		return fmt.Sprintf("%s's %s •••• %s", reminder.Owner.Name, reminder.Card.Name, reminder.Card.LastFourDigits)
	default:
		return fmt.Sprintf("%s de %s •••• %s", reminder.Card.Name, reminder.Owner.Name, reminder.Card.LastFourDigits)
	}
}

func reminderData(kind ReminderKind, card *CardReminder) map[string]string {
	data := map[string]string{
		"type": "payment_reminder",
		"kind": string(kind),
	}
	if card == nil {
		return data
	}

	data["card_id"] = card.Card.ID.String()
	data["owner_id"] = card.Card.OwnerID.String()
	if card.Owner != nil && !card.Owner.IsSelf {
		data["owner_name"] = card.Owner.Name
	}
	return data
}

func normalizeLanguage(language string) string {
	language = strings.TrimSpace(strings.ToLower(language))
	if language == "" {
		return "es"
	}
	if idx := strings.Index(language, "-"); idx > 0 {
		language = language[:idx]
	}
	return language
}
