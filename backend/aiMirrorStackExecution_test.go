package main

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"testing"

	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writeback"
	"immich-places-backend/internal/ai/writepreview"
)

func TestAIMirrorCombinedStackWritesOnlyAnalyzedNamespaceAndPreservesSiblingMetadata(t *testing.T) {
	f := stackWriteFixture(t)
	ctx := context.Background()
	f.writer.capabilities.Capabilities = []string{"description", "stack_gps", "metadata"}
	var mu sync.Mutex
	values := map[string][]map[string]any{}
	reads := map[string]int{}
	sends := map[string]int{}
	for _, id := range f.members {
		values[id] = []map[string]any{{"key": "unrelated", "value": map[string]any{"text": "keep private sibling metadata"}}}
	}
	handler := func(w http.ResponseWriter, r *http.Request) bool {
		if !strings.HasSuffix(r.URL.Path, "/metadata") {
			return false
		}
		id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/assets/"), "/metadata")
		mu.Lock()
		defer mu.Unlock()
		if _, ok := values[id]; !ok {
			t.Error("unexpected metadata asset")
			w.WriteHeader(404)
			return true
		}
		if r.Method == "GET" {
			reads[id]++
			_ = json.NewEncoder(w).Encode(values[id])
			return true
		}
		if r.Method != "PUT" || id != f.draft.AssetID {
			t.Error("metadata expanded to sibling or another method")
			w.WriteHeader(400)
			return true
		}
		var body struct {
			Items []map[string]any `json:"items"`
		}
		if json.NewDecoder(r.Body).Decode(&body) != nil || len(body.Items) != 1 || body.Items[0]["key"] != writepreview.MirrorNamespace {
			t.Error("expanded namespace payload")
			w.WriteHeader(400)
			return true
		}
		values[id] = append(values[id], body.Items[0])
		sends[id]++
		w.WriteHeader(200)
		return true
	}
	f.handle = handler
	d, err := f.writer.drafts.edit(ctx, testUserID, f.draft.ID, f.draft.Revision, drafts.Edit{Mirror: json.RawMessage(`{"languages":["en"]}`)})
	if err != nil {
		t.Fatal(err)
	}
	d, err = f.writer.drafts.edit(ctx, testUserID, d.ID, d.Revision, drafts.Edit{State: "staged"})
	if err != nil {
		t.Fatal(err)
	}
	review, err := f.writer.drafts.observeStack(ctx, testUserID, d.ID, d.Revision, f.members, f.image.service)
	if err != nil {
		t.Fatal(err)
	}
	preview, err := (&aiWritePreviewSession{drafts: f.writer.drafts, images: f.image.service, mirrorDisclosure: writepreview.MirrorDisclosure}).createStack(ctx, testUserID, d.ID, d.Revision, review.ID)
	if err != nil {
		t.Fatal(err)
	}
	op, err := f.writer.confirm(ctx, testUserID, writeback.Confirmation{PreviewID: preview.Plan.ID, Digest: preview.Digest, Key: "stack-mirror"})
	if err != nil {
		t.Fatal(err)
	}
	mutation := stackMutationServer(t, f)
	mutation.handle = handler
	for range len(f.members) + 1 {
		if err = f.writer.runOne(ctx, testUserID, op.ID); err != nil {
			t.Fatal(err)
		}
	}
	saved, err := f.writer.get(ctx, testUserID, op.ID, false)
	if err != nil || saved.Status != "succeeded" || !saved.Verified {
		t.Fatal("combined stack did not verify", err)
	}
	for _, id := range f.members {
		if mutation.sent(id) != 1 {
			t.Fatal("standard target sent more than once", id)
		}
		if id == f.draft.AssetID {
			if sends[id] != 1 || len(values[id]) != 2 || reads[id] < 1 {
				t.Fatal("analyzed mirror did not preserve unrelated key")
			}
			continue
		}
		if sends[id] != 0 || reads[id] != 0 || !reflect.DeepEqual(values[id], []map[string]any{{"key": "unrelated", "value": map[string]any{"text": "keep private sibling metadata"}}}) || f.metadata[id]["exifInfo"].(map[string]any)["description"] != "sibling private text" {
			t.Fatal("sibling metadata/description changed or was read", id)
		}
	}
}
