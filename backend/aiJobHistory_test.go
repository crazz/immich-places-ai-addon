package main

import (
	"bytes"
	"context"
	"testing"

	"immich-places-backend/internal/ai/jobs"
)

func TestAIJobHistoryRemainsPrivateAfterSourceRemoval(t *testing.T) {
	f := newAIJobFixture(t)
	ctx := context.Background()
	seedAsset(t, f.db, selectionA, nil, nil, "2026-09-20")
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
	if err != nil {
		t.Fatal(err)
	}
	selectionSQL(t, f.db, "DELETE FROM assets WHERE userID=? AND immichID=?", testUserID, selectionA)
	if _, err = f.store.Get(ctx, "foreign", job.ID); err != jobs.ErrDenied {
		t.Fatal("foreign job read", err)
	}
	if _, err = f.store.ReadAnalysis(ctx, "foreign", id); err != jobs.ErrDenied {
		t.Fatal("foreign result read", err)
	}
	record, err := f.store.ReadAnalysis(ctx, testUserID, id)
	want, _ := completion.Proposal.MarshalJSON()
	if err != nil || !bytes.Equal(record.Payload, want) {
		t.Fatal("source deletion removed history", err)
	}
	record.Payload[0] = '!'
	record.Metadata.Languages[0] = "changed"
	again, err := f.store.ReadAnalysis(ctx, testUserID, id)
	if err != nil || !bytes.Equal(again.Payload, want) || again.Metadata.Languages[0] != "en" {
		t.Fatal("caller changed retained history", err)
	}
}

func TestAIJobExplicitNewRunPreservesEarlierResult(t *testing.T) {
	f := newAIJobFixture(t)
	ctx := context.Background()
	f.input.AssetIDs = []string{selectionA}
	f.input.MaxCalls = 3
	var ids []string
	for run := range 2 {
		if run == 1 {
			input := providerInput()
			input.Model = "new-model"
			if _, err := f.db.updateAIProvider(ctx, testUserID, "profile", 1, input); err != nil {
				t.Fatal(err)
			}
			f.input.Key = "new-run"
			f.input.Revision = 2
		}
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
		ids = append(ids, id)
	}
	first, err := f.store.ReadAnalysis(ctx, testUserID, ids[0])
	if err != nil {
		t.Fatal(err)
	}
	second, err := f.store.ReadAnalysis(ctx, testUserID, ids[1])
	if err != nil {
		t.Fatal(err)
	}
	if ids[0] == ids[1] || first.JobID == second.JobID || first.Metadata.Revision != 1 || first.Metadata.Model != "manual-model" || second.Metadata.Revision != 2 || second.Metadata.Model != "new-model" {
		t.Fatal("reanalysis replaced earlier history")
	}
}
