package main

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"immich-places-backend/internal/ai/contextual"
)

func contextRequest(t *testing.T, f *aiImageFixture) aiContextRequest {
	t.Helper()
	data, _ := json.Marshal(f.metadata())
	digest, err := aiImageSourceDigest(data, selectionA)
	if err != nil {
		t.Fatal(err)
	}
	b := contextual.Binding{Owner: testUserID, Installation: f.store.binding, Asset: selectionA, Selection: "frozen-selection", Profile: "context-profile", Revision: 1, SourceDigest: digest}
	return aiContextRequest{Binding: b, Consent: contextual.Consent{Binding: b, Version: contextual.Version, Classes: []contextual.Class{contextual.Hint}}, Hint: "Beside a bridge", Window: 6 * time.Hour, Authorize: func(context.Context) error { return nil }}
}

func TestAIContextReadsCurrentTargetWithoutUnconsentedSources(t *testing.T) {
	f := newAIImageFixture(t)
	s := &aiContextPreparer{images: f.service}
	req := contextRequest(t, f)
	var before int
	if err := f.db.db.QueryRow("SELECT total_changes()").Scan(&before); err != nil {
		t.Fatal(err)
	}
	b, err := s.prepare(context.Background(), req)
	if err != nil || b == nil {
		t.Fatal("authorized context missing", err)
	}
	p, _ := b.Projection()
	if !strings.Contains(string(p), "Beside a bridge") || strings.Contains(string(p), selectionA) {
		t.Fatal("wrong context projection")
	}
	count, bad := f.counts()
	if count != 1 || bad != 0 {
		t.Fatal("unselected context or image fetched", count, bad)
	}
	var after int
	if err := f.db.db.QueryRow("SELECT total_changes()").Scan(&after); err != nil || after != before {
		t.Fatal("context wrote database", err)
	}
}
