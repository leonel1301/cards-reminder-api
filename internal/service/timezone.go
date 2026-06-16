package service

import (
	"strings"
	"time"
)

const DefaultTimezone = "America/Lima"

func NormalizeTimezone(timezone string) string {
	timezone = strings.TrimSpace(timezone)
	if timezone == "" {
		return DefaultTimezone
	}
	if _, err := time.LoadLocation(timezone); err != nil {
		return DefaultTimezone
	}
	return timezone
}

func ResolveTimezone(timezone, fallback string) string {
	timezone = strings.TrimSpace(timezone)
	if timezone != "" {
		if _, err := time.LoadLocation(timezone); err == nil {
			return timezone
		}
	}

	fallback = strings.TrimSpace(fallback)
	if fallback != "" {
		if _, err := time.LoadLocation(fallback); err == nil {
			return fallback
		}
	}

	return DefaultTimezone
}

func IsLocalReminderHour(now time.Time, timezone string, hour int, fallbackTimezone string) bool {
	loc := ResolveLocation(ResolveTimezone(timezone, fallbackTimezone))
	return now.In(loc).Hour() == hour
}
