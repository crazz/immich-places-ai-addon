package main

import (
	"context"
	"database/sql"
	"net/http"
	"reflect"
	"testing"

	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writeback"
)

func TestAIMirrorFailureRetainsCatalogRefreshExactDraftAndPartialHistoryAfterReopen(t *testing.T) {
	f := mirrorWriteFixture(t)
	mirrorMutationServer(t, f)
	ctx := context.Background()
	f.handle = func(w http.ResponseWriter, r *http.Request) bool {
		if r.Method == "PUT" {
			f.mu.Lock()
			f.mirrorSends++
			f.mu.Unlock()
			w.WriteHeader(400)
			return true
		}
		return false
	}
	op, err := f.writer.confirm(ctx, testUserID, writeback.Confirmation{PreviewID: f.preview.Plan.ID, Digest: f.preview.Digest, Key: "retained-partial"})
	if err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if err = f.writer.runOne(ctx, testUserID, op.ID); err != nil {
			t.Fatal(err)
		}
	}
	before, err := f.writer.get(ctx, testUserID, op.ID, false)
	if err != nil {
		t.Fatal(err)
	}
	f.f.reopen(t)
	f.writer.drafts.results.jobs = f.f.store
	saved, err := f.writer.get(ctx, testUserID, op.ID, false)
	if err != nil || !reflect.DeepEqual(before, saved) || saved.Status != "partial" || saved.Mirror.Status != "retryable" || !saved.Targets[0].Verified || !saved.Targets[0].Refreshed {
		t.Fatal("partial outcomes or local refresh lost", err)
	}
	var retained drafts.Draft
	if err = f.writer.drafts.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var err error
		retained, err = f.writer.drafts.read(ctx, tx, testUserID, f.draft.ID)
		return err
	}); err != nil || !reflect.DeepEqual(f.draft, retained) {
		t.Fatal("failed optional mirror changed local geometry or language records", err)
	}
	var lat, lon float64
	if err = f.f.db.db.QueryRow(`SELECT latitude,longitude FROM assets WHERE userID=? AND immichID=?`, testUserID, f.draft.AssetID).Scan(&lat, &lon); err != nil || lat != 0 || lon != 12 {
		t.Fatal("verified catalog GPS lost", lat, lon, err)
	}
	page, err := f.writer.history(ctx, testUserID, f.draft.ID, "")
	if err != nil || len(page.Items) != 1 || page.Items[0].Status != "partial" {
		t.Fatal("history hid incomplete mirror", err)
	}
	if f.standardSends != 1 || f.mirrorSends != 1 {
		t.Fatal("reopen or failure recovery repeated standard fields")
	}
}
