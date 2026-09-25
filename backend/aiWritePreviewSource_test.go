package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestAIWritePreviewRejectsMalformedBoundedAndRedirectedSource(t *testing.T) {
	for _, change := range []string{"shape", "duplicate", "range", "string", "oversize", "redirect", "denied", "trashed", "wrongID", "wrongType", "hidden", "checksum", "owner"} {
		t.Run(change, func(t *testing.T) {
			f, image, _, draft, meta := writePreviewFixture(t)
			var followed atomic.Int32
			sink := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { followed.Add(1) }))
			defer sink.Close()
			image.handle = func(w http.ResponseWriter, r *http.Request) bool {
				switch change {
				case "shape":
					meta["exifInfo"] = []any{}
				case "duplicate":
					raw, _ := json.Marshal(meta)
					_, _ = w.Write([]byte(strings.Replace(string(raw), `"exifInfo":{"latitude":null,"longitude":null}`, `"exifInfo":{"latitude":0,"latitude":1}`, 1)))
					return true
				case "range":
					meta["exifInfo"] = map[string]any{"latitude": 91}
				case "string":
					meta["exifInfo"] = map[string]any{"longitude": "0"}
				case "oversize":
					_, _ = w.Write([]byte(strings.Repeat(" ", (1<<20)+1)))
					return true
				case "redirect":
					http.Redirect(w, r, sink.URL, http.StatusFound)
					return true
				case "denied":
					w.WriteHeader(403)
					return true
				case "trashed":
					meta["isTrashed"] = true
				case "wrongID":
					meta["id"] = selectionB
				case "wrongType":
					meta["type"] = "VIDEO"
				case "hidden":
					meta["visibility"] = "locked"
				case "checksum":
					meta["checksum"] = "invalid"
				case "owner":
					meta["ownerId"] = "invalid"
				}
				_ = json.NewEncoder(w).Encode(meta)
				return true
			}
			reads, bad := image.counts()
			body, _ := json.Marshal(map[string]any{"draftId": draft.ID, "draftRevision": draft.Revision})
			rec := aiRequest(writePreviewHandler(f, image), "POST", "/ai/write-previews", string(body), aiTestOrigin, true)
			if rec.Code != 503 || !strings.Contains(rec.Body.String(), "SOURCE_UNAVAILABLE") {
				t.Fatal("unsafe source failure", rec.Code, rec.Body.String())
			}
			var count int
			if err := f.db.db.QueryRow("SELECT count(*) FROM ai_write_previews").Scan(&count); err != nil || count != 0 {
				t.Fatal("failed read published", count, err)
			}
			if after, failures := image.counts(); after != reads+1 || failures != bad || followed.Load() != 0 {
				t.Fatal("read retried, mutated or followed redirect", after, failures, followed.Load())
			}
		})
	}
}
