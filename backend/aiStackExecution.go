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

type aiStackAttempt struct {
	store             *aiWriteStore
	parent            writeback.Operation
	assetID, token    string
	guardToken        string
	authority         aiImageAuthority
	analyzedAuthority aiImageAuthority
}

func (s *aiWriteStore) runStackTarget(ctx context.Context, parent writeback.Operation) error {
	for _, target := range parent.Targets {
		if target.Status != "queued" && target.Status != "writing" && target.Status != "verifying" {
			continue
		}
		step, err := writeback.TargetStep(parent, target.AssetID)
		if err != nil {
			return err
		}
		a := &aiStackAttempt{store: s, parent: parent, assetID: target.AssetID}
		if target.Status != "queued" {
			claimed, err := a.recover(ctx)
			if err != nil {
				return err
			}
			if claimed {
				return a.Verify(ctx, step)
			}
			continue
		}
		if !s.permitsPlan(parent.Plan) {
			return a.Stop(ctx, step, "canceled", "WRITE_DISABLED")
		}
		expires, err := time.Parse(time.RFC3339Nano, parent.Plan.ExpiresAt)
		if err != nil || !s.drafts.results.jobs.now().Before(expires) {
			return a.Stop(ctx, step, "expired", "APPROVAL_EXPIRED")
		}
		return (writeback.Executor{Store: a, Source: a, Mutation: a, Publication: a}).Execute(ctx, step)
	}
	if parent.Mirror != nil {
		return s.runMirror(ctx, parent)
	}
	return nil
}

func (a *aiStackAttempt) current(ctx context.Context, tx *sql.Tx) (writeback.Operation, error) {
	parent, err := a.store.readStackOperation(ctx, tx, a.parent.Plan.Owner, a.parent.ID, false)
	if err != nil || parent.Digest != a.parent.Digest {
		return writeback.Operation{}, drafts.ErrUnavailable
	}
	return writeback.TargetStep(parent, a.assetID)
}

func (a *aiStackAttempt) Reserve(ctx context.Context, op writeback.Operation, noop bool) error {
	return a.store.drafts.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		current, err := a.current(ctx, tx)
		if err != nil {
			return err
		}
		if current.Status != "queued" || current.Generation != op.Generation || current.Attempts >= 2 {
			return writeback.Failure("WRITE_BUSY")
		}
		if err = a.dispatchAuthority(ctx, tx); err != nil {
			return err
		}
		if tx.QueryRowContext(ctx, `SELECT guardToken FROM ai_stack_write_targets WHERE userID=? AND installationID=? AND operationID=? AND assetID=?`, op.Plan.Owner, op.Plan.Installation, op.ID, a.assetID).Scan(&a.guardToken) != nil {
			return drafts.ErrStorage
		}
		var held string
		if tx.QueryRowContext(ctx, `SELECT token FROM ai_write_target_guards WHERE installationID=? AND assetID=?`, op.Plan.Installation, a.assetID).Scan(&held) != nil || held != a.guardToken {
			return writeback.Failure("TARGET_BUSY")
		}
		a.token = uuid.NewString()
		attempts, state := current.Attempts+1, "writing"
		if noop {
			attempts, state = current.Attempts, "verifying"
		}
		now := a.store.drafts.results.jobs.now()
		if _, err = tx.ExecContext(ctx, `UPDATE ai_stack_write_targets SET attempts=?,generation=generation+1,leaseToken=?,leaseUntil=?,completionKnown=?,senderActive=?,noop=?,dueAt=?,status=? WHERE userID=? AND installationID=? AND operationID=? AND assetID=?`, attempts, a.token, now.Add(45*time.Second).UnixNano(), noop, !noop, noop, now.Add(time.Second).UnixNano(), state, op.Plan.Owner, op.Plan.Installation, op.ID, a.assetID); err != nil {
			return drafts.ErrStorage
		}
		return a.event(ctx, tx, "reserved", attempts)
	})
}

func (a *aiStackAttempt) Send(ctx context.Context, op writeback.Operation) writeback.Completion {
	err := a.store.drafts.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if err := a.fence(ctx, tx); err != nil {
			return err
		}
		var leaseUntil int64
		if tx.QueryRowContext(ctx, `SELECT leaseUntil FROM ai_stack_write_targets WHERE userID=? AND installationID=? AND operationID=? AND assetID=?`, op.Plan.Owner, op.Plan.Installation, op.ID, a.assetID).Scan(&leaseUntil) != nil || leaseUntil <= a.store.drafts.results.jobs.now().UnixNano() {
			return writeback.Failure("STALE_WORKER")
		}
		return a.dispatchAuthority(ctx, tx)
	})
	if err != nil {
		return writeback.Completion{Known: true, Code: "not_sent"}
	}
	transport := &immichwrite.Transport{Endpoint: a.store.images.endpoint}
	var outcome immichwrite.Outcome
	if op.Plan.Description != nil {
		outcome = transport.SendStandard(ctx, a.authority.key, op.Plan)
	} else {
		outcome = transport.Send(ctx, a.authority.key, a.assetID, op.Plan.Intended)
	}
	return writeback.Completion{Known: outcome.CompletionKnown, Code: outcome.Code}
}

func (a *aiStackAttempt) Sent(ctx context.Context, op writeback.Operation, result writeback.Completion) error {
	err := a.store.drafts.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if err := a.fence(ctx, tx); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE ai_stack_write_targets SET completionKnown=?,senderActive=0,status='verifying',code=?,dueAt=? WHERE userID=? AND installationID=? AND operationID=? AND assetID=?`, result.Known, result.Code, a.store.drafts.results.jobs.now().Add(time.Second).UnixNano(), op.Plan.Owner, op.Plan.Installation, op.ID, a.assetID); err != nil {
			return drafts.ErrStorage
		}
		current, err := a.current(ctx, tx)
		if err != nil {
			return err
		}
		return a.event(ctx, tx, result.Code, current.Attempts)
	})
	if err != nil && result.Known {
		cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
		defer cancel()
		_ = a.store.drafts.write(cleanup, func(ctx context.Context, tx *sql.Tx) error {
			_, err := tx.ExecContext(ctx, `DELETE FROM ai_write_target_guards WHERE installationID=? AND assetID=? AND token=? AND NOT EXISTS(SELECT 1 FROM users WHERE ID=?)`, op.Plan.Installation, a.assetID, a.guardToken, op.Plan.Owner)
			return err
		})
	}
	return err
}

func (a *aiStackAttempt) fence(ctx context.Context, tx *sql.Tx) error {
	var token string
	if a.token == "" || tx.QueryRowContext(ctx, `SELECT leaseToken FROM ai_stack_write_targets WHERE userID=? AND installationID=? AND operationID=? AND assetID=?`, a.parent.Plan.Owner, a.parent.Plan.Installation, a.parent.ID, a.assetID).Scan(&token) != nil || token != a.token {
		return writeback.Failure("STALE_WORKER")
	}
	return nil
}

func (a *aiStackAttempt) Stop(ctx context.Context, op writeback.Operation, status, code string) error {
	return a.store.drafts.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		result, err := tx.ExecContext(ctx, `UPDATE ai_stack_write_targets SET status=?,code=? WHERE userID=? AND installationID=? AND operationID=? AND assetID=? AND status='queued' AND generation=?`, status, code, op.Plan.Owner, op.Plan.Installation, op.ID, a.assetID, op.Generation)
		if err != nil {
			return drafts.ErrStorage
		}
		if n, err := result.RowsAffected(); err != nil || n != 1 {
			return writeback.Failure("STALE_WORKER")
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM ai_write_target_guards WHERE token=(SELECT guardToken FROM ai_stack_write_targets WHERE userID=? AND installationID=? AND operationID=? AND assetID=?)`, op.Plan.Owner, op.Plan.Installation, op.ID, a.assetID); err != nil {
			return drafts.ErrStorage
		}
		return a.event(ctx, tx, code, op.Attempts)
	})
}

func (a *aiStackAttempt) event(ctx context.Context, tx *sql.Tx, code string, attempts int) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO ai_stack_write_events(userID,installationID,operationID,assetID,code,at,attempt) SELECT ?,?,?,?,?,?,? WHERE (SELECT count(*) FROM ai_stack_write_events WHERE userID=? AND installationID=? AND operationID=? AND assetID=?)<100`, a.parent.Plan.Owner, a.parent.Plan.Installation, a.parent.ID, a.assetID, code, a.store.drafts.results.jobs.now().UnixNano(), attempts, a.parent.Plan.Owner, a.parent.Plan.Installation, a.parent.ID, a.assetID)
	return err
}
