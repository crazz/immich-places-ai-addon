package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"

	"immich-places-backend/internal/ai/providers"
	"immich-places-backend/internal/aiadapters/providerhttp"
)

func translationProvider(t *testing.T, f *aiJobFixture, s *aiTranslationStore, handle http.HandlerFunc) *providers.Dispatcher {
	t.Helper()
	server := httptest.NewServer(handle)
	t.Cleanup(server.Close)
	input := providerInput()
	input.BaseURL = server.URL + "/v1"
	if _, err := f.db.updateAIProvider(context.Background(), testUserID, "profile", 1, input); err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal([]map[string]any{{"baseURL": server.URL + "/v1", "addressClass": "local", "allowedCIDRs": []string{"127.0.0.0/8"}}})
	policy, err := providers.ParseEgressPolicy(string(raw))
	if err != nil {
		t.Fatal(err)
	}
	s.fingerprint = policyFingerprint(policy)
	return newAIProviderDispatcher(f.db, true, policy, providerhttp.New(providerhttp.Options{MaxResponseBytes: 64 << 10}))
}

func TestAITranslationExecutionKeepsPartialResultsWithoutRetryOrDraftMutation(t *testing.T) {
	f, s, req := translationFixture(t)
	var hits atomic.Int32
	dispatcher := translationProvider(t, f, s, func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		raw, _ := io.ReadAll(r.Body)
		if strings.Contains(string(raw), "image_url") || strings.Contains(string(raw), req.DraftID) || !strings.Contains(string(raw), req.Basis) {
			t.Error("unapproved translation payload")
		}
		if strings.Contains(string(raw), "language uk") {
			http.Error(w, "private provider failure", 500)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"finish_reason": "stop", "message": map[string]any{"role": "assistant", "content": `{"language":"en","status":"complete","text":"An uncertain bridge."}`}}}})
	})
	req.ProfileRevision = 2
	ctx := context.Background()
	before, _ := s.drafts.get(ctx, testUserID, req.DraftID)
	run, err := s.submit(ctx, testUserID, req)
	if err != nil {
		t.Fatal(err)
	}
	for range 2 {
		worked, err := s.runOne(ctx, dispatcher)
		if err != nil || !worked {
			t.Fatal("translation not executed", worked, err)
		}
	}
	if worked, err := s.runOne(ctx, dispatcher); err != nil || worked || hits.Load() != 2 {
		t.Fatal("hidden regeneration", hits.Load(), worked, err)
	}
	after, err := s.submit(ctx, testUserID, req)
	if err != nil || after.ID != run.ID || after.Items[0].State != "complete" || after.Items[0].Text == nil || after.Items[1].State != "failed" || strings.Contains(after.Items[1].Failure, "private") {
		t.Fatal("partial outcomes lost", after, err)
	}
	draft, _ := s.drafts.get(ctx, testUserID, req.DraftID)
	if !reflect.DeepEqual(before, draft) {
		t.Fatal("generation modified draft")
	}
}
