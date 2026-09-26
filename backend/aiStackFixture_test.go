package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
)

type aiStackFixture struct {
	*aiStandardWriteFixture
	mu       sync.Mutex
	stackID  string
	members  []string
	metadata map[string]map[string]any
	handle   func(http.ResponseWriter, *http.Request) bool
}

func stackWriteFixture(t *testing.T) *aiStackFixture {
	t.Helper()
	base := standardWriteFixture(t, []string{"gps", "description"})
	f := &aiStackFixture{aiStandardWriteFixture: base, stackID: uuid.NewString(), members: []string{base.draft.AssetID, selectionB, selectionID(3)}, metadata: map[string]map[string]any{}}
	for i, id := range f.members {
		seedAsset(t, base.f.db, id, nil, nil, "2026-09-20")
		if _, err := base.f.db.db.Exec(`UPDATE assets SET stackID=?,stackPrimaryAssetID=? WHERE userID=? AND immichID=?`, f.stackID, base.draft.AssetID, testUserID, id); err != nil {
			t.Fatal(err)
		}
		meta := base.image.metadata()
		if id == base.draft.AssetID {
			meta = base.meta
		} else {
			meta["exifInfo"] = map[string]any{"latitude": float64(i), "longitude": nil, "description": "sibling private text"}
		}
		meta["id"] = id
		meta["stack"] = map[string]any{"id": f.stackID, "primaryAssetId": base.draft.AssetID, "assetCount": len(f.members)}
		f.metadata[id] = meta
	}
	f.writer.capabilities.Capabilities = []string{"description", "stack_gps"}
	f.image.handle = func(w http.ResponseWriter, r *http.Request) bool {
		if f.handle != nil && f.handle(w, r) {
			return true
		}
		f.mu.Lock()
		defer f.mu.Unlock()
		if r.URL.Path == "/api/stacks/"+f.stackID {
			members := make([]map[string]any, 0, len(f.members))
			for _, id := range f.members {
				members = append(members, f.metadata[id])
			}
			if err := json.NewEncoder(w).Encode(map[string]any{"id": f.stackID, "primaryAssetId": f.draft.AssetID, "assets": members}); err != nil {
				t.Error(err)
			}
			return true
		}
		if meta, ok := f.metadata[strings.TrimPrefix(r.URL.Path, "/api/assets/")]; ok {
			if err := json.NewEncoder(w).Encode(meta); err != nil {
				t.Error(err)
			}
			return true
		}
		return false
	}
	return f
}
