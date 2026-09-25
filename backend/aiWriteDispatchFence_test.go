package main

import (
	"context"
	"net/http"
	"testing"
)

func TestAIWriteDisableDuringPreflightRetiresQueuedApproval(t *testing.T) {
	w := newAIWriteFixture(t)
	w.handle = func(_ http.ResponseWriter, r *http.Request) bool {
		if r.Method == "GET" {
			w.mu.Lock()
			w.enabled = false
			w.mu.Unlock()
		}
		return false
	}
	_ = w.writer.runOne(context.Background(), testUserID, w.op.ID)
	if op := w.status(t); op.Status != "canceled" {
		t.Fatal("approval survived disable", op.Status)
	}
	w.handle = nil
	w.enabled = true
	w.run(t)
	if sends, _ := w.counts(); sends != 0 {
		t.Fatal("stale approval dispatched", sends)
	}
}
