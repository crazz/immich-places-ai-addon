package analysis_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"immich-places-backend/internal/ai/analysis"
	"immich-places-backend/internal/aiadapters/providerhttp"
)

func TestVisualResultOwnsProposalAndMetadata(t *testing.T) {
	_, req, _, _ := visualFixture(t)
	tokens := 7
	content, err := os.ReadFile("../results/testdata/ai-analysis-result.unknown.json")
	if err != nil {
		t.Fatal(err)
	}
	runner, _ := analysis.New(analysis.Protocol{Encode: providerhttp.EncodeVisual, Parse: func([]byte) (analysis.Response, error) {
		return analysis.Response{Content: append([]byte(nil), content...), Usage: &analysis.Usage{TotalTokens: &tokens}}, nil
	}}, func(ctx context.Context, r analysis.DispatchRequest) (analysis.DispatchReply, error) {
		return analysis.DispatchReply{Body: []byte("{}")}, r.Authorize(ctx)
	})
	result, err := runner.RunVisual(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	tokens = 99
	req.Languages[0] = "uk"
	clear(content)
	info := result.Info()
	if *info.Usage.TotalTokens != 7 || info.Languages[0] != "en" {
		t.Fatal("retained metadata aliases caller input")
	}
	*info.Usage.TotalTokens = 888
	info.Languages[0] = "de"
	doc, _ := result.Proposal().Data()
	doc.Warnings[0] = "changed"
	stable, _ := result.Proposal().Data()
	if stable.Warnings[0] == "changed" || *result.Info().Usage.TotalTokens != 7 || result.Info().Languages[0] != "en" {
		t.Fatal("accessor mutation changed result")
	}
	raw, err := json.Marshal(result)
	if err != nil || string(raw) != "{}" {
		t.Fatal("default serialization disclosed private result")
	}
}

func TestVisualDefaultFormattingDoesNotExposeProposalOrBinding(t *testing.T) {
	runner, req, _, _ := visualFixture(t)
	result, err := runner.RunVisual(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range []any{result, *result} {
		for _, format := range []string{"%v", "%+v", "%#v"} {
			diagnostic := fmt.Sprintf(format, value)
			if len(diagnostic) > 128 || strings.Contains(diagnostic, req.Asset) || strings.Contains(diagnostic, "data:") {
				t.Fatal("default formatting disclosed retained result")
			}
		}
	}
}
