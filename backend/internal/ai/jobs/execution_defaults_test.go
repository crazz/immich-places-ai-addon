package jobs

import (
	"encoding/json"
	"testing"
)

func TestExecutionDefaultsKeepExactIdentityAndExplicitRestrictions(t *testing.T) {
	input := executionPolicyFixture()
	empty := ExecutionPolicies{}
	policy, id, ok := empty.Resolve(input.Binding)
	if !ok || !policy.Valid() || policy.EvidenceRef != "" || policy.Currency != "" {
		t.Fatal("invalid or falsely attested defaults")
	}
	_, again, _ := empty.Resolve(input.Binding)
	if id != again {
		t.Fatal("unstable default identity")
	}
	raw, _ := json.Marshal([]ExecutionPolicy{policy})
	if _, err := ParseExecutionPolicies(string(raw)); err == nil {
		t.Fatal("server defaults accepted as operator attestation")
	}
	raw, _ = json.Marshal([]ExecutionPolicy{input})
	explicit, err := ParseExecutionPolicies(string(raw))
	if err != nil {
		t.Fatal(err)
	}
	_, want, _ := explicit.Find(input.Binding)
	_, got, ok := explicit.Resolve(input.Binding)
	if !ok || got != want {
		t.Fatal("explicit policy replaced with defaults")
	}
	for _, change := range []func(*ExecutionBinding){
		func(b *ExecutionBinding) { b.Revision++ },
		func(b *ExecutionBinding) { b.Model = "changed-model" },
		func(b *ExecutionBinding) { b.Installation = "00000000-0000-4000-8000-000000000002" },
		func(b *ExecutionBinding) {
			b.EgressFingerprint = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
		},
	} {
		binding := input.Binding
		change(&binding)
		_, changedID, ok := empty.Resolve(binding)
		if !ok || changedID == id {
			t.Fatal("default identity ignored changed authority")
		}
		if _, _, ok := explicit.Resolve(binding); ok {
			t.Fatal("defaults bypassed stale explicit restrictions")
		}
	}
	invalid := input.Binding
	invalid.Owner = ""
	if _, _, ok := empty.Resolve(invalid); ok {
		t.Fatal("invalid default binding accepted")
	}
}
