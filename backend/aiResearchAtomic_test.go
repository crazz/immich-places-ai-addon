package main

import (
	"context"
	"immich-places-backend/internal/ai/jobs"
	"testing"
)

func TestAIResearchPublicationRollsBackAnswerReferencesAndHistoryTogether(t *testing.T) {
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
	completion, err := p.executor(f.analyzer)(ctx, lease, jobs.Guard{Authorize: func(c context.Context) error { return store.Authorize(c, lease) }, Reserve: func(c context.Context) error { return store.Reserve(c, lease) }})
	if err != nil {
		t.Fatal(err)
	}
	selectionSQL(t, f.image.db, `CREATE TRIGGER reject_research_history BEFORE INSERT ON ai_result_history BEGIN SELECT RAISE(ABORT,'synthetic publication failure'); END`)
	if _, err = store.Complete(ctx, lease, completion); err == nil {
		t.Fatal("publication ignored storage failure")
	}
	for _, table := range []string{"ai_analyses", "ai_result_history"} {
		var count int
		if err = f.image.db.db.QueryRow("SELECT count(*) FROM " + table).Scan(&count); err != nil || count != 0 {
			t.Fatal("partial publication", table, err)
		}
	}
	retained, err := p.store.Get(ctx, testUserID, job.ID)
	if err != nil || retained.Items[0].State != "running" || retained.Calls != 1 {
		t.Fatal("partial terminal state", err)
	}
	if f.hits.Load() != 1 {
		t.Fatal("completion repeated provider request")
	}
}
