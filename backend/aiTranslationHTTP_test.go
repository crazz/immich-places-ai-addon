package main

import (
	"context"
	"encoding/json"
	"testing"
)

func TestAITranslationHTTPAdmitsAndReadsPrivateRunWithoutDispatch(t *testing.T) {
	_, s, req := translationFixture(t)
	h := newAITranslationHandler(s, aiTestOrigin)
	body, _ := json.Marshal(req)
	rec := aiRequest(h, "POST", "/ai/translations", string(body), aiTestOrigin, true)
	if rec.Code != 200 {
		t.Fatalf("translation endpoint unavailable: %d %s", rec.Code, rec.Body.String())
	}
	var run aiTranslationRun
	if json.Unmarshal(rec.Body.Bytes(), &run) != nil || run.ID == "" || run.Items[0].State != "queued" {
		t.Fatal("invalid admission response")
	}
	read := aiRequest(h, "GET", "/ai/translations/"+run.ID, "", "", true)
	if read.Code != 200 || read.Body.String() != rec.Body.String() || read.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("private read changed or cached run", read.Code)
	}
}

func TestAITranslationHTTPExplicitAdoptionAndCancellation(t *testing.T) {
	f, s, req := translationFixture(t)
	run, err := s.submit(context.Background(), testUserID, req)
	if err != nil {
		t.Fatal(err)
	}
	selectionSQL(t, f.db, "UPDATE ai_translation_items SET state='complete',text='Reviewed translation.' WHERE runID=? AND language='en'", run.ID)
	h := newAITranslationHandler(s, aiTestOrigin)
	body, _ := json.Marshal(map[string]any{"revision": req.Revision, "languages": []string{"en"}})
	rec := aiRequest(h, "POST", "/ai/translations/"+run.ID+"/adopt", string(body), aiTestOrigin, true)
	if rec.Code != 200 {
		t.Fatal("adoption endpoint unavailable", rec.Code, rec.Body.String())
	}
	rec = aiRequest(h, "POST", "/ai/translations/"+run.ID+"/cancel", "{}", aiTestOrigin, true)
	if rec.Code != 200 {
		t.Fatal("cancel endpoint unavailable", rec.Code, rec.Body.String())
	}
	var saved aiTranslationRun
	if json.Unmarshal(rec.Body.Bytes(), &saved) != nil || saved.Items[0].State != "complete" || saved.Items[1].State != "canceled" {
		t.Fatal("cancel discarded completed text", saved)
	}
}

func TestAITranslationHTTPListsRetainedDraftHistory(t *testing.T) {
	f, s, req := translationFixture(t)
	ctx := context.Background()
	first, err := s.submit(ctx, testUserID, req)
	if err != nil {
		t.Fatal(err)
	}
	selectionSQL(t, f.db, "UPDATE ai_translation_items SET state='failed' WHERE runID=?", first.ID)
	req.Key = "second"
	f.now = f.now.Add(1)
	second, err := s.submit(ctx, testUserID, req)
	if err != nil {
		t.Fatal(err)
	}
	h := newAITranslationHandler(s, aiTestOrigin)
	for _, tc := range []struct {
		before string
		ids    []string
	}{{"", []string{second.ID, first.ID}}, {second.ID, []string{first.ID}}} {
		rec := aiRequest(h, "GET", "/ai/translations?draftId="+req.DraftID+"&before="+tc.before, "", "", true)
		var page struct {
			IDs  []string `json:"ids"`
			Next string   `json:"next"`
		}
		if rec.Code != 200 || json.Unmarshal(rec.Body.Bytes(), &page) != nil || len(page.IDs) != len(tc.ids) {
			t.Fatal("history list unavailable", rec.Code, rec.Body.String())
		}
		for i, id := range page.IDs {
			if id != tc.ids[i] {
				t.Fatal("unstable history order")
			}
		}
	}
}

func TestAITranslationHTTPRejectsForeignOrMalformedRequests(t *testing.T) {
	_, s, req := translationFixture(t)
	h := newAITranslationHandler(s, aiTestOrigin)
	body, _ := json.Marshal(req)
	for _, tc := range []struct {
		body, origin string
		auth         bool
		status       int
	}{
		{string(body), aiTestOrigin, false, 401}, {string(body), "", true, 403},
		{`null`, aiTestOrigin, true, 400}, {`{"hiddenImage":"secret"}`, aiTestOrigin, true, 400},
		{`{"confirmed":true,"confirmed":false}`, aiTestOrigin, true, 400},
	} {
		if rec := aiRequest(h, "POST", "/ai/translations", tc.body, tc.origin, tc.auth); rec.Code != tc.status {
			t.Fatal("unsafe admission accepted", rec.Code, tc.status)
		}
	}
	if rec := aiRequest(h, "GET", "/ai/translations/foreign-id", "", "", true); rec.Code != 404 {
		t.Fatal("foreign read exposed", rec.Code)
	}
}
