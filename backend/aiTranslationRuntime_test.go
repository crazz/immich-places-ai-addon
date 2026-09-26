package main

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestAITranslationRuntimeJoinsShutdownWithoutPublishing(t *testing.T) {
	f, s, req := translationFixture(t)
	entered, returned := make(chan struct{}), make(chan struct{})
	dispatcher := translationProvider(t, f, s, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		close(entered)
		select {
		case <-r.Context().Done():
			close(returned)
		case <-time.After(5 * time.Second):
		}
	})
	req.ProfileRevision = 2
	req.Languages = []string{"en"}
	if _, err := s.submit(context.Background(), testUserID, req); err != nil {
		t.Fatal(err)
	}
	runtime := &aiTranslationRuntime{store: s, dispatcher: dispatcher}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- runtime.run(ctx, func() {}) }()
	select {
	case <-entered:
	case err := <-done:
		t.Fatal("runtime did not execute", err)
	case <-time.After(3 * time.Second):
		t.Fatal("runtime did not start")
	}
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("shutdown did not join sender")
	}
	select {
	case <-returned:
	case <-time.After(time.Second):
		t.Fatal("upstream context not canceled")
	}
	run, err := s.submit(context.Background(), testUserID, req)
	if err != nil || run.Items[0].Text != nil || run.Items[0].State != "interrupted" {
		t.Fatal("shutdown outcome not retained", run, err)
	}
}

func TestAITranslationRuntimeDisabledRetainsQueuedHistory(t *testing.T) {
	f, s, req := translationFixture(t)
	run, err := s.submit(context.Background(), testUserID, req)
	if err != nil {
		t.Fatal(err)
	}
	f.store.enabled = false
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if err := (&aiTranslationRuntime{store: s}).run(ctx, func() {}); err != nil {
		t.Fatal(err)
	}
	got, err := s.submit(context.Background(), testUserID, req)
	if err != nil || got.ID != run.ID || got.Items[0].State != "queued" {
		t.Fatal("disabled runtime altered queued history", got, err)
	}
}
