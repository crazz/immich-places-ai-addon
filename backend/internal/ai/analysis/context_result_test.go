package analysis_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"immich-places-backend/internal/ai/analysis"
	"immich-places-backend/internal/ai/contextual"
	"immich-places-backend/internal/ai/results"
	"immich-places-backend/internal/aiadapters/providerhttp"
)

func contextResponseRunner(t *testing.T, content []byte) (*analysis.Runner, *int) {
	t.Helper()
	calls := 0
	runner, err := analysis.New(analysis.Protocol{Encode: providerhttp.EncodeVisual, Parse: providerhttp.ParseVisual}, func(ctx context.Context, in analysis.DispatchRequest) (analysis.DispatchReply, error) {
		if err := in.Authorize(ctx); err != nil {
			return analysis.DispatchReply{}, err
		}
		calls++
		body, _ := json.Marshal(map[string]any{"choices": []any{map[string]any{"finish_reason": "stop", "message": map[string]any{"role": "assistant", "content": string(content)}}}})
		return analysis.DispatchReply{Body: body}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return runner, &calls
}

func TestEmptyContextKeepsRequestedModeThroughResult(t *testing.T) {
	_, req, _, _ := visualFixture(t)
	binding := contextBundle(t, req).Info().Binding
	var err error
	req.Context, err = contextual.Build(contextual.Input{Binding: binding, Consent: contextual.Consent{Binding: binding, Version: contextual.Version, Classes: []contextual.Class{contextual.Neighbors}}, Window: time.Hour})
	if err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile("../results/testdata/ai-analysis-result.unknown.json")
	if err != nil {
		t.Fatal(err)
	}
	runner, calls := contextResponseRunner(t, content)
	result, err := runner.RunContext(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	info := result.Info()
	if info.Mode != results.ContextAssisted || info.Context == nil || len(info.Context.Sources) != 0 || len(info.Context.Omissions) != 2 || *calls != 1 {
		t.Fatal("empty Context mode/provenance lost")
	}
}

func TestContextResultOwnsExactPrivateProvenance(t *testing.T) {
	_, req, _, _ := visualFixture(t)
	req.Context = contextBundle(t, req)
	content, err := os.ReadFile("../results/testdata/ai-analysis-result.unknown.json")
	if err != nil {
		t.Fatal(err)
	}
	runner, _ := contextResponseRunner(t, content)
	result, err := runner.RunContext(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	info := result.Info()
	if info.Mode != results.ContextAssisted || info.PromptVersion != analysis.ContextPromptVersion || info.Context == nil || info.Context.Digest != req.Context.Info().Digest || info.Context.Version != contextual.Version {
		t.Fatal("context provenance lost")
	}
	info.Context.Sources[0].ID = "changed"
	info.Context.Consent.Classes[0] = contextual.Neighbors
	info.Context.Binding.Owner = "changed"
	info.Context.Omissions = append(info.Context.Omissions, "changed")
	next := result.Info()
	if next.Context.Sources[0].ID != "hint" || next.Context.Consent.Classes[0] != contextual.Hint || next.Context.Binding.Owner != req.Owner || len(next.Context.Omissions) != 0 {
		t.Fatal("metadata aliases result")
	}
	for _, private := range []string{req.Owner, "Beside a bridge", "selection-private"} {
		if strings.Contains(fmt.Sprintf("%v %#v", result, result), private) {
			t.Fatal("private diagnostics")
		}
	}
}
