package main

import (
	"context"
	"immich-places-backend/internal/ai/review"
	"testing"
	"time"
)

func TestAIResultPaginationExcludesLaterTiedAndBackwardCompletions(t *testing.T) {
	f := newAIJobFixture(t)
	ctx := context.Background()
	s := &aiResultStore{jobs: f.store}
	job, err := f.store.Submit(ctx, f.input)
	if err != nil {
		t.Fatal(err)
	}
	if err = f.store.Cancel(ctx, testUserID, job.ID); err != nil {
		t.Fatal(err)
	}
	q := review.Query{Limit: 1}
	first, err := s.list(ctx, testUserID, q)
	if err != nil || len(first.Items) != 1 || first.NextCursor == "" {
		t.Fatal("missing bounded continuation", err)
	}
	for n, key := range []string{"tied", "backward"} {
		f.input.Key = key
		f.now = f.now.Add(-time.Duration(n) * time.Hour)
		fresh, err := f.store.Submit(ctx, f.input)
		if err != nil {
			t.Fatal(err)
		}
		if err = f.store.Cancel(ctx, testUserID, fresh.ID); err != nil {
			t.Fatal(err)
		}
	}
	q.Cursor = first.NextCursor
	next, err := s.list(ctx, testUserID, q)
	if err != nil || len(next.Items) != 1 || next.Items[0].JobID != job.ID || next.Items[0].ID == first.Items[0].ID || next.NextCursor != "" {
		t.Fatal("later completion entered stable pages", next, err)
	}
	if _, err = s.list(ctx, "foreign", q); err == nil {
		t.Fatal("foreign cursor accepted")
	}
	q.State = "failed"
	if _, err = s.list(ctx, testUserID, q); err == nil {
		t.Fatal("cursor changed query")
	}
	q = review.Query{Limit: 30}
	refreshed, err := s.list(ctx, testUserID, q)
	if err != nil || len(refreshed.Items) != 6 {
		t.Fatal("refresh omitted newer work", err)
	}
}
