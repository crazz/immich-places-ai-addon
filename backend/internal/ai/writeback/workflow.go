package writeback

import (
	"context"
	"immich-places-backend/internal/ai/writepreview"
)

type Completion struct {
	Known bool
	Code  string
}
type Repository interface {
	Reserve(context.Context, Operation, bool) error
	Sent(context.Context, Operation, Completion) error
	Stop(context.Context, Operation, string, string) error
}
type Reader interface {
	Read(context.Context, Operation) (writepreview.Metadata, error)
}
type Sender interface {
	Send(context.Context, Operation) Completion
}
type Publisher interface {
	Verify(context.Context, Operation) error
}
type Executor struct {
	Store       Repository
	Source      Reader
	Mutation    Sender
	Publication Publisher
}

func (e Executor) Execute(ctx context.Context, op Operation) error {
	fresh, err := e.Source.Read(ctx, op)
	if err != nil {
		return e.Store.Stop(ctx, op, "failed", "SOURCE_UNAVAILABLE")
	}
	if fresh.ImageIdentity != op.Plan.ImageIdentity {
		return e.Store.Stop(ctx, op, "conflict", "SOURCE_CHANGED")
	}
	if !writepreview.EqualGPS(fresh.GPS, op.Plan.Before) {
		return e.Store.Stop(ctx, op, "conflict", "IMMICH_CONFLICT")
	}
	noop := writepreview.Diff(op.Plan) == "unchanged"
	if err = e.Store.Reserve(ctx, op, noop); err != nil {
		// Stop is conditional on the still-queued generation. An uncertain commit
		// that already reserved a sender cannot release its guard through this path.
		if stopped := e.Store.Stop(ctx, op, "canceled", "APPROVAL_UNAVAILABLE"); stopped != nil {
			return err
		}
		return nil
	}
	if !noop {
		outcome := e.Mutation.Send(ctx, op)
		if err = e.Store.Sent(ctx, op, outcome); err != nil {
			return err
		}
	}
	return e.Publication.Verify(ctx, op)
}
