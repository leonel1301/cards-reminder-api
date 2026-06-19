package i18n

import (
	"sort"
	"strconv"
	"strings"
)

// NormalizeLanguage maps a language tag to "en" or "es". Empty or unsupported values default to "es".
func NormalizeLanguage(language string) string {
	tag := NormalizeLanguageTag(language)
	if tag == "en" {
		return "en"
	}
	return "es"
}

// NormalizeLanguageTag extracts the primary language subtag, or "" if input is empty.
func NormalizeLanguageTag(language string) string {
	language = strings.TrimSpace(strings.ToLower(language))
	if language == "" {
		return ""
	}
	if idx := strings.Index(language, "-"); idx > 0 {
		language = language[:idx]
	}
	return language
}

// ParseAcceptLanguage parses an Accept-Language header and returns "en" or "es".
func ParseAcceptLanguage(header string) string {
	header = strings.TrimSpace(header)
	if header == "" {
		return "es"
	}

	type candidate struct {
		lang string
		q    float64
	}

	candidates := make([]candidate, 0, 2)
	for _, part := range strings.Split(header, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		segments := strings.Split(part, ";")
		tag := NormalizeLanguageTag(segments[0])
		if tag != "en" && tag != "es" {
			continue
		}

		q := 1.0
		for _, segment := range segments[1:] {
			segment = strings.TrimSpace(segment)
			if !strings.HasPrefix(segment, "q=") {
				continue
			}
			if parsed, err := strconv.ParseFloat(strings.TrimPrefix(segment, "q="), 64); err == nil {
				q = parsed
			}
		}

		candidates = append(candidates, candidate{lang: tag, q: q})
	}

	if len(candidates) == 0 {
		return "es"
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].q > candidates[j].q
	})

	return candidates[0].lang
}
