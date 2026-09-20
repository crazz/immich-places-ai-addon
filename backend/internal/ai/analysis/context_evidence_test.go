package analysis_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"
)

func TestContextCanonicalEvidenceCannotInventSourceOrGeometryAuthority(t *testing.T) {
	for _, claim := range []string{"valid", "invented", "source-radius", "extent", "alignment", "language", "geometry"} {
		t.Run(claim, func(t *testing.T) {
			_, req, _, reservations := visualFixture(t)
			req.Context = contextBundle(t, req)
			req.Languages = []string{"en", "uk"}
			content, err := os.ReadFile("../results/testdata/ai-analysis-result.synthetic.json")
			if err != nil {
				t.Fatal(err)
			}
			var doc map[string]any
			if err := json.Unmarshal(content, &doc); err != nil {
				t.Fatal(err)
			}
			candidate := doc["candidates"].([]any)[0].(map[string]any)
			candidate["source_refs"] = []string{"hint"}
			doc["observations"] = append(doc["observations"].([]any), map[string]any{"id": "context-observation", "kind": "provided_context", "text": "User-supplied hint, not verification."})
			candidate["evidence_refs"] = []string{"obs-1", "context-observation"}
			camera := candidate["camera_location"].(map[string]any)
			switch claim {
			case "invented":
				candidate["source_refs"] = []string{"invented-source"}
			case "source-radius":
				camera["radius_basis"] = "source_reported"
			case "extent":
				camera["radius_basis"] = "context_extent"
			case "alignment":
				candidate["camera_direction"].(map[string]any)["method"] = "known_viewpoint_alignment"
			case "language":
				doc["descriptions"].([]any)[0].(map[string]any)["language"] = "de"
			case "geometry":
				camera["latitude"] = 91
			}
			content, err = json.Marshal(doc)
			if err != nil {
				t.Fatal(err)
			}
			runner, calls := contextResponseRunner(t, content)
			result, err := runner.RunContext(context.Background(), req)
			if claim == "valid" {
				if err != nil || result == nil {
					t.Fatal("valid separate context evidence rejected", err)
				}
				parsed, err := result.Proposal().Data()
				if err != nil || parsed.Outcome != "located" || parsed.Candidates[0].SourceRefs[0] != "hint" || parsed.Observations[1].Kind != "provided_context" {
					t.Fatal("evidence changed", err)
				}
			} else if err == nil || result != nil {
				t.Fatal("unsupported claim accepted")
			}
			if *calls != 1 || *reservations != 1 {
				t.Fatal("hidden retry/repair", *calls, *reservations)
			}
		})
	}
}
