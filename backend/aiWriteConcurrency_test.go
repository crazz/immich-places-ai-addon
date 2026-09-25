package main

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestAIWriteActiveSenderCannotBeStolenAfterLeaseExpiry(t *testing.T) {
	w := newAIWriteFixture(t)
	entered, release := make(chan struct{}), make(chan struct{})
	w.handle = func(out http.ResponseWriter, r *http.Request) bool {
		if r.Method == "PATCH" {
			close(entered)
			<-release
		}
		return false
	}
	done := make(chan error, 1)
	go func() { done <- w.writer.runOne(context.Background(), testUserID, w.op.ID) }()
	<-entered
	w.f.now = w.f.now.Add(time.Minute)
	_ = w.writer.runOne(context.Background(), testUserID, w.op.ID)
	op := w.status(t)
	close(release)
	firstErr := <-done
	if op.Generation != 1 || op.Status != "writing" {
		t.Fatal("stole running sender", op)
	}
	if firstErr != nil {
		t.Fatal(firstErr)
	}
	if sends, _ := w.counts(); sends != 1 {
		t.Fatal(sends)
	}
}
