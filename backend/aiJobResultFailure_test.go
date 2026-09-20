package main

import (
	"context"
	"testing"

	"immich-places-backend/internal/ai/jobs"
	"immich-places-backend/internal/ai/results"
)

func TestAIJobInvalidOrFailedResultNeverPartiallySucceeds(t *testing.T) {
	for _, failure := range []string{"uninitialized", "wrong-languages", "source", "image", "prompt", "schema", "unreserved", "insert"} {
		t.Run(failure, func(t *testing.T) {
			f := newAIJobFixture(t)
			ctx := context.Background()
			if failure == "wrong-languages" {
				f.input.Languages = []string{"uk"}
				f.input.PrimaryLanguage = "uk"
			}
			job, err := f.store.Submit(ctx, f.input)
			if err != nil {
				t.Fatal(err)
			}
			lease, ok, err := f.store.Claim(ctx, jobs.DefaultPolicy())
			if err != nil || !ok {
				t.Fatal(err)
			}
			if failure != "unreserved" {
				if err = f.store.Reserve(ctx, lease); err != nil {
					t.Fatal(err)
				}
			}
			completion := aiJobCompletion(t)
			switch failure {
			case "uninitialized":
				completion.Proposal = results.Proposal{}
			case "source":
				completion.SourceDigest = "bad"
			case "image":
				completion.ImageDigest = "bad"
			case "prompt":
				completion.PromptVersion = "other"
			case "schema":
				completion.SchemaVersion = "other"
			case "insert":
				selectionSQL(t, f.db, `CREATE TRIGGER reject_analysis BEFORE INSERT ON ai_analyses BEGIN SELECT RAISE(ABORT,'private result detail'); END`)
			}
			if id, err := f.store.Complete(ctx, lease, completion); err == nil || id != "" {
				t.Fatal("invalid completion accepted")
			}
			var count int
			if err = f.db.db.QueryRow("SELECT count(*) FROM ai_analyses").Scan(&count); err != nil || count != 0 {
				t.Fatal("partial result", count, err)
			}
			kept, err := f.store.Get(ctx, testUserID, job.ID)
			if err != nil || kept.Items[0].State != "running" {
				t.Fatal("partial terminal transition", err)
			}
		})
	}
}
