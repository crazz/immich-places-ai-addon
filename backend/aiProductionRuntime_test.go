package main

import (
	"context"
	"net/http"
	"testing"
	"time"

	"immich-places-backend/internal/ai/jobs"
)

func TestAIProductionRuntimeJoinsShutdownAndRecoversAcceptedWork(t *testing.T) {
	f, p, req := productionFixture(t)
	cfg := &Config{AIEnabled: true, ImmichURL: f.image.server.URL, AIExecutionPolicies: p.policies, AIProviderEgressPolicy: f.analyzer.dispatcher.Policy, AIJobSettings: jobs.ConsumerSettings{Policy: jobs.DefaultPolicy(), Heartbeat: 30 * time.Second, Idle: 100 * time.Millisecond}}
	runtime, err := newAIProductionRuntime(f.image.db, cfg, p.selections, f.analyzer.dispatcher)
	if err != nil || runtime.jobs == nil {
		t.Fatal("production runtime unavailable", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	entered := make(chan struct{})
	returned := make(chan struct{})
	f.handle = func(w http.ResponseWriter, r *http.Request) bool {
		close(entered)
		<-r.Context().Done()
		close(returned)
		return true
	}
	job, err := runtime.jobs.submit(context.Background(), testUserID, req)
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- runtime.run(ctx, func() {}) }()
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		t.Fatal("consumer did not execute accepted job")
	}
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("worker shutdown was not joined")
	}
	select {
	case <-returned:
	case <-time.After(time.Second):
		t.Fatal("upstream request not canceled")
	}
	progress, err := runtime.jobs.progress(context.Background(), testUserID, job.ID)
	if err != nil || progress.Items[0].ResultID != nil || progress.Usage.Calls != 1 {
		t.Fatal("shutdown published or refunded", progress, err)
	}
	f.handle = nil
	restarted, err := newAIProductionRuntime(f.image.db, cfg, p.selections, f.analyzer.dispatcher)
	if err != nil {
		t.Fatal(err)
	}
	future := time.Now().Add(4 * time.Minute)
	restarted.jobs.store.now = func() time.Time { return future }
	// Recovery consumes the attempt, so this one-call job terminates without another dispatch.
	ctx2, cancel2 := context.WithCancel(context.Background())
	defer cancel2()
	done2 := make(chan error, 1)
	go func() { done2 <- restarted.run(ctx2, func() {}) }()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		progress, err = restarted.jobs.progress(context.Background(), testUserID, job.ID)
		if err == nil && progress.Counts["failed"] == 1 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	cancel2()
	if err := <-done2; err != nil {
		t.Fatal(err)
	}
	if progress.Counts["failed"] != 1 || f.hits.Load() != 1 || progress.Usage.ReservedTokens != 104000 {
		t.Fatal("restart retried exhausted allowance", progress)
	}
}
