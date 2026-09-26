package providerhttp

import (
	"encoding/json"
	"immich-places-backend/internal/ai/translations"
	"strings"
	"unicode/utf8"
)

func ParseTranslation(body []byte, tag string) (translations.Outcome, error) {
	if len(body) > 64<<10 || !utf8.Valid(body) {
		return translations.Outcome{}, ErrTranslation
	}
	reply, err := ParseVisual(body)
	if err != nil {
		return translations.Outcome{}, ErrTranslation
	}
	var out translations.Outcome
	fields, err := visualObject(reply.Content)
	if err != nil || len(fields) != 3 || fields["language"] == nil || fields["status"] == nil || fields["text"] == nil || json.Unmarshal(reply.Content, &out) != nil {
		return out, ErrTranslation
	}
	if out.Language != tag || (out.Status != "complete" && out.Status != "unavailable") {
		return translations.Outcome{}, ErrTranslation
	}
	if out.Status == "complete" && (out.Text == nil || strings.TrimSpace(*out.Text) == "" || len(*out.Text) > 16<<10 || !utf8.ValidString(*out.Text)) {
		return translations.Outcome{}, ErrTranslation
	}
	if out.Status == "unavailable" && out.Text != nil {
		return translations.Outcome{}, ErrTranslation
	}
	return out, nil
}
