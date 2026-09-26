package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

type aiStackMutationFixture struct {
	mu     sync.Mutex
	counts map[string]int
	handle func(http.ResponseWriter, *http.Request) bool
}

func stackMutationServer(t *testing.T, f *aiStackFixture) *aiStackMutationFixture {
	t.Helper()
	m := &aiStackMutationFixture{counts: map[string]int{}}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") != "private-immich-image-key" {
			w.WriteHeader(403)
			return
		}
		id := strings.TrimPrefix(r.URL.Path, "/api/assets/")
		if r.Method == "PATCH" {
			m.mu.Lock()
			m.counts[id]++
			m.mu.Unlock()
		}
		if m.handle != nil && m.handle(w, r) {
			return
		}
		if r.Method == "GET" && f.image.handle(w, r) {
			return
		}
		f.mu.Lock()
		defer f.mu.Unlock()
		meta, exists := f.metadata[id]
		if !exists || r.Method != "PATCH" || r.URL.Path != "/api/assets/"+id {
			t.Error("unexpected fixture request")
			http.NotFound(w, r)
			return
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Error(err)
			w.WriteHeader(400)
			return
		}
		for key, value := range payload {
			meta["exifInfo"].(map[string]any)[key] = value
		}
		w.WriteHeader(200)
	}))
	t.Cleanup(server.Close)
	f.image.service.endpoint = server.URL
	return m
}

func (m *aiStackMutationFixture) sent(id string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.counts[id]
}
