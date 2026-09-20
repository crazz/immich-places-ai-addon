package providerhttp

import (
	"encoding/json"
	"testing"
)

func TestLimitedVisualUsesOnlyTheAttestedOutputField(t *testing.T) {
	for _, field := range []string{"max_tokens", "max_completion_tokens"} {
		raw, err := EncodeLimitedVisual("model", "instruction", "data:image/jpeg;base64,AQID", "json", field, 123, 15<<20, 10<<20)
		if err != nil {
			t.Fatal(err)
		}
		var body map[string]json.RawMessage
		if json.Unmarshal(raw, &body) != nil || string(body[field]) != "123" {
			t.Fatal("output ceiling missing", string(raw))
		}
		other := "max_tokens"
		if field == other {
			other = "max_completion_tokens"
		}
		if _, ok := body[other]; ok {
			t.Fatal("unapproved output field")
		}
	}
	for _, args := range []struct {
		field          string
		output         int64
		request, image int
	}{
		{"unknown", 123, 15 << 20, 10 << 20}, {"max_tokens", 0, 15 << 20, 10 << 20}, {"max_tokens", 123, 1, 10 << 20}, {"max_tokens", 123, 15 << 20, 2},
	} {
		if _, err := EncodeLimitedVisual("model", "instruction", "data:image/jpeg;base64,AQID", "json", args.field, args.output, args.request, args.image); err == nil {
			t.Fatal("unattested envelope accepted", args)
		}
	}
}
