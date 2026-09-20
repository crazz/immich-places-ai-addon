package main

import (
	"context"
	"encoding/json"
	"immich-places-backend/internal/ai/review"
	"testing"
)

func TestAIResultHTTPProtectsBoundedDisabledReadsAndItemDetails(t *testing.T) {
	f, p, req := productionFixture(t)
	productionSession(t, f)
	ctx := context.Background()
	job, err := p.submit(ctx, testUserID, req)
	if err != nil {
		t.Fatal(err)
	}
	if err = p.store.Cancel(ctx, testUserID, job.ID); err != nil {
		t.Fatal(err)
	}
	p.store.enabled = false
	h := newAIResultHandler(&aiResultStore{jobs: p.store}, f.image.service)
	before, _ := f.image.counts()
	rec := aiRequest(h, "GET", "/ai/results", "", "", true)
	var page review.Page
	if rec.Code != 200 || json.Unmarshal(rec.Body.Bytes(), &page) != nil || len(page.Items) != 1 || rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("history route unavailable", rec.Code, rec.Body.String())
	}
	path := "/ai/jobs/" + job.ID + "/items/" + job.Items[0].ID + "/result"
	rec = aiRequest(h, "GET", path, "", "", true)
	if rec.Code != 200 {
		t.Fatal("safe failed detail unavailable", rec.Code)
	}
	for _, query := range []string{"limit=101", "state=running", "unknown=value", "cursor=forged", "limit=1&limit=2"} {
		if rec = aiRequest(h, "GET", "/ai/results?"+query, "", "", true); rec.Code != 400 {
			t.Errorf("bad query status %s: %d", query, rec.Code)
		}
	}
	if rec = aiRequest(h, "GET", "/ai/results", "", "", false); rec.Code != 401 {
		t.Fatal("unauthenticated read", rec.Code)
	}
	for _, path := range []string{"/ai/results/" + selectionB, "/ai/jobs/" + job.ID + "/items/" + selectionB + "/result"} {
		if rec = aiRequest(h, "GET", path, "", "", true); rec.Code != 404 {
			t.Fatal("missing/foreign existence disclosed", rec.Code)
		}
	}
	after, bad := f.image.counts()
	if before != after || bad != 0 || f.hits.Load() != 0 {
		t.Fatal("history contacted upstream")
	}
}
