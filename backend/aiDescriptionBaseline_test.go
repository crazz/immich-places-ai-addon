package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestAIDescriptionBaselineRejectsLossyAndMalformedText(t *testing.T) {
	for _, raw := range []string{`{"description":23}`, `{"description":{}}`, `{"description":"\ud800"}`, `{"description":"\udc00"}`, `{"description":"\ud800x"}`, `{"description":"` + strings.Repeat("a", (64<<10)+1) + `"}`, `{"description":"x","Description":"y"}`, `{"description":"` + string([]byte{0xff}) + `"}`} {
		if got := aiDescriptionBaseline(json.RawMessage(raw)); got != nil {
			t.Fatal("malformed description treated as readable text")
		}
	}
	for _, tc := range []struct{ raw, presence, value string }{
		{`null`, "absent", ""}, {`{}`, "absent", ""}, {`{"description":null}`, "null", ""},
		{`{"description":""}`, "value", ""}, {`{"description":"\ud83d\uddfa️\r\n"}`, "value", "🗺️\r\n"},
	} {
		got := aiDescriptionBaseline(json.RawMessage(tc.raw))
		if got == nil || got.Presence != tc.presence || got.Value != tc.value {
			t.Fatal("valid exact presence/value lost")
		}
	}
}
