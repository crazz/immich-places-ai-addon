package main

import (
	"context"
	"immich-places-backend/internal/ai/jobs"
	"path/filepath"
	"testing"
	"time"
)

func TestAIResearchRestartNeverResendsAfterDefaultReservation(t *testing.T) {
	f, p, req := researchProductionFixture(t)
	ctx := context.Background()
	job, err := p.submit(ctx, testUserID, req)
	if err != nil {
		t.Fatal(err)
	}
	store := &aiProductionWorkerStore{aiJobStore: p.store, production: p}
	lease, claimed, err := store.Claim(ctx, jobs.DefaultPolicy())
	if err != nil || !claimed {
		t.Fatal("claim failed", err)
	}
	if err = store.Reserve(ctx, lease); err != nil {
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
	now := time.Now().Add(11 * time.Minute)
	reopened := newAIJobStore(db, p.store.binding, true, func() time.Time { return now })
	if recovered, err := reopened.Recover(ctx); err != nil || recovered != 1 {
		t.Fatal("recovery failed", recovered, err)
	}
	if _, claimed, err = reopened.Claim(ctx, jobs.DefaultPolicy()); err != nil || claimed {
		t.Fatal("uncertain image was automatically reserved again", err)
	}
	got, err := reopened.Get(ctx, testUserID, job.ID)
	if err != nil || got.Calls != 1 || got.Items[0].State != "failed" {
		t.Fatal("reservation or terminal outcome lost", got, err)
	}
	if f.hits.Load() != 0 {
		t.Fatal("recovery contacted provider")
	}
}
