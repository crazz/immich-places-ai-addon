package results

import (
	"strings"
	"unicode/utf8"

	"golang.org/x/text/language"
)

type validationContext struct {
	mode      Mode
	languages map[string]bool
	primary   string
	sources   map[string]Source
}

func validateContext(input Context) (validationContext, error) {
	invalid := func() (validationContext, error) {
		return validationContext{}, failure("invalid_context", "context", "/")
	}
	if input.Mode != Visual && input.Mode != ContextAssisted && input.Mode != Research {
		return invalid()
	}
	if len(input.Languages) < 1 || len(input.Languages) > 10 || len(input.Sources) > 100 || input.Mode == Visual && len(input.Sources) != 0 {
		return invalid()
	}
	result := validationContext{mode: input.Mode, languages: map[string]bool{}, sources: map[string]Source{}}
	for _, raw := range input.Languages {
		tag, ok := normalizeLanguage(raw)
		if !ok || result.languages[tag] {
			return invalid()
		}
		result.languages[tag] = true
	}
	primary, ok := normalizeLanguage(input.PrimaryLanguage)
	if !ok || !result.languages[primary] {
		return invalid()
	}
	result.primary = primary
	for _, source := range input.Sources {
		if !validID(source.ID) {
			return invalid()
		}
		if _, duplicate := result.sources[source.ID]; duplicate {
			return invalid()
		}
		result.sources[source.ID] = source
	}
	return result, nil
}

func normalizeLanguage(raw string) (string, bool) {
	if len(raw) == 0 || len(raw) > 128 {
		return "", false
	}
	for _, c := range raw {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-') {
			return "", false
		}
	}
	tag, err := language.Parse(raw)
	if err != nil || len(tag.String()) > 128 {
		return "", false
	}
	return tag.String(), true
}

func validID(id string) bool {
	return len(id) <= 128 && utf8.ValidString(id) && strings.TrimSpace(id) != ""
}
