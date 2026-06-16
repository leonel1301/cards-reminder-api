package service

import (
	"testing"
	"time"
)

func TestNormalizeTimezone(t *testing.T) {
	if got := NormalizeTimezone("America/New_York"); got != "America/New_York" {
		t.Fatalf("got %q", got)
	}
	if got := NormalizeTimezone(""); got != DefaultTimezone {
		t.Fatalf("empty should default to %s, got %q", DefaultTimezone, got)
	}
	if got := NormalizeTimezone("Invalid/Zone"); got != DefaultTimezone {
		t.Fatalf("invalid should default to %s, got %q", DefaultTimezone, got)
	}
}

func TestIsLocalReminderHour(t *testing.T) {
	// 2026-06-09 13:00 UTC = 08:00 America/Lima (UTC-5, no DST)
	now := mustParseTime(t, "2026-06-09T13:00:00Z")
	if !IsLocalReminderHour(now, "America/Lima", 8, DefaultTimezone) {
		t.Fatal("expected 8 AM in Lima")
	}
	if IsLocalReminderHour(now, "America/New_York", 8, DefaultTimezone) {
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
