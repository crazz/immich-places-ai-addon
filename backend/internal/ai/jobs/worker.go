package jobs

import (
	"context"
	"errors"
	"time"
)

type Store interface {
	Recover(context.Context) (int, error)
	Claim(context.Context, Policy) (Lease, bool, error)
	Authorize(context.Context, Lease) error
	Reserve(context.Context, Lease) error
	Heartbeat(context.Context, Lease, Policy) error
	Complete(context.Context, Lease, Completion) (string, error)
	Fail(context.Context, Lease, Failure, time.Duration) error
}

type Guard struct{ Authorize, Reserve func(context.Context) error }
type Execute func(context.Context, Lease, Guard) (Completion, error)
type Worker struct {
	Store   Store
	Policy  Policy
	Execute Execute
	Jitter  func() time.Duration
}

func (w Worker) RunOne(ctx context.Context, ticks <-chan time.Time) (bool, error) {
	if w.Store == nil || w.Execute == nil || ticks == nil || !w.Policy.Valid() {
		return false, ErrInvalid
	}
	parent := ctx
	ctx, cancel := context.WithTimeout(parent, 120*time.Second)
	defer cancel()
	if _, err := w.Store.Recover(ctx); err != nil {
		return false, err
	}
	lease, claimed, err := w.Store.Claim(ctx, w.Policy)
	if err != nil || !claimed {
		return false, err
	}
	if lease.Input.Mode == "research" {
		cancel()
		duration := w.Policy.ResearchDuration
		if duration == 0 {
			duration = 10 * time.Minute
		}
		ctx, cancel = context.WithTimeout(parent, duration)
		defer cancel()
	}
	guard := Guard{
		Authorize: func(caller context.Context) error {
			if caller.Err() != nil {
				return caller.Err()
			}
			return w.Store.Authorize(ctx, lease)
		},
		Reserve: func(caller context.Context) error {
			if caller.Err() != nil {
				return caller.Err()
			}
			return w.Store.Reserve(ctx, lease)
		},
	}
	done := make(chan struct{})
	monitored := w.monitor(ctx, cancel, lease, ticks, done)
	completion, err := w.Execute(ctx, lease, guard)
	close(done)
	if monitorErr := <-monitored; monitorErr != nil {
		return true, monitorErr
	}
	if ctx.Err() != nil {
		return true, ctx.Err()
	}
	if err != nil {
		return true, w.recordFailure(ctx, lease, err)
	}
	_, err = w.Store.Complete(ctx, lease, completion)
	if errors.Is(err, ErrInvalid) || errors.Is(err, ErrDenied) || errors.Is(err, ErrBudget) {
		return true, w.recordFailure(ctx, lease, err)
	}
	return true, err
}

func (w Worker) recordFailure(ctx context.Context, lease Lease, err error) error {
	if errors.Is(err, ErrStorage) || errors.Is(err, ErrLease) {
		return err
	}
	jitter := time.Duration(0)
	if w.Jitter != nil {
		jitter = w.Jitter()
	}
	return w.Store.Fail(ctx, lease, ExecutionFailure(err), jitter)
}

func (w Worker) monitor(ctx context.Context, cancel context.CancelFunc, lease Lease, ticks <-chan time.Time, done <-chan struct{}) <-chan error {
	result := make(chan error, 1)
	go func() {
		for {
			select {
			case <-done:
				result <- nil
				return
			case <-ctx.Done():
				result <- ctx.Err()
				return
			case _, ok := <-ticks:
				if !ok {
					cancel()
					result <- ErrLease
					return
				}
				if err := w.Store.Heartbeat(ctx, lease, w.Policy); err != nil {
					cancel()
					result <- err
					return
				}
			}
		}
	}()
	return result
}
