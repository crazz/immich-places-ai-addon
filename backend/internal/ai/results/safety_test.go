package results

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestFailuresAreBoundedSortedAndNeverEchoPayloads(t *testing.T) {
	data := mutateFixture(t, "synthetic", func(m map[string]any) {
		template := firstCandidate(m)
		candidates := []any{}
		for i := 0; i < 20; i++ {
			candidate := map[string]any{}
			for key, value := range template {
				candidate[key] = value
			}
			candidate["id"] = fmt.Sprintf("candidate-%d", i)
			refs := []any{}
			for j := 0; j < 100; j++ {
				refs = append(refs, fmt.Sprintf("private-source-secret-%d", j))
			}
			candidate["evidence_refs"] = refs
			candidates = append(candidates, candidate)
		}
		m["candidates"] = candidates
		m["selected_candidate_id"] = "candidate-0"
	})
	v := testValidator(t)
	ctx := visualContext("en", "uk")
	var previous string
	for i := 0; i < 10; i++ {
		rejected := requireFailure(t, v, data, ctx, "semantic_violation")
		encoded, err := json.Marshal(rejected)
		if err != nil {
			t.Fatal(err)
		}
		if len(rejected.Findings) > 20 || len(encoded) > 4096 || strings.Contains(string(encoded), "private-source") || strings.Contains(string(encoded), "50.001") {
			t.Fatalf("unbounded/leaky failure: findings=%d bytes=%d", len(rejected.Findings), len(encoded))
		}
		for j := 1; j < len(rejected.Findings); j++ {
			before, after := rejected.Findings[j-1], rejected.Findings[j]
			if before.Path > after.Path || before.Path == after.Path && before.Code > after.Code {
				t.Fatal("findings not sorted")
			}
		}
		if i > 0 && string(encoded) != previous {
			t.Fatal("nondeterministic failure")
		}
		previous = string(encoded)
	}
}

func TestZeroValuesCannotMasqueradeAsValidatedResults(t *testing.T) {
	var empty Proposal
	if _, err := empty.MarshalJSON(); err == nil {
		t.Fatal("zero proposal serialized")
	}
	if _, err := empty.Data(); err == nil {
		t.Fatal("zero proposal returned typed data")
	}
	if _, err := empty.PolicyVersion(); err == nil {
		t.Fatal("zero proposal has validation policy")
	}
	for _, validator := range []*Validator{nil, {}} {
		requireFailure(t, validator, fixture(t, "unknown"), visualContext("en"), "unavailable")
	}
}

func TestValidatedProposalOwnsAllReturnedData(t *testing.T) {
	v := testValidator(t)
	input := fixture(t, "synthetic")
	ctx := visualContext("en", "uk")
	proposal, err := v.Validate(input, ctx)
	if err != nil {
		t.Fatal(err)
	}
	before, err := proposal.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	expected := string(before)
	for i := range input {
		input[i] = 'x'
	}
	ctx.Languages[0] = "fr"
	data, err := proposal.Data()
	if err != nil {
		t.Fatal(err)
	}
	data.Observations[0].Text = "changed"
	data.Candidates[0].EvidenceRefs[0] = "changed"
	data.Candidates[0].CameraLocation.Latitude = "-90"
	*data.Descriptions[0].Text = "changed"
	*data.SelectedCandidateID = "changed"
	data.Warnings[0] = "changed"
	before[0] = 'x'
	after, err := proposal.MarshalJSON()
	if err != nil || string(after) != expected {
		t.Fatal("caller mutation changed validated data", err)
	}
	policy, err := proposal.PolicyVersion()
	if err != nil || policy != "analysis-result-v1" {
		t.Fatal("missing semantic policy", policy, err)
	}
}

func TestModelCannotInjectEnvelopeOrLeakUnknownKeys(t *testing.T) {
	v := testValidator(t)
	for _, key := range []string{"user_id", "installation_id", "asset_id", "job_id", "source_authorization", "verified", "approval", "write_plan", "private-secret-property"} {
		data := mutateFixture(t, "unknown", func(m map[string]any) { m[key] = "https://private.invalid/secret?coordinate=50.001" })
		rejected := requireFailure(t, v, data, visualContext("en"), "schema_violation")
		encoded, err := json.Marshal(rejected)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(encoded), key) || strings.Contains(string(encoded), "private") || strings.Contains(string(encoded), "50.001") {
			t.Fatal("unknown key/value leaked", string(encoded))
		}
	}
}
