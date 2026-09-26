package translations

import (
	"context"
	"errors"
	"immich-places-backend/internal/ai/analysis"
	"immich-places-backend/internal/ai/jobs"
	"time"
)

type Task struct {
	Owner, ID, Language, PolicyID string
	Request                       Request
	Policy                        jobs.ExecutionPolicy
}

type WorkerStore interface {
	Claim(context.Context) (Task, bool, error)
	Authorize(context.Context, Task) error
	Reserve(context.Context, Task) error
	RecordUsage(context.Context, Task, *analysis.Usage) error
	Finish(context.Context, Task, Outcome, string) error
}

type Provider interface {
	Translate(context.Context, Task, func(context.Context) error) (Outcome, *analysis.Usage, error)
}

type Failure string

func (f Failure) Error() string { return string(f) }

type Worker struct {
	Store    WorkerStore
	Provider Provider
}

func (w Worker) RunOne(ctx context.Context) (bool, error) {
	task, claimed, err := w.Store.Claim(ctx)
	if err != nil || !claimed {
		return false, err
	}
	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()
	finish := func(out Outcome, failure string) (bool, error) { return true, w.Store.Finish(ctx, task, out, failure) }
	failed := func(code string) (bool, error) {
		return finish(Outcome{Language: task.Language, Status: "failed"}, code)
	}
	authorize := func(ctx context.Context) error { return w.Store.Authorize(ctx, task) }
	if err = authorize(ctx); err != nil {
		return failed("authority_changed")
	}
	if err = w.Store.Reserve(ctx, task); err != nil {
		if errors.Is(err, jobs.ErrBudget) {
			return failed("budget")
		}
		return failed("storage")
	}
	out, usage, err := w.Provider.Translate(ctx, task, authorize)
	if recordErr := w.Store.RecordUsage(ctx, task, usage); recordErr != nil {
		return failed("storage")
	}
	if err != nil {
		var failure Failure
		if errors.As(err, &failure) {
			return failed(string(failure))
		}
		return failed("provider_failed")
	}
	return finish(out, "")
}
