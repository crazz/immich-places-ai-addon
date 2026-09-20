package main

import (
	"context"
	"github.com/pressly/goose/v3"
	"immich-places-backend/internal/ai/jobs"
	"path/filepath"
	"testing"
	"time"
)

func TestAIResearchAndLegacyAnswersSurviveHistoryMigrationAndReopen(t *testing.T) {
	f, p, req := researchProductionFixture(t)
	ctx := context.Background()
	research, err := p.submit(ctx, testUserID, req)
	if err != nil {
		t.Fatal(err)
	}
	store := &aiProductionWorkerStore{aiJobStore: p.store, production: p}
	worker := jobs.Worker{Store: store, Policy: jobs.DefaultPolicy(), Execute: p.executor(f.analyzer)}
	if _, err = worker.RunOne(ctx, make(chan time.Time)); err != nil {
		t.Fatal(err)
	}
	progress, err := p.progress(ctx, testUserID, research.ID)
	if err != nil || progress.Items[0].ResultID == nil {
		t.Fatal("Research result missing", err)
	}
	researchID := *progress.Items[0].ResultID
	req.Configuration.Mode = "visual"
	req.Configuration.Context = nil
	req.Consent.Configuration = req.Configuration
	req.IdempotencyKey = "legacy-second"
	if _, err = p.submit(ctx, testUserID, req); err != nil {
		t.Fatal(err)
	}
	lease, claimed, err := store.Claim(ctx, jobs.DefaultPolicy())
	if err != nil || !claimed {
		t.Fatal("legacy claim missing", err)
	}
	if err = store.Reserve(ctx, lease); err != nil {
		t.Fatal(err)
	}
	legacyID, err := store.Complete(ctx, lease, aiJobCompletion(t))
	if err != nil {
		t.Fatal(err)
	}
	before := map[string]string{}
	for _, id := range []string{researchID, legacyID} {
		var raw string
		if err = f.image.db.db.QueryRow("SELECT payload FROM ai_analyses WHERE id=?", id).Scan(&raw); err != nil {
			t.Fatal(err)
		}
		before[id] = raw
	}
	if err = goose.DownTo(f.image.db.db, "migrations", 24); err != nil {
		t.Fatal(err)
	}
	var seq int
	var name, path string
	if err = f.image.db.db.QueryRow("PRAGMA database_list").Scan(&seq, &name, &path); err != nil {
		t.Fatal(err)
	}
	f.image.db.close()
	db, err := newDatabase(filepath.Dir(path), "test-encryption-key")
	if err != nil {
		t.Fatal(err)
	}
	defer db.close()
	history := &aiResultStore{jobs: newAIJobStore(db, p.store.binding, false, time.Now)}
	for id, raw := range before {
		var retained string
		if err = db.db.QueryRow("SELECT payload FROM ai_analyses WHERE id=?", id).Scan(&retained); err != nil || retained != raw {
			t.Fatal("migration rewrote immutable answer", err)
		}
		detail, err := history.detail(ctx, testUserID, id, "", "")
		if err != nil || detail.Proposal == nil {
			t.Fatal("migrated history unavailable", err)
		}
		want := "1.0"
		if id == researchID {
			want = "2.0"
		}
		if detail.Proposal.SchemaVersion != want {
			t.Fatal("history version changed")
		}
	}
}
