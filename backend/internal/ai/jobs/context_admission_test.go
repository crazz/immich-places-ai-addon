package jobs

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestAdmissionRequiresExactContextConsent(t *testing.T) {
	cfg := map[string]any{"selectionToken": "11111111-1111-4111-8111-111111111111", "profileId": "p", "revision": 1, "mode": "context-assisted", "format": "strict", "languages": []string{"uk"}, "primaryLanguage": "uk", "policyId": strings.Repeat("a", 64), "limits": map[string]any{"maxCalls": 1, "maxTokens": 100, "outputTokens": 10}, "context": map[string]any{"version": "context-v1", "classes": []string{}, "hint": ""}}
	raw, _ := json.Marshal(map[string]any{"configuration": cfg, "consent": map[string]any{"version": "image-consent-v1", "image": true, "configuration": cfg}, "idempotencyKey": "context-first"})
	var req Admission
	if err := DecodeRequest(raw, &req); err != nil {
		t.Fatal("explicit empty Context consent rejected", err)
	}
	normalized, _, err := NormalizeAdmission(req)
	if err != nil || normalized.Configuration.Mode != "context-assisted" {
		t.Fatal("valid Context consent unavailable", err)
	}
}
