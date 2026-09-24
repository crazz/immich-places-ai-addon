package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"immich-places-backend/internal/ai/jobs"
)

func draftFixture(t *testing.T) (*aiJobFixture, string) {
	t.Helper()
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
	hash := sha256.Sum256([]byte("ai-session"))
	if err = f.db.createSession(ctx, hex.EncodeToString(hash[:]), testUserID, time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	return f, id
}

func draftHandler(f *aiJobFixture) http.Handler {
	return newAIResultHandler(&aiResultStore{jobs: f.store, origin: aiTestOrigin}, nil)
}

func TestAIDraftAcceptSurvivesReopenWithoutChangingAnalysis(t *testing.T) {
	f, id := draftFixture(t)
	before, err := f.store.ReadAnalysis(context.Background(), testUserID, id)
	if err != nil {
		t.Fatal(err)
	}
	rec := aiRequest(draftHandler(f), "POST", "/ai/results/"+id+"/draft", `{}`, aiTestOrigin, true)
	if rec.Code != 200 {
		t.Fatalf("accept: %d %s", rec.Code, rec.Body.String())
	}
	var saved struct {
		ID       string `json:"id"`
		Revision int    `json:"revision"`
		State    string `json:"state"`
		Camera   any    `json:"camera"`
		Baseline struct {
			Status string `json:"status"`
		} `json:"baseline"`
	}
	if err = json.Unmarshal(rec.Body.Bytes(), &saved); err != nil || saved.ID == "" || saved.Revision != 1 || saved.State != "draft" || saved.Camera != nil || saved.Baseline.Status != "unavailable" {
		t.Fatal("invalid draft", saved, err)
	}
	f.reopen(t)
	f.store.enabled = false
	restored := aiRequest(draftHandler(f), "GET", "/ai/drafts/"+saved.ID, "", "", true)
	if restored.Code != 200 || restored.Body.String() != rec.Body.String() {
		t.Fatal("draft not durable", restored.Code, restored.Body.String())
	}
	var after []byte
	err = f.db.db.QueryRow("SELECT payload FROM ai_analyses WHERE userID=? AND id=?", testUserID, id).Scan(&after)
	if err != nil || string(before.Payload) != string(after) {
		t.Fatal("original changed", err)
	}
}

func TestAIDraftAcceptanceIsIdempotent(t *testing.T) {
	f, id := draftFixture(t)
	h := draftHandler(f)
	first := aiRequest(h, "POST", "/ai/results/"+id+"/draft", `{}`, aiTestOrigin, true)
	second := aiRequest(h, "POST", "/ai/results/"+id+"/draft", `{}`, aiTestOrigin, true)
	if first.Code != 200 || second.Code != 200 || first.Body.String() != second.Body.String() {
		t.Fatal("acceptance was not idempotent")
	}
	var count int
	if err := f.db.db.QueryRow("SELECT count(*) FROM ai_drafts").Scan(&count); err != nil || count != 1 {
		t.Fatal(count, err)
	}
}

func TestAIDraftMutationRequiresOriginAndBoundedPrivateInput(t *testing.T) {
	f, id := draftFixture(t)
	for _, tc := range []struct {
		origin, body string
		auth         bool
		code         int
	}{
		{"", `{}`, true, 403}, {"https://foreign.example", `{}`, true, 403},
		{aiTestOrigin, `{}`, false, 401}, {aiTestOrigin, `{"assetId":"other"}`, true, 400},
		{aiTestOrigin, `null`, true, 400}, {aiTestOrigin, `{}`, true, 200},
	} {
		rec := aiRequest(draftHandler(f), "POST", "/ai/results/"+id+"/draft", tc.body, tc.origin, tc.auth)
		if rec.Code != tc.code {
			t.Fatalf("origin/auth/body: got %d want %d", rec.Code, tc.code)
		}
		if rec.Header().Get("Cache-Control") != "no-store" {
			t.Fatal("private response cacheable")
		}
	}
}

func TestAIDraftRejectsUnavailableAndCorruptProposals(t *testing.T) {
	f, id := draftFixture(t)
	for _, target := range []string{selectionA, "invalid"} {
		rec := aiRequest(draftHandler(f), "POST", "/ai/results/"+target+"/draft", `{}`, aiTestOrigin, true)
		if rec.Code != 404 {
			t.Fatal("invented proposal", rec.Code)
		}
	}
	selectionSQL(t, f.db, "DROP TRIGGER ai_analyses_immutable")
	selectionSQL(t, f.db, "UPDATE ai_analyses SET payload='{}' WHERE id=?", id)
	rec := aiRequest(draftHandler(f), "POST", "/ai/results/"+id+"/draft", `{}`, aiTestOrigin, true)
	if rec.Code != 404 {
		t.Fatal("accepted corrupt proposal", rec.Code)
	}
	var count int
	if err := f.db.db.QueryRow("SELECT count(*) FROM ai_drafts").Scan(&count); err != nil || count != 0 {
		t.Fatal(count, err)
	}
}

func TestAIDraftHistoryProjectsSavedReviewState(t *testing.T) {
	f, id := draftFixture(t)
	value := acceptedDraft(t, f, id)
	rec := draftPatch(f, value.ID, `"1"`, `{"state":"rejected"}`)
	if rec.Code != 200 {
		t.Fatal(rec.Code)
	}
	detail := aiRequest(draftHandler(f), "GET", "/ai/results/"+id, "", "", true)
	var result struct {
		Entry struct {
			ReviewState   string  `json:"reviewState"`
			DraftID       *string `json:"draftId"`
			DraftRevision *int    `json:"draftRevision"`
		} `json:"entry"`
	}
	if detail.Code != 200 || json.Unmarshal(detail.Body.Bytes(), &result) != nil {
		t.Fatal("detail unavailable")
	}
	if result.Entry.ReviewState != "rejected" || result.Entry.DraftID == nil || *result.Entry.DraftID != value.ID || result.Entry.DraftRevision == nil || *result.Entry.DraftRevision != 2 {
		t.Fatal("saved review state missing", result.Entry)
	}
}
