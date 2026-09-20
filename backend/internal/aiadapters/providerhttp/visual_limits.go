package providerhttp

import (
	"encoding/base64"
	"encoding/json"
	"strings"
)

func EncodeLimitedVisual(model, instruction, imageDataURL, format, outputField string, outputTokens int64, maxRequestBytes, maxImageBytes int) ([]byte, error) {
	if (outputField != "max_tokens" && outputField != "max_completion_tokens") || outputTokens < 1 || outputTokens > 1_000_000_000 || maxRequestBytes < 1 || maxRequestBytes > 15<<20 || maxImageBytes < 1 || maxImageBytes > 10<<20 {
		return nil, ErrVisualRequest
	}
	encoded := strings.TrimPrefix(imageDataURL, "data:image/jpeg;base64,")
	size := base64.StdEncoding.DecodedLen(len(encoded))
	if strings.HasSuffix(encoded, "==") {
		size -= 2
	} else if strings.HasSuffix(encoded, "=") {
		size--
	}
	if size > maxImageBytes {
		return nil, ErrVisualRequest
	}
	raw, err := EncodeVisual(model, instruction, imageDataURL, format)
	if err != nil {
		return nil, err
	}
	defer clear(raw)
	var body map[string]json.RawMessage
	if err = json.Unmarshal(raw, &body); err != nil {
		return nil, ErrVisualRequest
	}
	body[outputField], _ = json.Marshal(outputTokens)
	limited, err := json.Marshal(body)
	if err != nil || len(limited) > maxRequestBytes {
		clear(limited)
		return nil, ErrVisualRequest
	}
	return limited, nil
}
