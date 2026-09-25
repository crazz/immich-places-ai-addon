package main

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writeback"
	"immich-places-backend/internal/aiadapters/immichwrite"
)

func (s *aiWriteStore) runOne(ctx context.Context, owner, id string) error {
	op, err := s.get(ctx, owner, id, false)
	if err != nil {
		return err
	}
	if !s.enter(op) {
		return nil
	}
	defer s.leave(op)
	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	a := &aiWriteAttempt{store: s}
	if op.Status == "queued" && !s.available() {
		return a.Stop(ctx, op, "canceled", "WRITE_DISABLED")
	}
	if op.Status == "writing" || op.Status == "verifying" {
		ok, err := a.recover(ctx, op)
		if err != nil || !ok {
			return err
		}
		return a.Verify(ctx, op)
	}
	if op.Status != "queued" {
		return nil
	}
	return (writeback.Executor{Store: a, Source: a, Mutation: a, Publication: a}).Execute(ctx, op)
}

func (a *aiWriteAttempt) Reserve(ctx context.Context, op writeback.Operation, noop bool) error {
	s := a.store
	return s.drafts.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		current, err := s.read(ctx, tx, op.Plan.Owner, op.ID, false)
		if err != nil {
			return err
		}
		if current.Status != "queued" || current.Generation != op.Generation {
			return writeback.Failure("WRITE_BUSY")
		}
		if err = s.dispatchAuthority(ctx, tx, current, a.authority); err != nil {
			return err
		}
		a.token = uuid.NewString()
		attempts, state := current.Attempts+1, "writing"
		if noop {
			attempts, state = current.Attempts, "verifying"
		}
		if _, err = tx.ExecContext(ctx, `UPDATE ai_write_targets SET attempts=?,generation=generation+1,leaseToken=?,leaseUntil=?,completionKnown=?,senderActive=?,noop=?,dueAt=? WHERE userID=? AND installationID=? AND operationID=?`, attempts, a.token, s.drafts.results.jobs.now().Add(45*time.Second).UnixNano(), noop, !noop, noop, s.drafts.results.jobs.now().Add(time.Second).UnixNano(), op.Plan.Owner, op.Plan.Installation, op.ID); err != nil {
			return drafts.ErrStorage
		}
		if _, err = tx.ExecContext(ctx, `UPDATE ai_write_operations SET status=? WHERE userID=? AND installationID=? AND id=?`, state, op.Plan.Owner, op.Plan.Installation, op.ID); err != nil {
			return drafts.ErrStorage
		}
		current.Attempts = attempts
		return s.event(ctx, tx, current, "reserved")
	})
}

func (s *aiWriteStore) dispatchAuthority(ctx context.Context, tx *sql.Tx, op writeback.Operation, authority aiImageAuthority) error {
	if !s.available() {
		return writeback.Failure("WRITE_DISABLED")
	}
	expires, err := time.Parse(time.RFC3339Nano, op.Plan.ExpiresAt)
	if err != nil || !s.drafts.results.jobs.now().Before(expires) {
		return writeback.Failure("APPROVAL_EXPIRED")
	}
	current, err := s.drafts.read(ctx, tx, op.Plan.Owner, op.Plan.DraftID)
	if err != nil {
		return err
	}
	if current.Revision != op.Plan.DraftRevision || current.State != "staged" {
		return writeback.Failure("DRAFT_CONFLICT")
	}
	if err = s.drafts.checkAuthority(ctx, tx, op.Plan.Owner, op.Plan.TargetID, authority); err != nil {
		return err
	}
	var hash string
	if tx.QueryRowContext(ctx, `SELECT credentialHash FROM ai_write_operations WHERE userID=? AND installationID=? AND id=?`, op.Plan.Owner, op.Plan.Installation, op.ID).Scan(&hash) != nil || hash != aiWriteCredentialHash(authority.key) {
		return drafts.ErrUnavailable
	}
	return ctx.Err()
}

func (a *aiWriteAttempt) Send(ctx context.Context, op writeback.Operation) writeback.Completion {
	err := a.store.drafts.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if err := a.fence(ctx, tx, op); err != nil {
			return err
		}
		var leaseUntil int64
		if tx.QueryRowContext(ctx, `SELECT leaseUntil FROM ai_write_targets WHERE userID=? AND installationID=? AND operationID=?`, op.Plan.Owner, op.Plan.Installation, op.ID).Scan(&leaseUntil) != nil || leaseUntil <= a.store.drafts.results.jobs.now().UnixNano() {
			return writeback.Failure("STALE_WORKER")
		}
		return a.store.dispatchAuthority(ctx, tx, op, a.authority)
	})
	if err != nil {
		return writeback.Completion{Known: true, Code: "not_sent"}
	}
	result := (&immichwrite.Transport{Endpoint: a.store.images.endpoint}).Send(ctx, a.authority.key, op.Plan.TargetID, op.Plan.Intended)
	return writeback.Completion{Known: result.CompletionKnown, Code: result.Code}
}

func (a *aiWriteAttempt) fence(ctx context.Context, tx *sql.Tx, op writeback.Operation) error {
	var token string
	if tx.QueryRowContext(ctx, `SELECT leaseToken FROM ai_write_targets WHERE userID=? AND installationID=? AND operationID=?`, op.Plan.Owner, op.Plan.Installation, op.ID).Scan(&token) != nil || token != a.token {
		return writeback.Failure("STALE_WORKER")
	}
	return nil
}

func (a *aiWriteAttempt) Sent(ctx context.Context, op writeback.Operation, outcome writeback.Completion) error {
	err := a.store.drafts.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if err := a.fence(ctx, tx, op); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE ai_write_targets SET completionKnown=?,senderActive=0,dueAt=? WHERE userID=? AND installationID=? AND operationID=?`, outcome.Known, a.store.drafts.results.jobs.now().Add(time.Second).UnixNano(), op.Plan.Owner, op.Plan.Installation, op.ID); err != nil {
			return drafts.ErrStorage
		}
		if _, err := tx.ExecContext(ctx, `UPDATE ai_write_operations SET status='verifying',code=? WHERE userID=? AND installationID=? AND id=?`, outcome.Code, op.Plan.Owner, op.Plan.Installation, op.ID); err != nil {
			return drafts.ErrStorage
		}
		current, err := a.store.read(ctx, tx, op.Plan.Owner, op.ID, false)
		if err != nil {
			return err
		}
		return a.store.event(ctx, tx, current, outcome.Code)
	})
	if err != nil && outcome.Known {
		cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
		defer cancel()
		// Deletion removed the private target row. Retire only this opaque guard
		// after a definite sender completion, without recreating any private data.
		_ = a.store.drafts.write(cleanup, func(ctx context.Context, tx *sql.Tx) error {
			_, cleanupErr := tx.ExecContext(ctx, `DELETE FROM ai_write_target_guards WHERE installationID=? AND assetID=? AND token=? AND NOT EXISTS(SELECT 1 FROM users WHERE ID=?)`, op.Plan.Installation, op.Plan.TargetID, op.ID, op.Plan.Owner)
			return cleanupErr
		})
	}
	return err
}

func (a *aiWriteAttempt) Stop(ctx context.Context, op writeback.Operation, status, code string) error {
	return a.store.drafts.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		result, err := tx.ExecContext(ctx, `UPDATE ai_write_operations SET status=?,code=? WHERE userID=? AND installationID=? AND id=? AND status='queued' AND EXISTS(SELECT 1 FROM ai_write_targets t WHERE t.userID=ai_write_operations.userID AND t.installationID=ai_write_operations.installationID AND t.operationID=ai_write_operations.id AND t.generation=?)`, status, code, op.Plan.Owner, op.Plan.Installation, op.ID, op.Generation)
		if err != nil {
			return drafts.ErrStorage
		}
		n, err := result.RowsAffected()
		if err != nil || n != 1 {
			return writeback.Failure("STALE_WORKER")
		}
		if _, err = tx.ExecContext(ctx, `DELETE FROM ai_write_target_guards WHERE token=?`, op.ID); err != nil {
			return drafts.ErrStorage
		}
		return a.store.event(ctx, tx, op, code)
	})
}
