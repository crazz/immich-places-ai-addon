package main

import (
	"context"
	"sync"
	"testing"
	"time"

	"immich-places-backend/internal/ai/jobs"
)

func TestAIJobCompetingClaimsRespectOwnerAndGlobalCapacity(t *testing.T) {
	f := newAIJobFixture(t)
	ctx := context.Background()
	for _, owner := range []string{testUserID, "owner-b", "owner-c"} {
		if owner != testUserID {
			if err := f.db.createUser(ctx, owner, owner+"@example.com", "hash"); err != nil {
				t.Fatal(err)
			}
			if _, err := f.db.createAIProvider(ctx, owner, "profile", providerInput()); err != nil {
				t.Fatal(err)
			}
		}
		input := f.input
		input.Owner = owner
		if _, err := f.store.Submit(ctx, input); err != nil {
			t.Fatal(err)
		}
	}
	claimed := make(chan jobs.Lease, 12)
	var wg sync.WaitGroup
	for range 12 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			store := newAIJobStore(f.db, f.selection.binding, true, f.store.now)
			lease, ok, err := store.Claim(ctx, jobs.DefaultPolicy())
			if err != nil {
				t.Error(err)
			}
			if ok {
				claimed <- lease
			}
		}()
	}
	wg.Wait()
	close(claimed)
	owners, tokens := map[string]bool{}, map[string]bool{}
	for lease := range claimed {
		if owners[lease.Owner] || tokens[lease.Token] || len(lease.Token) < 16 || lease.Attempts != 1 || lease.Calls != 0 || !lease.ExpiresAt.After(f.now) {
			t.Fatal("invalid exclusive claim", lease)
		}
		owners[lease.Owner] = true
		tokens[lease.Token] = true
	}
	if len(tokens) != 2 || len(owners) != 2 {
		t.Fatal("capacity not enforced", len(tokens), len(owners))
	}
	var running int
	if err := f.db.db.QueryRow("SELECT count(*) FROM ai_job_items WHERE state='running'").Scan(&running); err != nil || running != 2 {
		t.Fatal("wrong durable claims", running, err)
	}
}

func TestAIJobFutureRetryWaitsUntilDue(t *testing.T) {
	f := newAIJobFixture(t)
	ctx := context.Background()
	job, err := f.store.Submit(ctx, f.input)
	if err != nil {
		t.Fatal(err)
	}
	due := f.now.Add(time.Minute)
	selectionSQL(t, f.db, "UPDATE ai_job_items SET state='retry_wait',nextAttemptAt=? WHERE jobID=?", due.UnixNano(), job.ID)
	if _, ok, err := f.store.Claim(ctx, jobs.DefaultPolicy()); err != nil || ok {
		t.Fatal("early retry", ok, err)
	}
	f.now = due
	lease, ok, err := f.store.Claim(ctx, jobs.DefaultPolicy())
	if err != nil || !ok || lease.Attempts != 1 {
		t.Fatal("due retry not claimed", ok, err)
	}
}
