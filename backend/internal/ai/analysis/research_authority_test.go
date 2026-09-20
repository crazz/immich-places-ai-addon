package analysis_test

import (
	"context"
	"errors"
	"immich-places-backend/internal/ai/analysis"
	"immich-places-backend/internal/ai/contextual"
	"immich-places-backend/internal/aiadapters/providerhttp"
	"testing"
	"time"
)

func TestResearchRejectsNeighborAuthorizationBeforeReservation(t *testing.T) {
	_, req, _, reservations := visualFixture(t)
	calls := 0
	runner, err := analysis.New(analysis.Protocol{Encode: providerhttp.EncodeResearch, Parse: providerhttp.ParseVisual}, func(context.Context, analysis.DispatchRequest) (analysis.DispatchReply, error) {
		calls++
		return analysis.DispatchReply{}, analysis.ErrInvalidRequest
	})
	if err != nil {
		t.Fatal(err)
	}
	info := contextBundle(t, req).Info()
	info.Consent.Classes = []contextual.Class{contextual.Neighbors}
	bundle, err := contextual.Build(contextual.Input{Binding: info.Binding, Consent: info.Consent, Window: time.Hour})
	if err != nil {
		t.Fatal(err)
	}
	req.Context = bundle
	result, err := runner.RunResearch(context.Background(), req)
	if !errors.Is(err, analysis.ErrInvalidRequest) || result != nil || calls != 0 || *reservations != 0 {
		t.Fatal("Research accepted hidden neighbors", err, calls, *reservations)
	}
}
