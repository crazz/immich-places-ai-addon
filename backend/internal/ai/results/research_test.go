package results

import (
	"encoding/json"
	"testing"
)

func researchContext() Context {
	ctx := visualContext("en", "uk")
	ctx.Mode = "research"
	return ctx
}

func researchFixture(t *testing.T, change func(map[string]any)) []byte {
	t.Helper()
	return mutateFixture(t, "synthetic", func(m map[string]any) {
		m["schema_version"] = "2.0"
		camera(m)["radius_basis"] = "model_estimate"
		camera(m)["estimated_radius_m"] = 500
		change(m)
	})
}

func TestResearchRetains500MeterEstimate(t *testing.T) {
	data := researchFixture(t, func(map[string]any) {})
	proposal, err := testValidator(t).Validate(data, researchContext())
	if err != nil {
		t.Fatal("useful estimate rejected", err)
	}
	doc, err := proposal.Data()
	if err != nil {
		t.Fatal(err)
	}
	if doc.SchemaVersion != "2.0" || doc.Candidates[0].CameraLocation.EstimatedRadiusM.String() != "500" {
		t.Fatal("estimate was changed")
	}
}

func TestResearchRetainsCoarseCityEstimate(t *testing.T) {
	data := researchFixture(t, func(m map[string]any) {
		camera(m)["granularity"] = "city"
		camera(m)["estimated_radius_m"] = 5000
		firstCandidate(m)["support_summary"] = "Low confidence: an approximate city point is the best available estimate."
	})
	proposal, err := testValidator(t).Validate(data, researchContext())
	if err != nil {
		t.Fatal("coarse estimate rejected", err)
	}
	doc, err := proposal.Data()
	if err != nil {
		t.Fatal(err)
	}
	if doc.Outcome != "located" || doc.Candidates[0].CameraLocation.EstimatedRadiusM.String() != "5000" {
		t.Fatal("coarse estimate withheld")
	}
}

func TestResearchPreservesExistingValidationAndUncertainty(t *testing.T) {
	v := testValidator(t)
	for _, radius := range []any{nil, 0, 50000, 1e100} {
		data := researchFixture(t, func(m map[string]any) {
			camera(m)["granularity"] = "region"
			camera(m)["estimated_radius_m"] = radius
			if radius == nil {
				camera(m)["radius_basis"] = "unknown"
			}
		})
		if _, err := v.Validate(data, researchContext()); err != nil {
			t.Fatalf("region/unknown estimate %v: %v", radius, err)
		}
	}
	for _, outcome := range []string{"located", "ambiguous", "unknown"} {
		data := researchFixture(t, func(m map[string]any) {
			var alternative map[string]any
			raw, err := json.Marshal(firstCandidate(m))
			if err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(raw, &alternative); err != nil {
				t.Fatal(err)
			}
			alternative["id"] = "alternative"
			m["candidates"] = append(m["candidates"].([]any), alternative)
			m["outcome"] = outcome
			if outcome != "located" {
				m["selected_candidate_id"] = nil
			}
			if outcome == "unknown" {
				m["candidates"] = []any{}
				for _, value := range m["descriptions"].([]any) {
					d := value.(map[string]any)
					d["basis"], d["candidate_id"] = "scene_only", nil
				}
			}
		})
		p, err := v.Validate(data, researchContext())
		if err != nil {
			t.Fatal(outcome, err)
		}
		d, err := p.Data()
		if err != nil || d.Outcome != outcome {
			t.Fatal("outcome changed", err)
		}
	}
	for _, change := range []func(map[string]any){
		func(m map[string]any) { camera(m)["estimated_radius_m"] = -1 },
		func(m map[string]any) { camera(m)["latitude"] = 91 },
		func(m map[string]any) { m["unexpected"] = true },
	} {
		requireFailure(t, v, researchFixture(t, change), researchContext(), "schema_violation")
	}
	requireFailure(t, v, researchFixture(t, func(m map[string]any) { camera(m)["estimated_radius_m"] = json.Number("1e400") }), researchContext(), "limit_exceeded")
	requireFailure(t, v, fixture(t, "synthetic"), researchContext(), "unsupported_schema")
	requireFailure(t, v, researchFixture(t, func(map[string]any) {}), visualContext("en", "uk"), "unsupported_schema")
}

func TestResearchIdentifiesItsValidationPolicy(t *testing.T) {
	proposal, err := testValidator(t).Validate(researchFixture(t, func(map[string]any) {}), researchContext())
	if err != nil {
		t.Fatal(err)
	}
	version, err := proposal.PolicyVersion()
	if err != nil || version != "analysis-result-v2" {
		t.Fatalf("policy = %q, %v", version, err)
	}
}
