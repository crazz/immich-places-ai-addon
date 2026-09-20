package providerhttp

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestVisualResponseRequiresOneOrdinaryCompleteAssistant(t *testing.T) {
	content := `{"schema_version":"1.0"}`
	envelope := func(finish, role string, extra map[string]any) []byte {
		message := map[string]any{"role": role, "content": content}
		for k, v := range extra {
			message[k] = v
		}
		data, _ := json.Marshal(map[string]any{"choices": []any{map[string]any{"finish_reason": finish, "message": message}}})
		return data
	}
	good := envelope("stop", "assistant", nil)
	parsed, err := ParseVisual(good)
	if err != nil || !bytes.Equal(parsed.Content, []byte(content)) {
		t.Fatal("complete response rejected", err)
	}
	cases := [][]byte{envelope("length", "assistant", nil), envelope("", "assistant", nil), envelope("stop", "user", nil), envelope("stop", "assistant", map[string]any{"refusal": "private refusal"}), envelope("stop", "assistant", map[string]any{"tool_calls": []any{}}), envelope("stop", "assistant", map[string]any{"function_call": map[string]any{}}), []byte(`{"choices":[]}`), []byte(`{"choices":[],"Choices":[]}`), append(good, good...)}
	for _, data := range cases {
		if _, err := ParseVisual(data); err == nil {
			t.Fatal("incomplete or ambiguous envelope accepted")
		}
	}
}

func TestVisualUsageRemainsOptionalBoundedAndIndependent(t *testing.T) {
	for _, usage := range []string{`{"prompt_tokens":12,"completion_tokens":4,"total_tokens":16}`, `null`, `{}`, `{"total_tokens":-1}`, `{"total_tokens":1000000001}`, `{"total_tokens":"private"}`, `{"total_tokens":1,"total_tokens":2}`} {
		raw := []byte(`{"choices":[{"finish_reason":"stop","message":{"role":"assistant","content":"{}"}}],"usage":` + usage + `}`)
		parsed, err := ParseVisual(raw)
		if err != nil {
			t.Fatal(err)
		}
		if usage == `{"prompt_tokens":12,"completion_tokens":4,"total_tokens":16}` {
			if parsed.Usage == nil || parsed.Usage.TotalTokens == nil || *parsed.Usage.TotalTokens != 16 {
				t.Fatal("valid usage missing")
			}
		} else if parsed.Usage != nil {
			t.Fatal("invalid usage should remain unknown")
		}
	}
}

func TestVisualResponseBoundsAndAmbiguityAreRejected(t *testing.T) {
	cases := [][]byte{make([]byte, (1<<20)+1), []byte{0xff}, []byte(`{"choices":[{},{}]}`), []byte(`{"choices":[{"finish_reason":"stop","finish_reason":"length","message":{"role":"assistant","content":"{}"}}]}`), []byte(`{"choices":[{"finish_reason":"stop","message":{"role":"assistant","content":42}}]}`)}
	for _, data := range cases {
		if _, err := ParseVisual(data); err == nil {
			t.Fatal("unbounded or ambiguous response accepted")
		}
	}
}
