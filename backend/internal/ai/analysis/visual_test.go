package analysis_test

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/jpeg"
	"os"
	"strings"
	"testing"

	"immich-places-backend/internal/ai/analysis"
	"immich-places-backend/internal/ai/images"
	"immich-places-backend/internal/aiadapters/providerhttp"
)

func visualFixture(t *testing.T) (*analysis.Runner, analysis.Request, *int, *int) {
	t.Helper()
	var encoded bytes.Buffer
	if err := jpeg.Encode(&encoded, image.NewRGBA(image.Rect(0, 0, 8, 4)), nil); err != nil {
		t.Fatal(err)
	}
	prepared, err := images.PrepareRaster(context.Background(), encoded.Bytes(), "image/jpeg", images.Binding{Owner: "owner-private", Installation: "install-private", Asset: "asset-private", SourceDigest: "source-private"}, images.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(prepared.Release)
	calls, reservations := 0, 0
	request := analysis.Request{Owner: "owner-private", Installation: "install-private", Asset: "asset-private", ProfileID: "profile-private", Revision: 1, Model: "model", Format: "strict", Languages: []string{"en"}, PrimaryLanguage: "en", Image: prepared, Guard: analysis.Guard{Authorize: func(context.Context) error { return nil }, Reserve: func(context.Context) error { reservations++; return nil }}}
	fixture, err := os.ReadFile("../results/testdata/ai-analysis-result.unknown.json")
	if err != nil {
		t.Fatal(err)
	}
	runner, err := analysis.New(analysis.Protocol{Encode: providerhttp.EncodeVisual, Parse: providerhttp.ParseVisual}, func(ctx context.Context, req analysis.DispatchRequest) (analysis.DispatchReply, error) {
		if err := req.Authorize(ctx); err != nil {
			return analysis.DispatchReply{}, err
		}
		calls++
		for _, secret := range []string{"owner-private", "install-private", "asset-private", "profile-private", "source-private"} {
			if bytes.Contains(req.Body, []byte(secret)) {
				t.Fatal("private binding in provider payload")
			}
		}
		if !bytes.Contains(req.Body, []byte("Visual")) || !bytes.Contains(req.Body, []byte("schema_version")) {
			t.Fatal("missing controlled contract")
		}
		raw, _ := json.Marshal(map[string]any{"choices": []any{map[string]any{"finish_reason": "stop", "message": map[string]any{"role": "assistant", "content": string(fixture)}}}})
		return analysis.DispatchReply{Body: raw}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return runner, request, &calls, &reservations
}

func TestVisualAttemptReturnsValidatedUnknownWithoutPrivateInput(t *testing.T) {
	runner, req, calls, reservations := visualFixture(t)
	result, err := runner.RunVisual(context.Background(), req)
	if err != nil || result == nil {
		t.Fatal("attempt failed", err)
	}
	doc, err := result.Proposal().Data()
	if err != nil || doc.Outcome != "unknown" {
		t.Fatal("unknown not preserved", err)
	}
	if *calls != 1 || *reservations != 1 {
		t.Fatal("dispatch count")
	}
	info := result.Info()
	if info.Image.Binding.Asset != req.Asset || info.Format != "strict" || !strings.Contains(info.PromptVersion, "visual") {
		t.Fatal("missing handoff")
	}
}
