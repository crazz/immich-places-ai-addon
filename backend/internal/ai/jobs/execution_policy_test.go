package jobs

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func executionPolicyFixture() ExecutionPolicy {
	return ExecutionPolicy{Version: "execution-v1", Binding: ExecutionBinding{Owner: "owner", Installation: "00000000-0000-4000-8000-000000000001", Profile: "profile", Revision: 1, Model: "model", EgressFingerprint: strings.Repeat("a", 64)}, OutputField: "max_completion_tokens", MaxInputTokens: 100000, MaxOutputTokens: 4000, MaxRequestBytes: 15 << 20, MaxImageBytes: 10 << 20, EvidenceRef: "synthetic-attestation"}
}

func TestExecutionPolicyCopiesTariffsAndRedactsDiagnostics(t *testing.T) {
	p := executionPolicyFixture()
	input, output := int64(100), int64(200)
	p.Currency, p.InputMicrosPerMillion, p.OutputMicrosPerMillion = "USD", &input, &output
	raw, _ := json.Marshal([]ExecutionPolicy{p})
	policies, err := ParseExecutionPolicies(string(raw))
	if err != nil {
		t.Fatal(err)
	}
	copy, digest, ok := policies.Find(p.Binding)
	if !ok {
		t.Fatal("policy missing")
	}
	*copy.InputMicrosPerMillion = 999
	*copy.OutputMicrosPerMillion = 999
	again, next, ok := policies.Find(p.Binding)
	if !ok || next != digest || *again.InputMicrosPerMillion != 100 || *again.OutputMicrosPerMillion != 200 {
		t.Fatal("caller changed policy tariff or identity")
	}
	for _, private := range []string{p.EvidenceRef, p.Binding.Installation, p.Binding.EgressFingerprint} {
		if strings.Contains(fmt.Sprintf("%v %#v %v %#v", policies, policies, p, p), private) {
			t.Fatal("attestation details leaked in diagnostics")
		}
	}
}

func TestExecutionPolicyMatchesOnlyExactAttestedBinding(t *testing.T) {
	input := executionPolicyFixture()
	raw, err := json.Marshal([]ExecutionPolicy{input})
	if err != nil {
		t.Fatal(err)
	}
	policies, err := ParseExecutionPolicies(string(raw))
	if err != nil {
		t.Fatal(err)
	}
	p, digest, found := policies.Find(input.Binding)
	if !found || p.OutputField != input.OutputField || p.MaxInputTokens != input.MaxInputTokens || len(digest) != 64 {
		t.Fatal("exact policy unavailable")
	}
	other := input.Binding
	other.Revision++
	if _, _, found := policies.Find(other); found {
		t.Fatal("policy silently followed revision")
	}
	other = input.Binding
	other.Owner = "foreign"
	if _, _, found := policies.Find(other); found {
		t.Fatal("foreign policy matched")
	}
}

func TestExecutionPolicyRejectsIncompleteAndUnboundedAttestations(t *testing.T) {
	for _, mutate := range []func(*ExecutionPolicy){
		func(p *ExecutionPolicy) { p.Version = "unknown" },
		func(p *ExecutionPolicy) { p.Binding.Owner = "" },
		func(p *ExecutionPolicy) { p.Binding.Installation = "bad" },
		func(p *ExecutionPolicy) { p.Binding.Profile = "bad\nprofile" },
		func(p *ExecutionPolicy) { p.Binding.Revision = 0 },
		func(p *ExecutionPolicy) { p.Binding.Model = "" },
		func(p *ExecutionPolicy) { p.Binding.EgressFingerprint = "bad" },
		func(p *ExecutionPolicy) { p.OutputField = "max_guessed_tokens" },
		func(p *ExecutionPolicy) { p.MaxInputTokens = 0 },
		func(p *ExecutionPolicy) { p.MaxOutputTokens = 1000000001 },
		func(p *ExecutionPolicy) { p.MaxRequestBytes = (15 << 20) + 1 },
		func(p *ExecutionPolicy) { p.MaxImageBytes = (10 << 20) + 1 },
		func(p *ExecutionPolicy) { p.EvidenceRef = "" },
		func(p *ExecutionPolicy) { p.Currency = "USD" },
	} {
		p := executionPolicyFixture()
		mutate(&p)
		raw, err := json.Marshal([]ExecutionPolicy{p})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := ParseExecutionPolicies(string(raw)); err == nil {
			t.Fatal("incomplete/unbounded policy accepted", string(raw))
		}
	}
	valid, _ := json.Marshal(executionPolicyFixture())
	for _, raw := range []string{"null", `[{"unknown":true}]`, "[" + string(valid) + "," + string(valid) + "]", "[] []"} {
		if _, err := ParseExecutionPolicies(raw); err == nil {
			t.Fatal("invalid policy collection accepted")
		}
	}
}

func TestExecutionPolicyRejectsAmbiguousJSON(t *testing.T) {
	raw, _ := json.Marshal([]ExecutionPolicy{executionPolicyFixture()})
	for _, bad := range []string{
		strings.Replace(string(raw), `"version":"execution-v1"`, `"version":"unknown","Version":"execution-v1"`, 1),
		strings.Replace(string(raw), `"owner":"owner"`, `"owner":"foreign","Owner":"owner"`, 1),
		strings.Replace(string(raw), `"synthetic-attestation"`, string([]byte{'"', 0xff, '"'}), 1),
	} {
		if _, err := ParseExecutionPolicies(bad); err == nil {
			t.Fatal("ambiguous or malformed policy JSON accepted")
		}
	}
}
