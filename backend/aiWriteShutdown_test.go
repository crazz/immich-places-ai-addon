package main

import (
	"context"
	"net/http"
	"path/filepath"
	"testing"
	"time"
)

func TestAIWriteShutdownRetainsPotentiallySentReservation(t *testing.T) {
	w := newAIWriteFixture(t)
	entered, release := make(chan struct{}), make(chan struct{})
	w.handle = func(out http.ResponseWriter, r *http.Request) bool {
		if r.Method != "PATCH" {
			return false
		}
		close(entered)
		<-release
		return true
	}
	lock, err := acquireAIWriteLock(filepath.Join(t.TempDir(), "shutdown.lock"))
	if err != nil {
		t.Fatal(err)
	}
	runtime := &aiWriteRuntime{store: w.writer, lock: lock}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() { runtime.run(ctx); close(done) }()
	<-entered
	cancel()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		close(release)
		t.Fatal("shutdown did not cancel sender")
	}
	close(release)
	op := w.status(t)
	if op.Attempts != 1 || op.Verified || (op.Status != "writing" && op.Status != "verifying") {
		t.Fatal("false shutdown outcome", op)
	}
	if sends, _ := w.counts(); sends != 1 {
		t.Fatal(sends)
	}
	var guards int
	if err = w.f.db.db.QueryRow("SELECT count(*) FROM ai_write_target_guards").Scan(&guards); err != nil || guards != 1 {
		t.Fatal(guards, err)
	}
}
