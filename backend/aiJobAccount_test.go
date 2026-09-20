package main

import (
	"context"
	"testing"

	"immich-places-backend/internal/ai/jobs"
)

func TestAIJobAccountDeletionCascadesOnlyOwnedPrivateHistory(t *testing.T) {
	f := newAIJobFixture(t)
	ctx := context.Background()
	f.input.AssetIDs = []string{selectionA}
	f.input.MaxCalls = 3
	if err := f.db.createUser(ctx, "other-owner", "other-job@example.com", "hash"); err != nil {
		t.Fatal(err)
	}
	if _, err := f.db.createAIProvider(ctx, "other-owner", "profile", providerInput()); err != nil {
		t.Fatal(err)
	}
	for _, owner := range []string{testUserID, "other-owner"} {
		input := f.input
		input.Owner = owner
		if _, err := f.store.Submit(ctx, input); err != nil {
			t.Fatal(err)
		}
		lease, ok, err := f.store.Claim(ctx, jobs.DefaultPolicy())
		if err != nil || !ok {
			t.Fatal(err)
		}
		if err = f.store.Reserve(ctx, lease); err != nil {
			t.Fatal(err)
		}
		if _, err = f.store.Complete(ctx, lease, aiJobCompletion(t)); err != nil {
			t.Fatal(err)
		}
	}
	selectionSQL(t, f.db, "DELETE FROM users WHERE ID=?", testUserID)
	for _, table := range []string{"ai_jobs", "ai_job_items", "ai_analyses"} {
		var owned, other int
		if err := f.db.db.QueryRow("SELECT count(*) FROM "+table+" WHERE userID=?", testUserID).Scan(&owned); err != nil {
			t.Fatal(err)
		}
		if err := f.db.db.QueryRow("SELECT count(*) FROM " + table + " WHERE userID='other-owner'").Scan(&other); err != nil {
			t.Fatal(err)
		}
		if owned != 0 || other != 1 {
			t.Fatal("incorrect private cascade", table, owned, other)
		}
	}
}
