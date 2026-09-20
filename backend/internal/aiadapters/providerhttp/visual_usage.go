package providerhttp

import (
	"encoding/json"

	"immich-places-backend/internal/ai/analysis"
)

func visualUsage(data []byte) *analysis.Usage {
	fields, err := visualObject(data)
	if err != nil {
		return nil
	}
	usage := &analysis.Usage{}
	found := false
	for key, target := range map[string]**int{"prompt_tokens": &usage.PromptTokens, "completion_tokens": &usage.CompletionTokens, "total_tokens": &usage.TotalTokens} {
		raw, ok := fields[key]
		if !ok || string(raw) == "null" {
			continue
		}
		var value int
		if json.Unmarshal(raw, &value) != nil || value < 0 || value > 1000000000 {
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
