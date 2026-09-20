package providerhttp

import "testing"

func TestProductionUsageCanExposeExcessBeyondLargestAttestedAllowance(t *testing.T) {
	usage := ParseVisualUsage([]byte(`{"choices":[],"usage":{"prompt_tokens":1000000001,"completion_tokens":0,"total_tokens":1000000001}}`))
	if usage == nil || usage.PromptTokens == nil || *usage.PromptTokens != 1000000001 || usage.CompletionTokens == nil || *usage.CompletionTokens != 0 {
		t.Fatal("above-policy usage hidden as unknown")
	}
	for _, raw := range []string{`{"usage":{"prompt_tokens":-1}}`, `{"usage":{"prompt_tokens":1.5}}`, `{"usage":{"prompt_tokens":1,"prompt_tokens":2}}`, `{"usage":null}`} {
		if ParseVisualUsage([]byte(raw)) != nil {
			t.Fatal("untrustworthy usage accepted")
		}
	}
}
