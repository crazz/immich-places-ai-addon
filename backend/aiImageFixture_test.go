package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/color"
	"image/jpeg"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

type aiImageFixture struct {
	db          *Database
	service     *aiImagePreparer
	store       *aiSelectionStore
	server      *httptest.Server
	image       []byte
	mu          sync.Mutex
	paths       []string
	badRequests int
	handle      func(http.ResponseWriter, *http.Request) bool
}

func newAIImageFixture(t *testing.T) *aiImageFixture {
	t.Helper()
	f := &aiImageFixture{db: newTestDB(t)}
	seedAsset(t, f.db, selectionA, nil, nil, "2026-09-20")
	key := "private-immich-image-key"
	if err := f.db.updateImmichAPIKey(context.Background(), testUserID, &key); err != nil {
		t.Fatal(err)
	}
	img := image.NewNRGBA(image.Rect(0, 0, 8, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 8; x++ {
			img.SetNRGBA(x, y, color.NRGBA{R: 180, G: 80, A: 255})
		}
	}
	var b bytes.Buffer
	if err := jpeg.Encode(&b, img, nil); err != nil {
		t.Fatal(err)
	}
	f.image = b.Bytes()
	f.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		f.paths = append(f.paths, r.URL.RequestURI())
		valid := r.Method == "GET" && r.Header.Get("x-api-key") == key
		if !valid {
			f.badRequests++
		}
		f.mu.Unlock()
		if !valid {
			http.Error(w, "private-immich-image-key", 401)
			return
		}
		if f.handle != nil && f.handle(w, r) {
			return
		}
		switch r.URL.RequestURI() {
		case "/api/assets/" + selectionA:
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(f.metadata()); err != nil {
				t.Error(err)
			}
		case "/api/assets/" + selectionA + "/thumbnail?size=preview":
			w.Header().Set("Content-Type", "image/jpeg")
			if _, err := w.Write(f.image); err != nil {
				t.Error(err)
			}
		default:
			f.mu.Lock()
			f.badRequests++
			f.mu.Unlock()
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(f.server.Close)
	f.store = newAISelectionStore(f.db)
	f.store.enabled = true
	if err := f.store.bind(context.Background(), f.server.URL, "1"); err != nil {
		t.Fatal(err)
	}
	var err error
	f.service, err = newAIImagePreparer(f.db, f.store, f.server.URL)
	if err != nil {
		t.Fatal(err)
	}
	return f
}
func (f *aiImageFixture) metadata() map[string]any {
	return map[string]any{"id": selectionA, "ownerId": selectionID(99), "type": "IMAGE", "visibility": "timeline", "isTrashed": false, "checksum": base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{1}, 20)), "updatedAt": "2026-09-20T12:00:00Z", "stack": nil}
}
func (f *aiImageFixture) counts() (int, int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.paths), f.badRequests
}
