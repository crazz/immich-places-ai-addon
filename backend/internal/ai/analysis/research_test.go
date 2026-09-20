package analysis_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"testing"

	"immich-places-backend/internal/ai/analysis"
	"immich-places-backend/internal/ai/results"
	"immich-places-backend/internal/aiadapters/providerhttp"
)

func TestResearchSendsOneImageAndHintAndReturnsCoarseAnswer(t *testing.T) {
	_, req, _, reservations := visualFixture(t)
	req.Context = contextBundle(t, req)
	req.Languages = []string{"en", "uk"}
	content, err := os.ReadFile("../results/testdata/ai-analysis-result.synthetic.json")
	if err != nil {
		t.Fatal(err)
	}
	var answer map[string]any
	if err := json.Unmarshal(content, &answer); err != nil {
		t.Fatal(err)
	}
	answer["schema_version"] = "2.0"
	candidate := answer["candidates"].([]any)[0].(map[string]any)
	location := candidate["camera_location"].(map[string]any)
	location["estimated_radius_m"], location["radius_basis"], location["granularity"] = 5000, "model_estimate", "city"
	content, err = json.Marshal(answer)
	if err != nil {
		t.Fatal(err)
	}
	calls := 0
	runner, err := analysis.New(analysis.Protocol{Encode: providerhttp.EncodeResearch, Parse: providerhttp.ParseVisual}, func(ctx context.Context, in analysis.DispatchRequest) (analysis.DispatchReply, error) {
		if err := in.Authorize(ctx); err != nil {
			return analysis.DispatchReply{}, err
		}
		calls++
		for _, text := range []string{"Research", "Beside a bridge", "model_estimate", "2.0", "available research tools", "misleading hints"} {
			if !bytes.Contains(in.Body, []byte(text)) {
				t.Fatalf("missing %q", text)
			}
		}
		for _, secret := range []string{req.Owner, req.Installation, req.Asset, req.ProfileID, "selection-private", "source-private"} {
			if bytes.Contains(in.Body, []byte(secret)) {
				t.Fatal("private binding disclosed")
			}
		}
		if bytes.Count(in.Body, []byte("data:image/jpeg;base64,")) != 1 {
			t.Fatal("not exactly one prepared image")
		}
		raw, err := json.Marshal(map[string]any{"choices": []any{map[string]any{"finish_reason": "stop", "message": map[string]any{"role": "assistant", "content": string(content)}}}})
		return analysis.DispatchReply{Body: raw}, err
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := runner.RunResearch(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	doc, err := result.Proposal().Data()
	if err != nil || doc.Candidates[0].CameraLocation.EstimatedRadiusM.String() != "5000" {
		t.Fatal("coarse answer lost", err)
	}
	info := result.Info()
	if calls != 1 || *reservations != 1 || info.Mode != results.Research || info.SchemaVersion != "2.0" || info.PromptVersion != "research-v1" || info.Context == nil {
		t.Fatal("invalid Research handoff", info)
	}
}
