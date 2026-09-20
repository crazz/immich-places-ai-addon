package providerhttp

import (
	"encoding/json"
	"errors"
	"strings"
	"unicode/utf8"

	"immich-places-backend/internal/ai/providers"
	"immich-places-backend/internal/ai/results"
)

var ErrVisualRequest = errors.New("invalid Visual request")

func EncodeVisual(model, instruction, imageDataURL, format string) ([]byte, error) {
	if model == "" || len(model) > 256 || !utf8.ValidString(model) || instruction == "" || len(instruction) > 64<<10 || !utf8.ValidString(instruction) || !strings.HasPrefix(imageDataURL, "data:image/jpeg;base64,") || len(imageDataURL) > 10<<20 || (format != "strict" && format != "json") {
		return nil, ErrVisualRequest
	}
	rf := &ResponseFormat{Type: "json_object"}
	if format == "strict" {
		rf = &ResponseFormat{Type: "json_schema", JSONSchema: &JSONSchemaSpec{Name: "visual_analysis", Strict: true, Schema: results.CanonicalSchema()}}
	}
	raw, err := json.Marshal(ChatCompletionsRequest{Model: model, Stream: false, Messages: []chatMessage{{Role: "user", Content: []ChatMessageContent{{Type: "text", Text: instruction}, {Type: "image_url", ImageURL: &struct {
		URL string `json:"url"`
	}{URL: imageDataURL}}}}}, ResponseFormat: rf})
	if err != nil || len(raw) > providers.MaxRequestBytes {
		return nil, ErrVisualRequest
	}
	return raw, nil
}
