package main

import (
	"context"
	"testing"
	"time"

	"immich-places-backend/internal/ai/jobs"
)

type aiJobHeartbeatObserved struct {
	jobs.Store
	renewed chan error
}

func (s aiJobHeartbeatObserved) Heartbeat(ctx context.Context, lease jobs.Lease, policy jobs.Policy) error {
	err := s.Store.Heartbeat(ctx, lease, policy)
	s.renewed <- err
	return err
}

func TestAIJobWorkerRenewsActiveLeaseAndCompletes(t *testing.T) {
	f := newAIJobFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	job, err := f.store.Submit(ctx, f.input)
	if err != nil {
		t.Fatal(err)
	}
	completion := aiJobCompletion(t)
	started, release := make(chan jobs.Lease, 1), make(chan struct{})
	renewed := make(chan error, 1)
	ticks := make(chan time.Time, 1)
	worker := jobs.Worker{Store: aiJobHeartbeatObserved{Store: f.store, renewed: renewed}, Policy: jobs.DefaultPolicy(), Execute: func(ctx context.Context, lease jobs.Lease, guard jobs.Guard) (jobs.Completion, error) {
		if err := guard.Reserve(ctx); err != nil {
			return jobs.Completion{}, err
		}
		started <- lease
		select {
		case <-release:
			return completion, nil
		case <-ctx.Done():
			return jobs.Completion{}, ctx.Err()
		}
	}}
	done := make(chan error, 1)
	go func() { _, err := worker.RunOne(ctx, ticks); done <- err }()
	var lease jobs.Lease
	select {
	case lease = <-started:
	case err = <-done:
		t.Fatal("executor did not start", err)
	case <-ctx.Done():
		t.Fatal("executor timeout")
	}
	f.now = f.now.Add(30 * time.Second)
	ticks <- f.now
	select {
	case err = <-renewed:
		if err != nil {
			t.Fatal("heartbeat failed", err)
		}
	case <-ctx.Done():
		t.Fatal("heartbeat timeout")
	}
	var expiry int64
	err = f.db.db.QueryRowContext(ctx, "SELECT leaseExpiresAt FROM ai_job_items WHERE userID=? AND jobID=? AND id=?", lease.Owner, lease.JobID, lease.ItemID).Scan(&expiry)
	if err != nil || expiry != lease.ExpiresAt.Add(30*time.Second).UnixNano() {
		t.Fatal("lease was not renewed", expiry, err)
	}
	close(release)
	if err = <-done; err != nil {
		t.Fatal("renewed attempt failed", err)
	}
	got, err := f.store.Get(ctx, testUserID, job.ID)
	if err != nil || got.Items[0].State != "succeeded" || got.Calls != 1 {
		t.Fatal("renewed attempt did not complete", got.Items, got.Calls, err)
	}
}
