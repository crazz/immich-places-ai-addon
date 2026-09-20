package providerhttp

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"unicode/utf8"

	"immich-places-backend/internal/ai/analysis"
)

var ErrVisualResponse = errors.New("invalid or incomplete Visual response")

func ParseVisual(body []byte) (analysis.Response, error) {
	if len(body) > 1<<20 || !utf8.Valid(body) {
		return analysis.Response{}, ErrVisualResponse
	}
	root, err := visualObject(body)
	if err != nil {
		return analysis.Response{}, err
	}
	var choices []json.RawMessage
	if json.Unmarshal(root["choices"], &choices) != nil || len(choices) != 1 {
		return analysis.Response{}, ErrVisualResponse
	}
	choice, err := visualObject(choices[0])
	if err != nil {
		return analysis.Response{}, err
	}
	var finish string
	if json.Unmarshal(choice["finish_reason"], &finish) != nil || finish != "stop" {
		return analysis.Response{}, ErrVisualResponse
	}
	message, err := visualObject(choice["message"])
	if err != nil {
		return analysis.Response{}, err
	}
	var role, content string
	if json.Unmarshal(message["role"], &role) != nil || role != "assistant" || json.Unmarshal(message["content"], &content) != nil || content == "" {
		return analysis.Response{}, ErrVisualResponse
	}
	for _, field := range []string{"tool_calls", "function_call"} {
		if value := message[field]; len(value) > 0 && string(value) != "null" {
			return analysis.Response{}, ErrVisualResponse
		}
	}
	if value := message["refusal"]; len(value) > 0 && string(value) != "null" && string(value) != `""` {
		return analysis.Response{}, ErrVisualResponse
	}
	return analysis.Response{Content: []byte(content), Usage: visualUsage(root["usage"])}, nil
}

func visualObject(body []byte) (map[string]json.RawMessage, error) {
	decoder := json.NewDecoder(bytes.NewReader(body))
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		return nil, ErrVisualResponse
	}
	fields := map[string]json.RawMessage{}
	for decoder.More() {
		token, err = decoder.Token()
		key, ok := token.(string)
		if err != nil || !ok {
			return nil, ErrVisualResponse
		}
		lower := strings.ToLower(key)
		if _, exists := fields[lower]; exists || key != lower {
			return nil, ErrVisualResponse
		}
		var value json.RawMessage
		if decoder.Decode(&value) != nil {
			return nil, ErrVisualResponse
		}
		fields[key] = value
	}
	if token, err = decoder.Token(); err != nil || token != json.Delim('}') {
		return nil, ErrVisualResponse
	}
	if _, err = decoder.Token(); err != io.EOF {
		return nil, ErrVisualResponse
	}
	return fields, nil
}
