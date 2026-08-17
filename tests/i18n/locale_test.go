package i18n_test

import (
	"testing"

	"github.com/leonelortega/cards-reminder-api/internal/i18n"
)

func TestParseAcceptLanguage(t *testing.T) {
	tests := []struct {
		header string
		want   string
	}{
		{"", "es"},
		{"es", "es"},
		{"en", "en"},
		{"en-US", "en"},
		{"en-US,en;q=0.9,es;q=0.8", "en"},
		{"es-AR,en;q=0.8", "es"},
		{"fr,de", "es"},
		{"fr,en;q=0.5,es;q=0.4", "en"},
	}

	for _, tt := range tests {
		if got := i18n.ParseAcceptLanguage(tt.header); got != tt.want {
			t.Errorf("ParseAcceptLanguage(%q) = %q, want %q", tt.header, got, tt.want)
		}
	}
}

func TestFailedToGetPaymentCountMessages(t *testing.T) {
	if got := i18n.Error("en", i18n.ErrFailedToGetPaymentCount); got != "failed to get payment count" {
		t.Errorf("en = %q", got)
	}
	if got := i18n.Error("es", i18n.ErrFailedToGetPaymentCount); got != "no se pudo obtener el conteo de pagos" {
		t.Errorf("es = %q", got)
	}
}

func TestLessonProgressMessages(t *testing.T) {
	if got := i18n.Error("en", i18n.ErrFailedToListLessons); got != "failed to list completed lessons" {
		t.Errorf("list en = %q", got)
	}
	if got := i18n.Error("es", i18n.ErrFailedToMarkLesson); got != "no se pudo marcar la lección como completada" {
		t.Errorf("mark es = %q", got)
	}
}

func TestContractAnalyzeLimitMessages(t *testing.T) {
	if got := i18n.Error("en", i18n.ErrContractAnalyzeLimitReached); got != "beta limit reached: you can analyze up to 5 contracts" {
		t.Errorf("en = %q", got)
	}
	if got := i18n.Error("es", i18n.ErrContractAnalyzeLimitReached); got != "límite beta alcanzado: puedes analizar hasta 5 contratos" {
		t.Errorf("es = %q", got)
	}
}

func TestNormalizeLanguage(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", "es"},
		{"EN", "en"},
		{"en-US", "en"},
		{"es-AR", "es"},
		{"fr", "es"},
	}

	for _, tt := range tests {
		if got := i18n.NormalizeLanguage(tt.input); got != tt.want {
			t.Errorf("NormalizeLanguage(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
