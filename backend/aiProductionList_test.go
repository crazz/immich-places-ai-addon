package main

import (
	"context"
	"encoding/json"
	"net/url"
	"testing"
	"time"

	"immich-places-backend/internal/ai/jobs"
)

func TestAIProductionListingUsesStableOwnerBoundKeyset(t *testing.T) {
	f, p, req := productionFixture(t)
	productionSession(t, f)
	h := newAIJobHandler(p, aiTestOrigin)
	clock := time.Now()
	p.store.now = func() time.Time { return clock }
	ids := []string{}
	for _, key := range []string{"one", "two", "three"} {
		req.IdempotencyKey = key
		j, err := p.submit(context.Background(), testUserID, req)
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, j.ID)
		clock = clock.Add(time.Second)
	}
	read := func(path string) (int, []string, string) {
		t.Helper()
		r := aiRequest(h, "GET", path, "", "", true)
		var page struct {
			Items []struct {
				ID string `json:"id"`
			} `json:"items"`
			NextCursor string `json:"nextCursor"`
		}
		if r.Code == 200 && json.Unmarshal(r.Body.Bytes(), &page) != nil {
			t.Fatal("bad page")
		}
		items := []string{}
		for _, j := range page.Items {
			items = append(items, j.ID)
		}
		return r.Code, items, page.NextCursor
	}
	status, first, cursor := read("/ai/jobs?limit=2")
	if status != 200 || len(first) != 2 || first[0] != ids[2] || first[1] != ids[1] || cursor == "" {
		t.Fatal("first page", status, first, cursor)
	}
	req.IdempotencyKey = "newer"
	if _, err := p.submit(context.Background(), testUserID, req); err != nil {
		t.Fatal(err)
	}
	status, next, end := read("/ai/jobs?limit=2&cursor=" + url.QueryEscape(cursor))
	if status != 200 || len(next) != 1 || next[0] != ids[0] || end != "" {
		t.Fatal("unstable continuation", status, next, end)
	}
	for _, query := range []string{"limit=101", "limit=0", "limit=2&limit=3", "cursor=garbage", "unknown=1"} {
		status, _, _ = read("/ai/jobs?" + query)
		if status != 400 {
			t.Fatal("invalid page accepted", query, status)
		}
	}
	if _, err := p.list(context.Background(), "foreign", 2, cursor); err != jobs.ErrInvalid {
		t.Fatal("foreign cursor accepted", err)
	}
	p.store.enabled = false
	status, _, _ = read("/ai/jobs?limit=2")
	if status != 200 {
		t.Fatal("disabled list failed")
	}
}
