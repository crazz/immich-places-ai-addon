package analysis_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"immich-places-backend/internal/ai/analysis"
	"immich-places-backend/internal/aiadapters/providerhttp"
)

func TestVisualClassifiesUnsupportedFormatWithoutRetryOrPrivateError(t *testing.T) {

	_, req, _, _ := visualFixture(t)
	runner, _ := analysis.New(analysis.Protocol{Encode: providerhttp.EncodeVisual, Parse: providerhttp.ParseVisual}, func(context.Context, analysis.DispatchRequest) (analysis.DispatchReply, error) {
		return analysis.DispatchReply{}, errors.New("private database / credential marker")
	})
	_, err := runner.RunVisual(context.Background(), req)
	if !errors.Is(err, analysis.ErrUpstream) || strings.Contains(err.Error(), "private") {
		t.Fatal("raw error leaked", err)
	}
}

func TestVisualProtocolFailuresNeverExposePrivatePayloads(t *testing.T) {
	for _, atEncode := range []bool{true, false} {
		_, req, _, _ := visualFixture(t)
		protocol := analysis.Protocol{Encode: providerhttp.EncodeVisual, Parse: providerhttp.ParseVisual}
		if atEncode {
			protocol.Encode = func(string, string, string, string) ([]byte, error) { return nil, errors.New("private prompt") }
		} else {
			protocol.Parse = func([]byte) (analysis.Response, error) { return analysis.Response{}, errors.New("private response") }
		}
		runner, _ := analysis.New(protocol, func(ctx context.Context, r analysis.DispatchRequest) (analysis.DispatchReply, error) {
			return analysis.DispatchReply{Body: []byte("{}")}, r.Authorize(ctx)
		})
		result, err := runner.RunVisual(context.Background(), req)
		if err == nil || result != nil || strings.Contains(err.Error(), "private") {
			t.Fatal("protocol error disclosed payload", err)
		}
	}
}
