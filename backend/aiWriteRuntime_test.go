package main

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestAIWriteRuntimeSweepsDurableQueueAndStops(t *testing.T) {
	w := newAIWriteFixture(t)
	lock, err := acquireAIWriteLock(filepath.Join(t.TempDir(), "runtime.lock"))
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Close()
	runtime := &aiWriteRuntime{store: w.writer, lock: lock}
	if err = runtime.sweep(context.Background()); err != nil {
		t.Fatal(err)
	}
	if op := w.status(t); op.Status != "succeeded" {
		t.Fatal(op)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	done := make(chan struct{})
	go func() { runtime.run(ctx); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("unbounded shutdown")
	}
}
