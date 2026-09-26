package main

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"testing"

	"github.com/google/uuid"
	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writepreview"
)

type aiMirrorFixture struct {
	*aiStandardWriteFixture
	mu                         sync.Mutex
	namespace                  json.RawMessage
	standardSends, mirrorSends int
	handle                     func(http.ResponseWriter, *http.Request) bool
}

func mirrorWriteFixture(t *testing.T) *aiMirrorFixture {
	t.Helper()
	f := &aiMirrorFixture{aiStandardWriteFixture: standardWriteFixture(t, []string{"gps", "description"})}
	f.writer.capabilities.Capabilities = []string{"description", "metadata"}
	previous := f.image.handle
	f.image.handle = func(w http.ResponseWriter, r *http.Request) bool {
		if f.handle != nil && f.handle(w, r) {
			return true
		}
		f.mu.Lock()
		defer f.mu.Unlock()
		if r.URL.Path == "/api/assets/"+f.draft.AssetID+"/metadata" {
			if r.Method == "GET" {
				items := []map[string]any{{"key": "unrelated", "value": map[string]any{"keep": true}}}
				if len(f.namespace) > 0 {
					items = append(items, map[string]any{"key": writepreview.MirrorNamespace, "value": f.namespace})
				}
				_ = json.NewEncoder(w).Encode(items)
			} else {
				f.mirrorSends++
				t.Error("unconfigured metadata mutation")
				w.WriteHeader(500)
			}
			return true
		}
		return previous != nil && previous(w, r)
	}
	ctx := context.Background()
	d, err := f.writer.drafts.edit(ctx, testUserID, f.draft.ID, f.draft.Revision, drafts.Edit{Mirror: json.RawMessage(`{"languages":["en"]}`)})
	if err != nil {
		t.Fatal(err)
	}
	d, err = f.writer.drafts.edit(ctx, testUserID, d.ID, d.Revision, drafts.Edit{State: "staged"})
	if err != nil {
		t.Fatal(err)
	}
	f.draft = d
	s := &aiWritePreviewSession{drafts: f.writer.drafts, images: f.image.service, mirrorDisclosure: writepreview.MirrorDisclosure}
	f.preview, err = writepreview.Create(ctx, s, s, s, testUserID, d.ID, d.Revision, uuid.NewString(), f.f.store.now)
	if err != nil {
		t.Fatal(err)
	}
	return f
}
