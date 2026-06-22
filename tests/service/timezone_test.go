package service_test

import (
	"testing"
	"time"

	"github.com/leonelortega/cards-reminder-api/internal/service"
)

func TestNormalizeTimezone(t *testing.T) {
	if got := service.NormalizeTimezone("America/New_York"); got != "America/New_York" {
		t.Fatalf("got %q", got)
	}
	if got := service.NormalizeTimezone(""); got != service.DefaultTimezone {
		t.Fatalf("empty should default to %s, got %q", service.DefaultTimezone, got)
	}
	if got := service.NormalizeTimezone("Invalid/Zone"); got != service.DefaultTimezone {
		t.Fatalf("invalid should default to %s, got %q", service.DefaultTimezone, got)
	}
}

func TestIsLocalReminderHour(t *testing.T) {
	// 2026-06-09 13:00 UTC = 08:00 America/Lima (UTC-5, no DST)
	now := mustParseTime(t, "2026-06-09T13:00:00Z")
	if !service.IsLocalReminderHour(now, "America/Lima", 8, service.DefaultTimezone) {
		t.Fatal("expected 8 AM in Lima")
	}
	if service.IsLocalReminderHour(now, "America/New_York", 8, service.DefaultTimezone) {
		t.Fatal("did not expect 8 AM in New York at this UTC instant")
	}
}

func mustParseTime(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t.Fatalf("parse time: %v", err)
	}
	return parsed
}
