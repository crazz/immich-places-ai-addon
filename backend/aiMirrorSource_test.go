package main

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"
	"testing"
)

func TestAIMirrorBaselineUsesAuthorizedCompleteListAndNeverTreats400AsAbsent(t *testing.T) {
	_, image, store, draft, _ := draftBaselineFixture(t)
	var calls atomic.Int32
	status, body := http.StatusOK, `[]`
	image.handle = func(w http.ResponseWriter, r *http.Request) bool {
		calls.Add(1)
		if r.Method != "GET" || r.URL.Path != "/api/assets/"+draft.AssetID+"/metadata" || r.Header.Get("x-api-key") == "" {
			t.Error("wrong metadata read scope")
		}
		w.WriteHeader(status)
		fmt.Fprint(w, body)
		return true
	}
	baseline, authority, err := store.mirrorSource(context.Background(), testUserID, draft.AssetID, image.service)
	if err != nil || baseline.Present || authority.key == "" || calls.Load() != 1 {
		t.Fatal("complete-list absence not observed", baseline, err)
	}
	for _, tc := range []struct {
		status int
		body   string
	}{
		{400, `{"message":"not found"}`}, {403, `{}`}, {503, `[]`}, {200, `null`}, {200, `[{"key":"x","value":{}}`},
	} {
		status, body = tc.status, tc.body
		if _, _, err := store.mirrorSource(context.Background(), testUserID, draft.AssetID, image.service); err == nil {
			t.Fatal("unavailable read fabricated an empty namespace")
		}
	}
	before := calls.Load()
	if _, _, err := store.mirrorSource(context.Background(), "other-owner", draft.AssetID, image.service); err == nil || calls.Load() != before {
		t.Fatal("unauthorized namespace request escaped local guard")
	}
}
