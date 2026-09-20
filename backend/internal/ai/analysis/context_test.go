package analysis_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"immich-places-backend/internal/ai/analysis"
	"immich-places-backend/internal/ai/contextual"
	"immich-places-backend/internal/aiadapters/providerhttp"
)

func contextBundle(t *testing.T, req analysis.Request) *contextual.Bundle {
	t.Helper()
	info, _ := req.Image.Info()
	binding := contextual.Binding{Owner: req.Owner, Installation: req.Installation, Asset: req.Asset, Selection: "selection-private", Profile: req.ProfileID, Revision: req.Revision, SourceDigest: info.Binding.SourceDigest}
	b, err := contextual.Build(contextual.Input{Binding: binding, Consent: contextual.Consent{Binding: binding, Version: contextual.Version, Classes: []contextual.Class{contextual.Hint}}, Hint: "Beside a bridge", Window: 6 * time.Hour})
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestContextAttemptSendsOnlyExactBundleOnce(t *testing.T) {
	_, req, _, reservations := visualFixture(t)
	req.Context = contextBundle(t, req)
	calls := 0
	content, err := os.ReadFile("../results/testdata/ai-analysis-result.unknown.json")
	if err != nil {
		t.Fatal(err)
	}
	runner, err := analysis.New(analysis.Protocol{Encode: providerhttp.EncodeVisual, Parse: providerhttp.ParseVisual}, func(ctx context.Context, in analysis.DispatchRequest) (analysis.DispatchReply, error) {
		if err := in.Authorize(ctx); err != nil {
			return analysis.DispatchReply{}, err
		}
		calls++
		if !bytes.Contains(in.Body, []byte("Context-assisted")) || !bytes.Contains(in.Body, []byte("Beside a bridge")) {
			t.Fatal("context missing from wire")
		}
		for _, secret := range []string{req.Owner, req.Installation, req.Asset, req.ProfileID, "selection-private", "source-private"} {
			if bytes.Contains(in.Body, []byte(secret)) {
				t.Fatal("private source binding disclosed")
			}
		}
		raw, _ := json.Marshal(map[string]any{"choices": []any{map[string]any{"finish_reason": "stop", "message": map[string]any{"role": "assistant", "content": string(content)}}}})
		return analysis.DispatchReply{Body: raw}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := runner.RunContext(context.Background(), req)
	if err != nil || result == nil {
		t.Fatal("context attempt failed", err)
	}
	doc, err := result.Proposal().Data()
	if err != nil || doc.Outcome != "unknown" || calls != 1 || *reservations != 1 {
		t.Fatal("incorrect outcome or call count", err, calls, *reservations)
	}
}

func TestVisualRejectsContextInsteadOfInheritingIt(t *testing.T) {
	runner, req, calls, reservations := visualFixture(t)
	req.Context = contextBundle(t, req)
	result, err := runner.RunVisual(context.Background(), req)
	if err == nil || result != nil || *calls != 0 || *reservations != 0 {
		t.Fatal("Visual accepted context", err)
	}
	req.Context = nil
	if result, err := runner.RunVisual(context.Background(), req); err != nil || result == nil || *calls != 1 {
		t.Fatal("ordinary Visual broken", err)
	}
}
