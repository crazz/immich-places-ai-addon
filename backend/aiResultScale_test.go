package main

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"immich-places-backend/internal/ai/review"
)

func TestAIResultHistoryUsesIndexedBoundedPagesForRetainedWorkload(t *testing.T) {
	f := newAIJobFixture(t)
	ctx := context.Background()
	f.input.AssetIDs = nil
	for n := 1; n <= 500; n++ {
		f.input.AssetIDs = append(f.input.AssetIDs, selectionID(n))
	}
	f.input.MaxCalls = 1500
	for n := 0; n < 20; n++ {
		f.input.Key = fmt.Sprintf("history-%d", n)
		job, err := f.store.Submit(ctx, f.input)
		if err != nil {
			t.Fatal(err)
		}
		if err = f.store.Cancel(ctx, testUserID, job.ID); err != nil {
			t.Fatal(err)
		}
	}
	s := &aiResultStore{jobs: f.store}
	q := review.Query{Limit: 100}
	seen := map[string]bool{}
	for pageNumber := 0; ; pageNumber++ {
		page, err := s.list(ctx, testUserID, q)
		if err != nil || len(page.Items) > 100 {
			t.Fatal("unbounded/failed page", err)
		}
		for _, entry := range page.Items {
			if seen[entry.ID] {
				t.Fatal("duplicate history item")
			}
			seen[entry.ID] = true
		}
		if page.NextCursor == "" {
			break
		}
		if pageNumber >= 100 {
			t.Fatal("pagination did not terminate")
		}
		q.Cursor = page.NextCursor
	}
	if len(seen) != 10000 {
		t.Fatal("history members lost", len(seen))
	}
	rows, err := f.db.db.QueryContext(ctx, `EXPLAIN QUERY PLAN `+aiResultSummarySQL+` WHERE h.userID=? AND h.installationID=? AND h.sequence<=? ORDER BY h.terminalAt DESC,h.jobID DESC,h.itemID DESC LIMIT ?`, testUserID, s.jobs.binding, 10000, 31)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var plan []string
	for rows.Next() {
		var id, parent, unused int
		var detail string
		if err = rows.Scan(&id, &parent, &unused, &detail); err != nil {
			t.Fatal(err)
		}
		plan = append(plan, detail)
	}
	if rows.Err() != nil {
		t.Fatal(rows.Err())
	}
	joined := strings.Join(plan, "\n")
	if !strings.Contains(joined, "ai_result_history_page") || strings.Contains(joined, "SCAN a ") {
		t.Fatal("history query lost bounded index path", joined)
	}
	t.Log(joined)
}
