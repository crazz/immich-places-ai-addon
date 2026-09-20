package main

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/pressly/goose/v3"
	"immich-places-backend/internal/ai/jobs"
)

func TestAIProductionUpgradeRetainsInternalHistoryWithoutConsentPromotion(t *testing.T) {
	f := newAIJobFixture(t)
	ctx := context.Background()
	old, err := f.store.Submit(ctx, f.input)
	if err != nil {
		t.Fatal(err)
	}
	lease, _, err := f.store.Claim(ctx, jobs.DefaultPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if err = f.store.Reserve(ctx, lease); err != nil {
		t.Fatal(err)
	}
	resultID, err := f.store.Complete(ctx, lease, aiJobCompletion(t))
	if err != nil {
		t.Fatal(err)
	}
	if err = goose.DownTo(f.db.db, "migrations", 22); err != nil {
		t.Fatal(err)
	}
	var seq int
	var name, path string
	if err = f.db.db.QueryRow("PRAGMA database_list").Scan(&seq, &name, &path); err != nil {
		t.Fatal(err)
	}
	f.db.close()
	db, err := newDatabase(filepath.Dir(path), "test-encryption-key")
	if err != nil {
		t.Fatal(err)
	}
	defer db.close()
	store := newAIJobStore(db, f.store.binding, true, f.store.now)
	if record, err := store.ReadAnalysis(ctx, testUserID, resultID); err != nil || record.JobID != old.ID {
		t.Fatal("old history lost", err)
	}
	s := &aiProductionWorkerStore{aiJobStore: store, production: &aiProductionJobs{store: store}}
	if _, claimed, err := s.Claim(ctx, jobs.DefaultPolicy()); err != nil || claimed {
		t.Fatal("old queued item promoted", claimed, err)
	}
	conns := []*sql.Conn{}
	defer func() {
		for _, c := range conns {
			c.Close()
		}
	}()
	for i := 0; i < 4; i++ {
		c, err := db.db.Conn(ctx)
		if err != nil {
			t.Fatal(err)
		}
		conns = append(conns, c)
		if _, err = c.ExecContext(ctx, "INSERT INTO ai_job_admissions VALUES('foreign','absent','production-v1','digest','{}','{}','policy','{}')"); err == nil {
			t.Fatal("new tables lost pooled foreign keys")
		}
	}
}
