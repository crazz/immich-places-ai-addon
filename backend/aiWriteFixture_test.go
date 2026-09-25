package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writeback"
	"immich-places-backend/internal/ai/writepreview"
)

type aiWriteFixture struct {
	f            *aiJobFixture
	writer       *aiWriteStore
	image        *aiImageFixture
	draft        drafts.Draft
	preview      writepreview.Preview
	op           writeback.Operation
	mu           sync.Mutex
	meta         map[string]any
	sends, reads int
	handle       func(http.ResponseWriter, *http.Request) bool
	enabled      bool
}

func newAIWriteFixture(t *testing.T) *aiWriteFixture {
	t.Helper()
	f, image, store, draft, meta := writePreviewFixture(t)
	w := &aiWriteFixture{f: f, image: image, draft: draft, meta: meta, enabled: true}
	w.preview = savedWritePreview(t, f, image, draft)
	server := httptest.NewServer(http.HandlerFunc(func(out http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/assets/"+draft.AssetID {
			t.Error("unexpected request", r.Method, r.URL.Path)
			out.WriteHeader(403)
			return
		}
		if r.Header.Get("x-api-key") != "private-immich-image-key" {
			out.WriteHeader(403)
			return
		}
		w.mu.Lock()
		if r.Method == "PATCH" {
			w.sends++
		} else {
			w.reads++
		}
		w.mu.Unlock()
		if w.handle != nil && w.handle(out, r) {
			return
		}
		w.mu.Lock()
		defer w.mu.Unlock()
		if r.Method == "PATCH" {
			var body map[string]float64
			if json.NewDecoder(r.Body).Decode(&body) != nil || len(body) != 2 {
				t.Error("invalid fields", body)
			}
			w.meta["exifInfo"] = body
		}
		if err := json.NewEncoder(out).Encode(w.meta); err != nil {
			t.Error(err)
		}
	}))
	t.Cleanup(server.Close)
	image.service.endpoint = server.URL
	w.writer = &aiWriteStore{drafts: store, enabled: func() bool { w.mu.Lock(); defer w.mu.Unlock(); return w.enabled }, profile: "immich-v3.2.2", images: image.service, sync: newSyncService(f.db, nil, nil)}
	store.results.writer = w.writer
	store.results.origin = aiTestOrigin
	var err error
	w.op, err = w.writer.confirm(context.Background(), testUserID, writeback.Confirmation{PreviewID: w.preview.Plan.ID, Digest: w.preview.Digest, Key: "operation-key"})
	if err != nil {
		t.Fatal(err)
	}
	return w
}
func (w *aiWriteFixture) counts() (int, int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.sends, w.reads
}
func (w *aiWriteFixture) status(t *testing.T) writeback.Operation {
	t.Helper()
	op, err := w.writer.get(context.Background(), testUserID, w.op.ID, false)
	if err != nil {
		t.Fatal(err)
	}
	return op
}
func (w *aiWriteFixture) run(t *testing.T) {
	t.Helper()
	if err := w.writer.runOne(context.Background(), testUserID, w.op.ID); err != nil {
		t.Fatal(err)
	}
}
