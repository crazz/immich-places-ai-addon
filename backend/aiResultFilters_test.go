package main

import (
	"context"
	"immich-places-backend/internal/ai/review"
	"testing"
)

func TestAIResultFiltersUseOnlyRetainedSourceLocalDateAndSelectedAlbum(t *testing.T) {
	f := newAIJobFixture(t)
	ctx := context.Background()
	s := &aiResultStore{jobs: f.store}
	f.input.AssetIDs = []string{selectionA}
	f.input.MaxCalls = 3
	album := "33333333-3333-4333-8333-333333333333"
	var ids []string
	for n, key := range []string{"dated", "legacy", "invalid-date"} {
		f.input.Key = key
		job, err := f.store.Submit(ctx, f.input)
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, job.ID)
		if n != 1 {
			date := "2026-09-20T00:30:00+14:00"
			if n == 2 {
				date = "2026-02-30T12:00:00"
			}
			selectionSQL(t, f.db, `INSERT INTO ai_job_launch VALUES(?,?,?,?,?,?)`, testUserID, job.ID, selectionA, date, album, "At launch")
		}
		if err = f.store.Cancel(ctx, testUserID, job.ID); err != nil {
			t.Fatal(err)
		}
	}
	seedAsset(t, f.db, selectionA, nil, nil, "2026-09-01")
	for _, tc := range []struct {
		name string
		q    review.Query
		want int
	}{
		{"recorded offset day", review.Query{Limit: 30, StartDate: "2026-09-20", EndDate: "2026-09-20"}, 1},
		{"no UTC or current fallback", review.Query{Limit: 30, StartDate: "2026-09-19", EndDate: "2026-09-19"}, 0},
		{"unknown dates", review.Query{Limit: 30, Undated: true}, 2},
		{"selected album only", review.Query{Limit: 30, Album: album}, 2},
		{"run", review.Query{Limit: 30, Job: ids[1]}, 1},
		{"asset", review.Query{Limit: 30, Asset: selectionB}, 0},
		{"execution", review.Query{Limit: 30, State: "failed"}, 0},
		{"outcome", review.Query{Limit: 30, Outcome: "unknown"}, 0},
	} {
		page, err := s.list(ctx, testUserID, tc.q)
		if err != nil || len(page.Items) != tc.want {
			t.Errorf("%s: got %d want %d (%v)", tc.name, len(page.Items), tc.want, err)
		}
	}
}
