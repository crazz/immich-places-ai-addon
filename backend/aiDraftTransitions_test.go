package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"immich-places-backend/internal/ai/drafts"
)

func acceptedDraft(t *testing.T, f *aiJobFixture, id string) drafts.Draft {
	t.Helper()
	rec := aiRequest(draftHandler(f), "POST", "/ai/results/"+id+"/draft", `{}`, aiTestOrigin, true)
	var value drafts.Draft
	if rec.Code != 200 || json.Unmarshal(rec.Body.Bytes(), &value) != nil {
		t.Fatal(rec.Code, rec.Body.String())
	}
	return value
}
func draftPatch(f *aiJobFixture, id, revision, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("PATCH", "/ai/drafts/"+id, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", aiTestOrigin)
	if revision != "" {
		req.Header.Set("If-Match", revision)
	}
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "ai-session"})
	rec := httptest.NewRecorder()
	draftHandler(f).ServeHTTP(rec, req)
	return rec
}
func TestAIDraftRejectReopenKeepsRevisionHistoryAndReacceptance(t *testing.T) {
	f, id := draftFixture(t)
	value := acceptedDraft(t, f, id)
	rejected := draftPatch(f, value.ID, `"1"`, `{"state":"rejected"}`)
	if rejected.Code != 200 {
		t.Fatalf("reject %d %s", rejected.Code, rejected.Body.String())
	}
	if err := json.Unmarshal(rejected.Body.Bytes(), &value); err != nil || value.Revision != 2 || value.State != "rejected" {
		t.Fatal(value, err)
	}
	reopened := draftPatch(f, value.ID, `"2"`, `{"state":"draft"}`)
	if reopened.Code != 200 {
		t.Fatal("reopen", reopened.Code)
	}
	latest := acceptedDraft(t, f, id)
	if latest.Revision != 3 || latest.State != "draft" {
		t.Fatal("reaccept overwrote edits", latest)
	}
	var count int
	if err := f.db.db.QueryRow("SELECT count(*) FROM ai_draft_revisions WHERE draftID=?", value.ID).Scan(&count); err != nil || count != 3 {
		t.Fatal(count, err)
	}
	if _, err := f.db.db.Exec("UPDATE ai_draft_revisions SET content='{}' WHERE draftID=?", value.ID); err == nil {
		t.Fatal("revision mutated")
	}
}

func TestAIDraftUnknownRequiresExplicitCameraAndGPSSelection(t *testing.T) {
	f, id := draftFixture(t)
	value := acceptedDraft(t, f, id)
	rec := draftPatch(f, value.ID, `"1"`, `{"state":"staged","fields":["gps"]}`)
	if rec.Code != 400 {
		t.Fatal("invented unknown camera", rec.Code)
	}
	rec = draftPatch(f, value.ID, `"1"`, `{"camera":{"latitude":0,"longitude":0},"fields":["gps"],"state":"staged"}`)
	if rec.Code != 200 {
		t.Fatalf("explicit zero camera: %d %s", rec.Code, rec.Body.String())
	}
	if json.Unmarshal(rec.Body.Bytes(), &value) != nil || value.Camera == nil || value.Camera.Latitude != 0 || value.Camera.Longitude != 0 || value.State != "staged" || value.AssetID != selectionA {
		t.Fatal("wrong stage", value)
	}
	rec = draftPatch(f, value.ID, `"2"`, `{"camera":{"latitude":1,"longitude":2}}`)
	if rec.Code != 200 || json.Unmarshal(rec.Body.Bytes(), &value) != nil || value.State != "draft" || value.Revision != 3 {
		t.Fatal("edit retained stage", rec.Code, value)
	}
}

func TestAIDraftConcurrentEditsRequireExactRevision(t *testing.T) {
	f, id := draftFixture(t)
	value := acceptedDraft(t, f, id)
	if rec := draftPatch(f, value.ID, "", `{"state":"rejected"}`); rec.Code != 428 {
		t.Fatal("missing precondition", rec.Code)
	}
	start := make(chan struct{})
	done := make(chan int, 2)
	for _, body := range []string{`{"camera":{"latitude":1,"longitude":2}}`, `{"camera":{"latitude":3,"longitude":4}}`} {
		go func(body string) { <-start; done <- draftPatch(f, value.ID, `"1"`, body).Code }(body)
	}
	close(start)
	a, b := <-done, <-done
	if !((a == 200 && b == 412) || (a == 412 && b == 200)) {
		t.Fatal("lost concurrency fence", a, b)
	}
	latest := acceptedDraft(t, f, id)
	if latest.Revision != 2 {
		t.Fatal("lost response replayed", latest.Revision)
	}
}

func TestAIDraftChangedStagedContentRequiresSeparateRestaging(t *testing.T) {
	f, id := draftFixture(t)
	value := acceptedDraft(t, f, id)
	staged := draftPatch(f, value.ID, `"1"`, `{"camera":{"latitude":0,"longitude":0},"fields":["gps"],"state":"staged"}`)
	if staged.Code != 200 {
		t.Fatal(staged.Code)
	}
	changed := draftPatch(f, value.ID, `"2"`, `{"camera":{"latitude":1,"longitude":2},"state":"staged"}`)
	if changed.Code != 200 || json.Unmarshal(changed.Body.Bytes(), &value) != nil || value.State != "draft" || value.Revision != 3 {
		t.Fatal("content change retained stage", changed.Code, value)
	}
	if rec := draftPatch(f, value.ID, `"3"`, `{"state":"staged"}`); rec.Code != 200 {
		t.Fatal(rec.Code)
	}
}
