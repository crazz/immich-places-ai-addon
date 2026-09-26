package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"immich-places-backend/internal/ai/writepreview"
)

func mirrorMutationServer(t *testing.T, f *aiMirrorFixture) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") != "private-immich-image-key" {
			w.WriteHeader(403)
			return
		}
		if r.Method == "GET" {
			if f.image.handle(w, r) {
				return
			}
			http.NotFound(w, r)
			return
		}
		if f.handle != nil && f.handle(w, r) {
			return
		}
		f.mu.Lock()
		defer f.mu.Unlock()
		switch {
		case r.Method == "PATCH" && r.URL.Path == "/api/assets/"+f.draft.AssetID:
			var payload map[string]any
			if json.NewDecoder(r.Body).Decode(&payload) != nil {
				t.Error("invalid standard payload")
				w.WriteHeader(400)
				return
			}
			for key, value := range payload {
				f.meta["exifInfo"].(map[string]any)[key] = value
			}
			f.standardSends++
		case r.Method == "PUT" && r.URL.Path == "/api/assets/"+f.draft.AssetID+"/metadata":
			var body struct {
				Items []struct {
					Key   string          `json:"key"`
					Value json.RawMessage `json:"value"`
				} `json:"items"`
			}
			if json.NewDecoder(r.Body).Decode(&body) != nil || len(body.Items) != 1 || body.Items[0].Key != writepreview.MirrorNamespace {
				t.Error("expanded mirror mutation")
				w.WriteHeader(400)
				return
			}
			f.namespace = body.Items[0].Value
			f.mirrorSends++
		default:
			t.Error("unexpected write target or method")
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(200)
	}))
	t.Cleanup(server.Close)
	f.image.service.endpoint = server.URL
}
