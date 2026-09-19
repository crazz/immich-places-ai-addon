package providerhttp

import (
	"strings"
	"testing"

	"immich-places-backend/internal/ai/capabilities"
)

func TestUnknownOrInvalidUsageIsNeverReportedAsZero(t *testing.T) {
	for _, raw := range []string{
		`{}`, `{"usage":null}`, `{"usage":{}}`, `{"usage":{"total_tokens":null}}`,
		`{"usage":{"total_tokens":"12"}}`, `{"usage":{"total_tokens":1.5}}`,
		`{"usage":{"total_tokens":-1}}`, `{"usage":{"total_tokens":1e100}}`,
		`{"usage":{"total_tokens":12}} trailing`, strings.Repeat(" ", MaxCapabilityResponseBytes+1),
	} {
		if got := ParseUsage([]byte(raw)); got != nil {
			t.Fatalf("invalid/absent usage became known: %+v", got)
		}
	}
	partial := ParseUsage([]byte(`{"usage":{"total_tokens":0}}`))
	if partial == nil || partial.TotalTokens == nil || *partial.TotalTokens != 0 || partial.PromptTokens != nil {
		t.Fatalf("explicit zero and absent fields must remain distinct: %+v", partial)
	}
	full := ParseUsage([]byte(`{"usage":{"prompt_tokens":2,"completion_tokens":3,"total_tokens":5}}`))
	combined := capabilities.CombineUsage(full, partial)
	if combined == nil || combined.TotalTokens == nil || *combined.TotalTokens != 5 || combined.PromptTokens != nil {
		t.Fatalf("partial probe metadata must not invent component totals: %+v", combined)
	}
	if capabilities.CombineUsage(full, nil) != nil || capabilities.CombineUsage(nil, full) != nil {
		t.Fatal("an omitted probe prevents an exact attempt total")
	}
	max := int(^uint(0) >> 1)
	if capabilities.CombineUsage(&capabilities.Usage{TotalTokens: &max}, full) != nil {
		t.Fatal("overflow must stay unknown")
	}
}
