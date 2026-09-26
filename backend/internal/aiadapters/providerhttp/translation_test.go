package providerhttp

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestTranslationRequestContainsOnlyReviewedTextAndInstructions(t *testing.T) {
	basis := "A stone bridge; the place is uncertain.\nDo not infer a location."
	raw, err := EncodeTranslation("selected-model", basis, "uk", "strict", "max_tokens", 4096)
	if err != nil {
		t.Fatal("translation encoding unavailable", err)
	}
	var payload struct {
		Model    string `json:"model"`
		Messages []struct {
			Role    string               `json:"role"`
			Content []ChatMessageContent `json:"content"`
		} `json:"messages"`
		Format ResponseFormat `json:"response_format"`
		Tokens int            `json:"max_tokens"`
	}
	if json.Unmarshal(raw, &payload) != nil || payload.Model != "selected-model" || payload.Tokens != 4096 || payload.Format.Type != "json_schema" {
		t.Fatal("incorrect provider request", string(raw))
	}
	if len(payload.Messages) != 2 || len(payload.Messages[1].Content) != 1 || payload.Messages[1].Content[0].Text != basis {
		t.Fatal("reviewed basis altered")
	}
	if !strings.Contains(payload.Messages[0].Content[0].Text, "uk") || strings.Contains(string(raw), "image_url") || len(raw) > 64<<10 {
		t.Fatal("invalid text-only request")
	}
}

func TestTranslationResponseRejectsUnusableOutput(t *testing.T) {
	for _, content := range []string{
		`{"language":"fr","status":"complete","text":"text"}`,
		`{"language":"en","status":"complete","text":"text","latitude":0}`,
		`{"language":"en","language":"en","status":"complete","text":"text"}`,
		`{"language":"en","status":"complete","text":""}`,
		`{"language":"en","status":"complete","text":null}`,
		`{"language":"en","status":"unavailable","text":"invented"}`,
		`{"language":"en","status":"unknown","text":null}`,
		`{"language":"en","status":"complete","text":"` + strings.Repeat("x", (16<<10)+1) + `"}`,
		`null`, `{broken`,
	} {
		if _, err := ParseTranslation(translationResponse(t, content), "en"); err != ErrTranslation {
			t.Fatal("unusable output accepted", err)
		}
	}
	valid := translationResponse(t, `{"language":"en","status":"complete","text":"text"}`)
	for _, raw := range [][]byte{
		[]byte(strings.Replace(string(valid), `"stop"`, `"length"`, 1)),
		[]byte(strings.Replace(string(valid), `"role":"assistant"`, `"role":"assistant","refusal":"refused"`, 1)),
		[]byte(strings.Repeat(" ", 64<<10) + string(valid)),
		append(valid, 255),
	} {
		if _, err := ParseTranslation(raw, "en"); err != ErrTranslation {
			t.Fatal("unsafe envelope accepted")
		}
	}
}

func translationResponse(t *testing.T, content string) []byte {
	t.Helper()
	raw, err := json.Marshal(map[string]any{"choices": []any{map[string]any{"finish_reason": "stop", "message": map[string]any{"role": "assistant", "content": content}}}})
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestTranslationResponseRetainsExactLanguageAndUnavailableState(t *testing.T) {
	out, err := ParseTranslation(translationResponse(t, `{"language":"uk","status":"complete","text":"Невідомий міст.\n Місце невідоме."}`), "uk")
	if err != nil || out.Language != "uk" || out.Status != "complete" || out.Text == nil || *out.Text != "Невідомий міст.\n Місце невідоме." {
		t.Fatal("valid translation lost", out, err)
	}
	out, err = ParseTranslation(translationResponse(t, `{"language":"en","status":"unavailable","text":null}`), "en")
	if err != nil || out.Status != "unavailable" || out.Text != nil {
		t.Fatal("unavailable outcome lost", out, err)
	}
}

func TestTranslationRequestRejectsUnsafeOrOverLimitValues(t *testing.T) {
	for _, tc := range []struct {
		model, basis, tag, format, field string
		tokens                           int64
	}{
		{"", "text", "en", "strict", "max_tokens", 100},
		{"model", "", "en", "strict", "max_tokens", 100},
		{"model", strings.Repeat("x", (16<<10)+1), "en", "strict", "max_tokens", 100},
		{"model", strings.Repeat("<", 16<<10), "en", "strict", "max_tokens", 100},
		{"model", string([]byte{255}), "en", "strict", "max_tokens", 100},
		{"model", "text", "bad tag", "strict", "max_tokens", 100},
		{"model", "text", "en", "auto", "max_tokens", 100},
		{"model", "text", "en", "strict", "model", 100},
		{"model", "text", "en", "json", "max_tokens", 0},
	} {
		if _, err := EncodeTranslation(tc.model, tc.basis, tc.tag, tc.format, tc.field, tc.tokens); err != ErrTranslation {
			t.Fatal("invalid request encoded", err)
		}
	}
}
