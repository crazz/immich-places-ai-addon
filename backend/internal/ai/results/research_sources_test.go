package results

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestResearchRetainsAnswerSources(t *testing.T) {
	data := researchFixture(t, func(m map[string]any) {
		m["sources"] = []any{map[string]any{"id": "reference", "url": "https://example.org/reference", "title": "Reference photo", "relevance": "The tower and shoreline match."}}
		firstCandidate(m)["source_refs"] = []any{"reference"}
	})
	proposal, err := testValidator(t).Validate(data, researchContext())
	if err != nil {
		t.Fatal("answer reference rejected", err)
	}
	raw, err := json.Marshal(proposal)
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatal(err)
	}
	sources, ok := document["sources"].([]any)
	if !ok || len(sources) != 1 || sources[0].(map[string]any)["url"] != "https://example.org/reference" {
		t.Fatal("answer reference lost")
	}
}

func TestResearchBoundsAnswerSourceBytesAndIdentities(t *testing.T) {
	data := researchFixture(t, func(m map[string]any) {
		m["sources"] = []any{map[string]any{"id": "reference", "url": strings.Repeat("é", 1025), "title": nil, "relevance": "Comparison."}}
	})
	requireFailure(t, testValidator(t), data, researchContext(), "semantic_violation")
}

func TestResearchKeepsEstimateWithMissingReference(t *testing.T) {
	data := researchFixture(t, func(m map[string]any) { firstCandidate(m)["source_refs"] = []any{"missing"} })
	proposal, err := testValidator(t).Validate(data, researchContext())
	if err != nil {
		t.Fatal("missing optional link discarded coordinates", err)
	}
	doc, err := proposal.Data()
	if err != nil || doc.Candidates[0].CameraLocation == nil {
		t.Fatal("coordinates lost", err)
	}
}

func TestResearchRejectsDuplicateAnswerSourceIdentity(t *testing.T) {
	data := researchFixture(t, func(m map[string]any) {
		source := map[string]any{"id": "same", "url": "https://example.org", "title": nil, "relevance": "Reference."}
		m["sources"] = []any{source, source}
	})
	requireFailure(t, testValidator(t), data, researchContext(), "semantic_violation")
}

func TestResearchBoundsSourceTextBytes(t *testing.T) {
	for _, field := range []string{"title", "relevance"} {
		data := researchFixture(t, func(m map[string]any) {
			source := map[string]any{"id": "reference", "url": "https://example.org", "title": nil, "relevance": "Reference."}
			limit := 512
			if field == "relevance" {
				limit = 2000
			}
			source[field] = strings.Repeat("é", limit/2+1)
			m["sources"] = []any{source}
		})
		requireFailure(t, testValidator(t), data, researchContext(), "semantic_violation")
	}
}

func TestResearchContextStillRequiresAuthorizedInput(t *testing.T) {
	data := researchFixture(t, func(m map[string]any) {
		m["observations"].([]any)[0].(map[string]any)["kind"] = "provided_context"
		firstCandidate(m)["camera_direction"] = nil
		firstCandidate(m)["source_refs"] = []any{"hint"}
	})
	ctx := researchContext()
	ctx.Sources = []Source{{ID: "hint"}}
	if _, err := testValidator(t).Validate(data, ctx); err != nil {
		t.Fatal("displayed hint rejected", err)
	}
	ctx.Sources = nil
	requireFailure(t, testValidator(t), data, ctx, "semantic_violation")
}
