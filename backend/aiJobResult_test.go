package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"immich-places-backend/internal/ai/jobs"
	"immich-places-backend/internal/ai/results"
)

func aiJobCompletion(t *testing.T) jobs.Completion {
	t.Helper()
	raw, err := os.ReadFile("internal/ai/results/testdata/ai-analysis-result.unknown.json")
	if err != nil {
		t.Fatal(err)
	}
	validator, err := results.New()
	if err != nil {
		t.Fatal(err)
	}
	proposal, err := validator.Validate(raw, results.Context{Mode: results.Visual, Completion: results.Complete, Languages: []string{"en"}, PrimaryLanguage: "en"})
	if err != nil {
		t.Fatal(err)
	}
	return jobs.Completion{Proposal: proposal, SourceDigest: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", ImageDigest: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", PromptVersion: "visual-v1", SchemaVersion: "1.0"}
}

func (f *aiJobFixture) reopen(t *testing.T) {
	t.Helper()
	var seq int
	var name, path string
	if err := f.db.db.QueryRow("PRAGMA database_list").Scan(&seq, &name, &path); err != nil {
		t.Fatal(err)
	}
	f.db.close()
	db, err := newDatabase(filepath.Dir(path), "test-encryption-key")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.close() })
	f.db = db
	f.selection.db = db
	f.store = newAIJobStore(db, f.input.Installation, true, func() time.Time { return f.now })
}

func TestAIJobUnknownCompletionIsAtomicAndDurable(t *testing.T) {
	f := newAIJobFixture(t)
	ctx := context.Background()
	job, err := f.store.Submit(ctx, f.input)
	if err != nil {
		t.Fatal(err)
	}
	lease, ok, err := f.store.Claim(ctx, jobs.DefaultPolicy())
	if err != nil || !ok {
		t.Fatal(err)
	}
	if err = f.store.Reserve(ctx, lease); err != nil {
		t.Fatal(err)
	}
	completion := aiJobCompletion(t)
	id, err := f.store.Complete(ctx, lease, completion)
	if err != nil || id == "" {
		t.Fatal("complete", err)
	}
	f.reopen(t)
	stored, err := f.store.ReadAnalysis(ctx, testUserID, id)
	if err != nil {
		t.Fatal(err)
	}
	expected, _ := completion.Proposal.MarshalJSON()
	if stored.Outcome != "unknown" || stored.ItemID != lease.ItemID || stored.JobID != job.ID || stored.Asset != lease.Asset || !bytes.Equal(stored.Payload, expected) || stored.Metadata.Revision != 1 || stored.Metadata.Model != "manual-model" || stored.Metadata.Installation != f.input.Installation || stored.Metadata.SourceDigest != completion.SourceDigest || stored.Metadata.PromptVersion != "visual-v1" || stored.Metadata.ValidationVersion != "analysis-result-v1" {
		t.Fatal("result provenance changed", stored.Metadata)
	}
	loaded, err := f.store.Get(ctx, testUserID, job.ID)
	if err != nil || loaded.Items[0].State != "succeeded" || loaded.Items[1].State != "queued" || loaded.Calls != 1 {
		t.Fatal("terminal state not durable", loaded.Items, err)
	}
}

func TestAIJobTerminalAnalysisCannotBeOverwritten(t *testing.T) {
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
	completion := aiJobCompletion(t)
	id, err := f.store.Complete(ctx, lease, completion)
	if err != nil {
		t.Fatal(err)
	}
	if repeated, err := f.store.Complete(ctx, lease, completion); err == nil || repeated != "" {
		t.Fatal("terminal lease reused")
	}
	if _, err = f.db.db.Exec("UPDATE ai_analyses SET payload='{}' WHERE id=?", id); err == nil {
		t.Fatal("immutable analysis updated")
	}
	stored, err := f.store.ReadAnalysis(ctx, testUserID, id)
	expected, _ := completion.Proposal.MarshalJSON()
	if err != nil || !bytes.Equal(stored.Payload, expected) {
		t.Fatal("analysis changed", err)
	}
}
