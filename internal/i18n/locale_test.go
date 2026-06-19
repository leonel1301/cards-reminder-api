package i18n

import "testing"

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
		if got := ParseAcceptLanguage(tt.header); got != tt.want {
			t.Errorf("ParseAcceptLanguage(%q) = %q, want %q", tt.header, got, tt.want)
		}
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
		if got := NormalizeLanguage(tt.input); got != tt.want {
			t.Errorf("NormalizeLanguage(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
