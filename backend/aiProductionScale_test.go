package main

import (
	"context"
	"path/filepath"
	"testing"

	"immich-places-backend/internal/ai/selection"
)

func TestAIProductionFiveHundredItemsStayBoundedAndPersistAcrossReopen(t *testing.T) {
	f, p, req := productionFixture(t)
	ctx := context.Background()
	ids := []string{selectionA}
	for i := 1; i < 500; i++ {
		id := selectionID(i + 100)
		seedAsset(t, f.image.db, id, nil, nil, "2026-09-20")
		ids = append(ids, id)
	}
	snapshot, err := p.selections.preview(ctx, selection.Input{Mode: "explicit", AssetIDs: ids, Scope: &selection.Scope{View: "all"}}, testUserID)
	if err != nil {
		t.Fatal(err)
	}
	req.Configuration.SelectionToken = *snapshot.SnapshotID
	req.Configuration.Limits.MaxCalls = 0
	req.Consent.Configuration = req.Configuration
	job, err := p.submit(ctx, testUserID, req)
	if err != nil || len(job.Items) != 500 || job.Input.MaxCalls != 500 {
		t.Fatal("500-item admission", err)
	}
	var sequence int
	var name, path string
	if err = f.image.db.db.QueryRow("PRAGMA database_list").Scan(&sequence, &name, &path); err != nil {
		t.Fatal(err)
	}
	f.image.db.close()
	database, err := newDatabase(filepath.Dir(path), "test-encryption-key")
	if err != nil {
		t.Fatal(err)
	}
	defer database.close()
	reopened := newAIJobStore(database, p.store.binding, true, p.store.now)
	p.store = reopened
	read, err := reopened.Get(ctx, testUserID, job.ID)
	if err != nil || len(read.Items) != 500 {
		t.Fatal("membership not durable", err)
	}
	progress, err := p.progress(ctx, testUserID, job.ID)
	if err != nil || len(progress.Items) != 500 || progress.Counts["total"] != 500 {
		t.Fatal("bounded detail", err)
	}
	page, err := p.list(ctx, testUserID, 20, "")
	if err != nil || len(page.Items) != 1 || len(page.Items[0].Items) != 0 || page.Items[0].Counts["total"] != 500 {
		t.Fatal("unbounded list detail", err)
	}
	if f.hits.Load() != 0 {
		t.Fatal("admission dispatched inference")
	}
}
