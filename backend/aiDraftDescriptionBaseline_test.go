package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"immich-places-backend/internal/ai/drafts"
)

func TestAIDraftDescriptionBaselinePersistsExactLargeText(t *testing.T) {
	f, analysis := draftFixture(t)
	value := acceptedDraft(t, f, analysis)
	image := newAIImageFixture(t)
	seedAsset(t, f.db, selectionA, nil, nil, "2026-09-20")
	key := "private-immich-image-key"
	if err := f.db.updateImmichAPIKey(context.Background(), testUserID, &key); err != nil {
		t.Fatal(err)
	}
	before := "  e\u0301\r\n" + strings.Repeat("\n", 20<<10) + "尾  "
	image.handle = func(w http.ResponseWriter, r *http.Request) bool {
		if r.URL.Path != "/api/assets/"+selectionA {
			return false
		}
		meta := image.metadata()
		meta["exifInfo"] = map[string]any{"latitude": nil, "longitude": nil, "description": before}
		if err := json.NewEncoder(w).Encode(meta); err != nil {
			t.Error(err)
		}
		return true
	}
	store := &aiDraftStore{results: &aiResultStore{jobs: f.store, origin: aiTestOrigin}}
	ctx := context.Background()
	observation, err := store.observe(ctx, testUserID, value.ID, value.Revision, image.service)
	if err != nil {
		t.Fatal(err)
	}
	assertText := func(baseline drafts.Baseline) {
		t.Helper()
		raw, err := json.Marshal(baseline)
		var result struct {
			Description *struct{ Presence, Value string } `json:"description"`
		}
		if err != nil || json.Unmarshal(raw, &result) != nil || result.Description == nil || result.Description.Presence != "value" || result.Description.Value != before {
			t.Fatal("exact description baseline missing or changed")
		}
	}
	assertText(observation.Baseline)
	ack, err := store.acknowledge(ctx, testUserID, value.ID, value.Revision, observation.ID, image.service)
	if err != nil {
		t.Fatal(err)
	}
	assertText(ack.Baseline)
	reloaded := &aiDraftStore{results: store.results}
	saved, err := reloaded.get(ctx, testUserID, value.ID)
	if err != nil || saved.Revision != 2 {
		t.Fatalf("reloaded revision: %d %v", saved.Revision, err)
	}
	assertText(saved.Baseline)
	if _, err := reloaded.get(ctx, "another-owner", value.ID); err == nil {
		t.Fatal("foreign owner read private baseline")
	}
	if _, bad := image.counts(); bad != 0 {
		t.Fatal("baseline review mutated Immich")
	}
}
