package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"sync/atomic"
	"testing"

	"immich-places-backend/internal/ai/drafts"
)

func TestAIDraftLocalOfflineDecisionsMakeNoUpstreamRequests(t *testing.T) {
	f, image, store, value, _ := draftBaselineFixture(t)
	var calls atomic.Int32
	image.handle = func(w http.ResponseWriter, r *http.Request) bool {
		calls.Add(1)
		http.Error(w, "offline", 503)
		return true
	}
	handler := newAIResultHandler(store.results, image.service)
	store.results.origin = aiTestOrigin
	for _, body := range []string{`{"camera":{"latitude":0,"longitude":0},"fields":["gps"]}`, `{"state":"staged"}`, `{"state":"rejected"}`, `{"state":"draft"}`} {
		rec := draftPatch(f, value.ID, `"`+strconv.Itoa(value.Revision)+`"`, body)
		if rec.Code != 200 || json.Unmarshal(rec.Body.Bytes(), &value) != nil {
			t.Fatal(rec.Code, rec.Body.String())
		}
	}
	if rec := aiRequest(handler, "POST", "/ai/results/"+value.AnalysisID+"/draft", `{}`, aiTestOrigin, true); rec.Code != 200 {
		t.Fatal(rec.Code)
	}
	if value.Baseline.Status != "unavailable" || calls.Load() != 0 {
		t.Fatal("local work read upstream", value.Baseline, calls.Load())
	}
	for _, cause := range []string{"offline", "hidden", "removed"} {
		if cause == "hidden" {
			selectionSQL(t, f.db, "UPDATE assets SET isHidden=1 WHERE userID=?", testUserID)
		}
		if cause == "removed" {
			selectionSQL(t, f.db, "DELETE FROM assets WHERE userID=?", testUserID)
		}
		before := calls.Load()
		if _, err := store.observe(context.Background(), testUserID, value.ID, value.Revision, image.service); err == nil {
			t.Fatal("unavailable source observed")
		}
		if cause != "offline" && calls.Load() != before {
			t.Fatal("unauthorized source read")
		}
		saved, err := store.edit(context.Background(), testUserID, value.ID, value.Revision, drafts.Edit{State: "draft"})
		if err != nil {
			t.Fatal(err)
		}
		value = saved
	}
}
