package providerhttp

import (
	"encoding/json"
	"unicode/utf8"

	"immich-places-backend/internal/ai/analysis"
)

func ParseVisualUsage(body []byte) *analysis.Usage {
	if len(body) > 1<<20 || !utf8.Valid(body) {
		return nil
	}
	fields, err := visualObject(body)
	if err != nil {
		return nil
	}
	values, err := visualObject(fields["usage"])
	if err != nil {
		return nil
	}
	usage := &analysis.Usage{}
	found := false
	for name, target := range map[string]**int{"prompt_tokens": &usage.PromptTokens, "completion_tokens": &usage.CompletionTokens, "total_tokens": &usage.TotalTokens} {
		raw, ok := values[name]
		if !ok || string(raw) == "null" {
			continue
		}
		var value int
		if json.Unmarshal(raw, &value) != nil || value < 0 || int64(value) > 1_000_000_000_000 {
			return nil
		}
		*target = &value
		found = true
	}
	if !found {
		return nil
	}
	return usage
}
