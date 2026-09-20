package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"immich-places-backend/internal/ai/jobs"
)

type aiJobHeartbeatFault struct{ jobs.Store }

func (s aiJobHeartbeatFault) Heartbeat(context.Context, jobs.Lease, jobs.Policy) error {
	return jobs.ErrStorage
}

func TestAIJobWorkerCancellationAndHeartbeatFailureJoinExecutor(t *testing.T) {
	for _, cause := range []string{"context", "heartbeat-failure", "job-cancel", "closed-heartbeat"} {
		t.Run(cause, func(t *testing.T) {
			f := newAIJobFixture(t)
			base := context.Background()
			job, err := f.store.Submit(base, f.input)
			if err != nil {
				t.Fatal(err)
			}
			completion := aiJobCompletion(t)
			ctx, cancel := context.WithTimeout(base, 2*time.Second)
			defer cancel()
			started, stopped := make(chan struct{}), make(chan struct{})
			ticks := make(chan time.Time, 1)
			var store jobs.Store = f.store
			if cause == "heartbeat-failure" {
				store = aiJobHeartbeatFault{Store: f.store}
			}
			worker := jobs.Worker{Store: store, Policy: jobs.DefaultPolicy(), Execute: func(ctx context.Context, lease jobs.Lease, guard jobs.Guard) (jobs.Completion, error) {
				if err := guard.Reserve(ctx); err != nil {
					return jobs.Completion{}, err
				}
				close(started)
				<-ctx.Done()
				close(stopped)
				return completion, nil
			}}
			done := make(chan error, 1)
			go func() { _, err := worker.RunOne(ctx, ticks); done <- err }()
			select {
			case <-started:
			case err := <-done:
				t.Fatal("executor did not start", err)
			case <-ctx.Done():
				t.Fatal("executor timeout")
			}
			expected := error(context.Canceled)
			switch cause {
			case "context":
				cancel()
			case "heartbeat-failure":
				ticks <- f.now
				expected = jobs.ErrStorage
			case "job-cancel":
				if err = f.store.Cancel(base, testUserID, job.ID); err != nil {
					t.Fatal(err)
				}
				ticks <- f.now
				expected = jobs.ErrDenied
			case "closed-heartbeat":
				close(ticks)
				expected = jobs.ErrLease
			}
			select {
			case err = <-done:
				if !errors.Is(err, expected) {
					t.Fatal("wrong cancellation", cause, err)
				}
			case <-time.After(3 * time.Second):
				t.Fatal("worker abandoned executor")
			}
			select {
			case <-stopped:
			default:
				t.Fatal("worker returned before executor stopped")
			}
			var count int
			if err = f.db.db.QueryRow("SELECT count(*) FROM ai_analyses").Scan(&count); err != nil || count != 0 {
				t.Fatal("late publication", count, err)
			}
		})
	}
}

func TestAIJobWorkerGuardsCannotOutliveAttemptContext(t *testing.T) {
	f := newAIJobFixture(t)
	base := context.Background()
	job, err := f.store.Submit(base, f.input)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(base)
	defer cancel()
	completion := aiJobCompletion(t)
	var authorization, reservation error
	worker := jobs.Worker{Store: f.store, Policy: jobs.DefaultPolicy(), Execute: func(ctx context.Context, lease jobs.Lease, guard jobs.Guard) (jobs.Completion, error) {
		cancel()
		authorization = guard.Authorize(context.Background())
		reservation = guard.Reserve(context.Background())
		return completion, nil
	}}
	if _, err = worker.RunOne(ctx, make(chan time.Time)); !errors.Is(err, context.Canceled) {
		t.Fatal("cancellation lost", err)
	}
	if !errors.Is(authorization, context.Canceled) || !errors.Is(reservation, context.Canceled) {
		t.Fatal("late guard reused authority", authorization, reservation)
	}
	got, err := f.store.Get(base, testUserID, job.ID)
	if err != nil || got.Calls != 0 {
		t.Fatal("post-cancellation reservation", got.Calls, err)
	}
}
