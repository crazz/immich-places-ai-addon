package analysis_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"immich-places-backend/internal/ai/analysis"
	"immich-places-backend/internal/aiadapters/providerhttp"
)

func TestVisualAlwaysValidatesOutcomeGeometryProvenanceAndLanguages(t *testing.T) {
	for _, kind := range []string{"located", "ambiguous", "bad-coordinate", "invented-source", "missing-language", "unsupported-direction"} {
		t.Run(kind, func(t *testing.T) {
			raw, err := os.ReadFile("../results/testdata/ai-analysis-result.synthetic.json")
			if err != nil {
				t.Fatal(err)
			}
			var doc map[string]any
			if json.Unmarshal(raw, &doc) != nil {
				t.Fatal("fixture")
			}
			candidates := doc["candidates"].([]any)
			candidate := candidates[0].(map[string]any)
			switch kind {
			case "ambiguous":
				copyRaw, _ := json.Marshal(candidate)
				var second map[string]any
				_ = json.Unmarshal(copyRaw, &second)
				second["id"] = "candidate-2"
				doc["candidates"] = append(candidates, second)
				doc["outcome"] = "ambiguous"
				doc["selected_candidate_id"] = nil
			case "bad-coordinate":
				candidate["camera_location"].(map[string]any)["latitude"] = 91
			case "invented-source":
				candidate["source_refs"] = []string{"private-foreign-source"}
			case "missing-language":
				doc["descriptions"] = doc["descriptions"].([]any)[:1]
			case "unsupported-direction":
				candidate["camera_direction"] = map[string]any{"azimuth_deg": 30, "reference": "true_north", "uncertainty_deg": nil, "method": "known_viewpoint_alignment"}
			}
			content, _ := json.Marshal(doc)
			_, req, _, _ := visualFixture(t)
			req.Languages = []string{"en", "uk"}
			calls := 0
			runner, _ := analysis.New(analysis.Protocol{Encode: providerhttp.EncodeVisual, Parse: providerhttp.ParseVisual}, func(ctx context.Context, r analysis.DispatchRequest) (analysis.DispatchReply, error) {
				if err := r.Authorize(ctx); err != nil {
					return analysis.DispatchReply{}, err
				}
				calls++
				body, _ := json.Marshal(map[string]any{"choices": []any{map[string]any{"finish_reason": "stop", "message": map[string]any{"role": "assistant", "content": string(content)}}}})
				return analysis.DispatchReply{Body: body}, nil
			})
			result, err := runner.RunVisual(context.Background(), req)
			if kind == "located" || kind == "ambiguous" {
				if err != nil || result == nil {
					t.Fatal(err)
				}
				accepted, _ := result.Proposal().Data()
				if accepted.Outcome != kind {
					t.Fatal("outcome changed")
				}
			} else if err == nil || result != nil {
				t.Fatal("invalid output accepted")
			}
			if calls != 1 {
				t.Fatal("hidden repair call")
			}
		})
	}
}
