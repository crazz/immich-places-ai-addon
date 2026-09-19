package providerhttp_test

import (
	"errors"
	"strings"
	"testing"

	"immich-places-backend/internal/ai/capabilities"
	"immich-places-backend/internal/aiadapters/providerhttp"
)

func TestEncodeImageProbeUsesSelectedModel(t *testing.T) {
	body, err := providerhttp.EncodeImageProbe(
		"gpt-5.6-sol",
		"Describe the single visible geometric shape.",
		"data:image/jpeg;base64,/9j/",
	)
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	for _, want := range []string{`"model":"gpt-5.6-sol"`, `"stream":false`, "data:image/jpeg;base64,"} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %s in %s", want, text)
		}
	}
	if strings.Contains(text, "response_format") {
		t.Fatal("image probe must not set response_format")
	}
}

func TestEncodeJSONProbeRequestsJSONObjectMode(t *testing.T) {
	body, err := providerhttp.EncodeJSONProbe(
		"gpt-5.6-sol",
		"Reply with one JSON object.",
		"data:image/jpeg;base64,/9j/",
	)
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	for _, want := range []string{
		`"model":"gpt-5.6-sol"`,
		`"stream":false`,
		`"response_format":{"type":"json_object"}`,
		"data:image/jpeg;base64,",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %s in %s", want, text)
		}
	}
}

func TestEncodeStrictProbeRequestsClosedSchema(t *testing.T) {
	body, err := providerhttp.EncodeStrictProbe(
		"gpt-5.6-sol",
		"Reply with one JSON object.",
		"data:image/jpeg;base64,/9j/",
	)
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	for _, want := range []string{
		`"type":"json_schema"`,
		`"strict":true`,
		`"name":"capability_labels"`,
		`"additionalProperties":false`,
		`"enum":["blue","red","green"]`,
		`"enum":["circle","square","triangle"]`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %s in %s", want, text)
		}
	}
	if strings.Contains(text, `"blue"`) && strings.Contains(text, "selected") {
		t.Fatal("must not expose the selected fixture answer")
	}
}

func TestParseAssistantTextRejectsRefusalTruncationAndTools(t *testing.T) {
	cases := []struct {
		name    string
		body    string
		wantErr error
	}{
		{"refusal", `{"model":"m","choices":[{"finish_reason":"stop","message":{"role":"assistant","content":"","refusal":"no"}}]}`, capabilities.ErrRefusal},
		{"truncation", `{"model":"m","choices":[{"finish_reason":"length","message":{"role":"assistant","content":"color: blue"}}]}`, capabilities.ErrTruncation},
		{"tools", `{"model":"m","choices":[{"finish_reason":"stop","message":{"role":"assistant","content":"hi","tool_calls":[{"id":"1"}]}}]}`, capabilities.ErrToolResponse},
		{"tools_finish_reason", `{"model":"m","choices":[{"finish_reason":"tool_calls","message":{"role":"assistant","content":null,"tool_calls":[{"id":"call_1","type":"function","function":{"name":"lookup","arguments":"{}"}}]}}]}`, capabilities.ErrToolResponse},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := providerhttp.ParseAssistantText([]byte(tc.body))
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("err=%v, want %v", err, tc.wantErr)
			}
			if (tc.name == "tools" || tc.name == "tools_finish_reason") && capabilities.ClassifyParseFailure(err) != "tool_response" {
				t.Fatalf("tool response classified as %q", capabilities.ClassifyParseFailure(err))
			}
		})
	}
}
