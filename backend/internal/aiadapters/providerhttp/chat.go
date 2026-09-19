package providerhttp

import (
	"encoding/json"
	"errors"
	"fmt"

	"immich-places-backend/internal/ai/capabilities"
)

const (
	MaxCapabilityRequestBytes  = 256 << 10
	MaxCapabilityResponseBytes = 64 << 10
	SyntheticSchemaName        = "capability_labels"
)

type ChatMessageContent struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	ImageURL *struct {
		URL string `json:"url"`
	} `json:"image_url,omitempty"`
}

type ResponseFormat struct {
	Type       string          `json:"type"`
	JSONSchema *JSONSchemaSpec `json:"json_schema,omitempty"`
}

type JSONSchemaSpec struct {
	Name   string          `json:"name"`
	Strict bool            `json:"strict"`
	Schema json.RawMessage `json:"schema"`
}

type ChatCompletionsRequest struct {
	Model          string          `json:"model"`
	Stream         bool            `json:"stream"`
	Messages       []chatMessage   `json:"messages"`
	ResponseFormat *ResponseFormat `json:"response_format,omitempty"`
}

type chatMessage struct {
	Role    string               `json:"role"`
	Content []ChatMessageContent `json:"content"`
}

type ChatCompletionsResponse struct {
	Model   string `json:"model"`
	Choices []struct {
		FinishReason string `json:"finish_reason"`
		Message      struct {
			Role      string          `json:"role"`
			Content   json.RawMessage `json:"content"`
			ToolCalls json.RawMessage `json:"tool_calls"`
			Refusal   string          `json:"refusal"`
		} `json:"message"`
	} `json:"choices"`
}

var syntheticLabelSchema = json.RawMessage(`{"type":"object","additionalProperties":false,"required":["color","shape"],"properties":{"color":{"type":"string","enum":["blue","red","green"]},"shape":{"type":"string","enum":["circle","square","triangle"]}}}`)

func EncodeImageProbe(model, instruction, imageDataURL string) ([]byte, error) {
	return encodeCapabilityProbe(model, instruction, imageDataURL, nil)
}

func EncodeJSONProbe(model, instruction, imageDataURL string) ([]byte, error) {
	return encodeCapabilityProbe(model, instruction, imageDataURL, &ResponseFormat{Type: "json_object"})
}

func EncodeStrictProbe(model, instruction, imageDataURL string) ([]byte, error) {
	return encodeCapabilityProbe(model, instruction, imageDataURL, &ResponseFormat{
		Type: "json_schema",
		JSONSchema: &JSONSchemaSpec{
			Name:   SyntheticSchemaName,
			Strict: true,
			Schema: syntheticLabelSchema,
		},
	})
}

func encodeCapabilityProbe(model, instruction, imageDataURL string, format *ResponseFormat) ([]byte, error) {
	if model == "" || instruction == "" || imageDataURL == "" {
		return nil, errors.New("capability probe requires model, instruction and image")
	}
	payload := ChatCompletionsRequest{
		Model:  model,
		Stream: false,
		Messages: []chatMessage{{
			Role: "user",
			Content: []ChatMessageContent{
				{Type: "text", Text: instruction},
				{Type: "image_url", ImageURL: &struct {
					URL string `json:"url"`
				}{URL: imageDataURL}},
			},
		}},
		ResponseFormat: format,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	if len(raw) > MaxCapabilityRequestBytes {
		return nil, fmt.Errorf("capability probe request exceeds %d bytes", MaxCapabilityRequestBytes)
	}
	return raw, nil
}

func ParseAssistantText(body []byte) (text, reportedModel string, err error) {
	if len(body) > MaxCapabilityResponseBytes {
		return "", "", errors.New("capability response exceeds local ceiling")
	}
	var response ChatCompletionsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", "", errors.New("capability response is not valid JSON")
	}
	if len(response.Choices) != 1 {
		return "", "", errors.New("capability response must contain exactly one choice")
	}
	choice := response.Choices[0]
	hasToolCalls := len(choice.Message.ToolCalls) > 0 && string(choice.Message.ToolCalls) != "null"
	if choice.FinishReason == "tool_calls" || hasToolCalls {
		return "", "", capabilities.ErrToolResponse
	}
	if choice.Message.Refusal != "" {
		return "", "", capabilities.ErrRefusal
	}
	if choice.FinishReason != "stop" {
		return "", "", capabilities.ErrTruncation
	}
	var content string
	if err := json.Unmarshal(choice.Message.Content, &content); err != nil || content == "" {
		return "", "", errors.New("capability response missing assistant text")
	}
	return content, response.Model, nil
}
