package providerhttp

import (
	"encoding/json"

	"immich-places-backend/internal/ai/capabilities"
)

func ParseUsage(body []byte) *capabilities.Usage {
	if len(body) > MaxCapabilityResponseBytes {
		return nil
	}
	var response struct {
		Usage *struct {
			PromptTokens     *int `json:"prompt_tokens"`
			CompletionTokens *int `json:"completion_tokens"`
			TotalTokens      *int `json:"total_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(body, &response); err != nil || response.Usage == nil {
		return nil
	}
	u := response.Usage
	for _, value := range []*int{u.PromptTokens, u.CompletionTokens, u.TotalTokens} {
		if value != nil && *value < 0 {
			return nil
		}
	}
	if u.PromptTokens == nil && u.CompletionTokens == nil && u.TotalTokens == nil {
		return nil
	}
	return &capabilities.Usage{PromptTokens: u.PromptTokens, CompletionTokens: u.CompletionTokens, TotalTokens: u.TotalTokens}
}
