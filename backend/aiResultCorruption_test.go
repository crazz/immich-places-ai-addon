package main

import (
	"context"
	"immich-places-backend/internal/ai/jobs"
	"immich-places-backend/internal/ai/review"
	"testing"
)

func TestAIResultDetailRejectsCorruptContractsWithoutBreakingHistory(t *testing.T) {
	for _, mutation := range []string{
		`UPDATE ai_analyses SET payload='{}'`,
		`UPDATE ai_analyses SET metadata=json_set(metadata,'$.ValidationVersion','unsupported')`,
		`UPDATE ai_analyses SET metadata=json_set(metadata,'$.SchemaVersion','future')`,
		`UPDATE ai_analyses SET metadata=json_set(metadata,'$.Installation','foreign')`,
		`UPDATE ai_analyses SET metadata=json_set(metadata,'$.Mode','context-assisted')`,
		`UPDATE ai_analyses SET metadata=json_set(metadata,'$.Languages',json('["uk"]'))`,
	} {
		t.Run(mutation, func(t *testing.T) {
			f := newAIJobFixture(t)
			ctx := context.Background()
			if _, err := f.store.Submit(ctx, f.input); err != nil {
				t.Fatal(err)
			}
			lease, ok, err := f.store.Claim(ctx, jobs.DefaultPolicy())
			if err != nil || !ok {
				t.Fatal(err)
			}
			if err = f.store.Reserve(ctx, lease); err != nil {
				t.Fatal(err)
			}
			id, err := f.store.Complete(ctx, lease, aiJobCompletion(t))
			if err != nil {
				t.Fatal(err)
			}
			selectionSQL(t, f.db, "DROP TRIGGER ai_analyses_immutable")
			selectionSQL(t, f.db, mutation)
			s := &aiResultStore{jobs: f.store}
			if _, err = s.detail(ctx, testUserID, id, "", ""); err == nil {
				t.Fatal("corrupt record accepted")
			}
			page, err := s.list(ctx, testUserID, review.Query{Limit: 30})
			if err != nil || len(page.Items) != 1 {
				t.Fatal("corrupt detail broke history", err)
			}
		})
	}
}
