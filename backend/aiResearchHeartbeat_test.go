package main

import (
	"context"
	"errors"
	"immich-places-backend/internal/ai/jobs"
	"testing"
	"time"
)

func TestAIResearchHeartbeatPreservesDeadlineAndCancellationFencesLateOutput(t *testing.T) {
	f, p, req := researchProductionFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	job, err := p.submit(ctx, testUserID, req)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	p.store.now = func() time.Time { return now }
	renewed := make(chan error, 1)
	started := make(chan context.Context, 1)
	ticks := make(chan time.Time, 1)
	store := &aiProductionWorkerStore{aiJobStore: p.store, production: p}
	worker := jobs.Worker{Store: aiJobHeartbeatObserved{Store: store, renewed: renewed}, Policy: jobs.DefaultPolicy(), Execute: func(call context.Context, lease jobs.Lease, guard jobs.Guard) (jobs.Completion, error) {
		if err := guard.Reserve(call); err != nil {
			return jobs.Completion{}, err
		}
		started <- call
		<-call.Done()
		return aiJobCompletion(t), nil
	}}
	done := make(chan error, 1)
	go func() { _, err := worker.RunOne(ctx, ticks); done <- err }()
	var call context.Context
	select {
	case call = <-started:
	case <-ctx.Done():
		t.Fatal("worker did not start")
	}
	deadline, _ := call.Deadline()
	for i := 0; i < 4; i++ {
		now = now.Add(time.Minute)
		ticks <- now
		select {
		case err := <-renewed:
			if err != nil {
				t.Fatal(err)
			}
		case <-ctx.Done():
			t.Fatal("heartbeat stalled")
		}
		current, _ := call.Deadline()
		if !current.Equal(deadline) {
			t.Fatal("heartbeat extended provider deadline")
		}
	}
	if err = p.store.Cancel(ctx, testUserID, job.ID); err != nil {
		t.Fatal(err)
	}
	ticks <- now
	select {
	case err = <-done:
		if !errors.Is(err, jobs.ErrDenied) {
			t.Fatal("cancel not enforced", err)
		}
	case <-ctx.Done():
		t.Fatal("worker did not stop")
	}
	var count int
	if err = f.image.db.db.QueryRow("SELECT count(*) FROM ai_analyses").Scan(&count); err != nil || count != 0 {
		t.Fatal("late result published", err)
	}
	if f.hits.Load() != 0 {
		t.Fatal("heartbeat contacted provider")
	}
}
