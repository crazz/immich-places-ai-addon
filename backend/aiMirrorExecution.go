package main

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writeback"
	"immich-places-backend/internal/ai/writepreview"
	"immich-places-backend/internal/aiadapters/immichwrite"
)

type aiMirrorAttempt struct {
	standard          *aiStackAttempt
	token, guardToken string
}

func (s *aiWriteStore) runMirror(ctx context.Context, parent writeback.Operation) error {
	if parent.Mirror == nil {
		return nil
	}
	a := &aiMirrorAttempt{standard: &aiStackAttempt{store: s, parent: parent, assetID: parent.Plan.TargetID}}
	if parent.Mirror.Status == "writing" || parent.Mirror.Status == "verifying" {
		claimed, err := a.recover(ctx)
		if err != nil || !claimed {
			return err
		}
		return a.verify(ctx)
	}
	if (parent.Mirror.Status != "blocked" && parent.Mirror.Status != "queued") || !writeback.MirrorStandardReady(parent) {
		return nil
	}
	if !s.permitsPlan(parent.Plan) {
		return a.stop(ctx, "canceled", "WRITE_DISABLED")
	}
	expires, err := time.Parse(time.RFC3339Nano, parent.Plan.ExpiresAt)
	if err != nil || !s.drafts.results.jobs.now().Before(expires) {
		return a.stop(ctx, "expired", "APPROVAL_EXPIRED")
	}
	fresh, err := a.read(ctx, false)
	if err != nil {
		return a.stop(ctx, "failed", "METADATA_UNAVAILABLE")
	}
	step, err := writeback.TargetStep(parent, parent.Plan.TargetID)
	if err != nil {
		return err
	}
	if !writeback.Readback(step, fresh, true).Verified {
		return a.stop(ctx, "conflict", "STANDARD_FIELDS_CHANGED")
	}
	if fresh.ImageIdentity != parent.Plan.ImageIdentity {
		return a.stop(ctx, "conflict", "SOURCE_CHANGED")
	}
	if fresh.Mirror == nil || !writeback.EqualMirrorBaseline(*fresh.Mirror, parent.Plan.Mirror.Before) {
		return a.stop(ctx, "conflict", "METADATA_CONFLICT")
	}
	noop := fresh.Mirror.Present && writepreview.EqualMirrorValue(fresh.Mirror.Value, parent.Plan.Mirror.Value)
	if err = a.reserve(ctx, noop); err != nil {
		if stopped := a.stop(ctx, "canceled", "APPROVAL_UNAVAILABLE"); stopped != nil {
			return err
		}
		return nil
	}
	if !noop {
		outcome := a.send(ctx)
		if err = a.sent(ctx, outcome); err != nil {
			return err
		}
	}
	return a.verify(ctx)
}

func (a *aiMirrorAttempt) reserve(ctx context.Context, noop bool) error {
	s, p := a.standard.store, a.standard.parent
	return s.drafts.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		current, err := s.readStackOperation(ctx, tx, p.Plan.Owner, p.ID, false)
		if err != nil {
			return err
		}
		m := current.Mirror
		if m == nil || !writeback.MirrorStandardReady(current) || (m.Status != "blocked" && m.Status != "queued") || m.Generation != p.Mirror.Generation || m.Attempts >= 2 {
			return writeback.Failure("WRITE_BUSY")
		}
		if err = a.standard.dispatchAuthority(ctx, tx); err != nil {
			return err
		}
		var record string
		if tx.QueryRowContext(ctx, `SELECT recordID FROM ai_mirror_records WHERE userID=? AND installationID=? AND assetID=?`, p.Plan.Owner, p.Plan.Installation, p.Plan.TargetID).Scan(&record) != nil || record != p.Plan.Mirror.RecordID {
			return drafts.ErrStorage
		}
		if tx.QueryRowContext(ctx, `SELECT t.guardToken FROM ai_stack_write_targets t JOIN ai_write_target_guards g ON g.installationID=t.installationID AND g.assetID=t.assetID AND g.token=t.guardToken WHERE t.userID=? AND t.installationID=? AND t.operationID=? AND t.assetID=?`, p.Plan.Owner, p.Plan.Installation, p.ID, p.Plan.TargetID).Scan(&a.guardToken) != nil {
			return writeback.Failure("TARGET_BUSY")
		}
		a.token = uuid.NewString()
		attempts, status := m.Attempts+1, "writing"
		if noop {
			attempts, status = m.Attempts, "verifying"
		}
		now := s.drafts.results.jobs.now()
		_, err = tx.ExecContext(ctx, `UPDATE ai_mirror_write_steps SET status=?,attempts=?,generation=generation+1,leaseToken=?,leaseUntil=?,completionKnown=?,senderActive=?,noop=?,reads=0,dueAt=?,updatedAt=? WHERE userID=? AND installationID=? AND operationID=? AND assetID=?`, status, attempts, a.token, now.Add(45*time.Second).UnixNano(), noop, !noop, noop, now.Add(time.Second).UnixNano(), now.UnixNano(), p.Plan.Owner, p.Plan.Installation, p.ID, p.Plan.TargetID)
		if err != nil {
			return drafts.ErrStorage
		}
		return a.event(ctx, tx, "reserved")
	})
}

func (a *aiMirrorAttempt) send(ctx context.Context) writeback.Completion {
	p, s := a.standard.parent, a.standard.store
	fresh, err := a.read(ctx, false)
	step, stepErr := writeback.TargetStep(p, p.Plan.TargetID)
	if err != nil || stepErr != nil || !writeback.Readback(step, fresh, true).Verified || fresh.ImageIdentity != p.Plan.ImageIdentity || fresh.Mirror == nil || !writeback.EqualMirrorBaseline(*fresh.Mirror, p.Plan.Mirror.Before) {
		return writeback.Completion{Known: true, Code: "not_sent"}
	}
	err = s.drafts.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if err := a.fence(ctx, tx); err != nil {
			return err
		}
		var until int64
		if tx.QueryRowContext(ctx, `SELECT leaseUntil FROM ai_mirror_write_steps WHERE userID=? AND installationID=? AND operationID=? AND assetID=?`, p.Plan.Owner, p.Plan.Installation, p.ID, p.Plan.TargetID).Scan(&until) != nil || until <= s.drafts.results.jobs.now().UnixNano() {
			return writeback.Failure("STALE_WORKER")
		}
		return a.standard.dispatchAuthority(ctx, tx)
	})
	if err != nil {
		return writeback.Completion{Known: true, Code: "not_sent"}
	}
	out := (&immichwrite.Transport{Endpoint: s.images.endpoint}).SendMetadata(ctx, a.standard.authority.key, p.Plan)
	return writeback.Completion{Known: out.CompletionKnown, Code: out.Code}
}

func (a *aiMirrorAttempt) sent(ctx context.Context, outcome writeback.Completion) error {
	p, s := a.standard.parent, a.standard.store
	err := s.drafts.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if err := a.fence(ctx, tx); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx, `UPDATE ai_mirror_write_steps SET status='verifying',code=?,completionKnown=?,senderActive=0,dueAt=? WHERE userID=? AND installationID=? AND operationID=? AND assetID=?`, outcome.Code, outcome.Known, s.drafts.results.jobs.now().Add(time.Second).UnixNano(), p.Plan.Owner, p.Plan.Installation, p.ID, p.Plan.TargetID)
		if err != nil {
			return drafts.ErrStorage
		}
		return a.event(ctx, tx, outcome.Code)
	})
	if err != nil && outcome.Known {
		cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
		defer cancel()
		_ = s.drafts.write(cleanup, func(ctx context.Context, tx *sql.Tx) error {
			_, cleanupErr := tx.ExecContext(ctx, `DELETE FROM ai_write_target_guards WHERE installationID=? AND assetID=? AND token=? AND NOT EXISTS(SELECT 1 FROM users WHERE ID=?)`, p.Plan.Installation, p.Plan.TargetID, a.guardToken, p.Plan.Owner)
			return cleanupErr
		})
	}
	return err
}

func (a *aiMirrorAttempt) fence(ctx context.Context, tx *sql.Tx) error {
	p := a.standard.parent
	var token string
	if a.token == "" || tx.QueryRowContext(ctx, `SELECT leaseToken FROM ai_mirror_write_steps WHERE userID=? AND installationID=? AND operationID=? AND assetID=?`, p.Plan.Owner, p.Plan.Installation, p.ID, p.Plan.TargetID).Scan(&token) != nil || token != a.token {
		return writeback.Failure("STALE_WORKER")
	}
	return nil
}

func (a *aiMirrorAttempt) event(ctx context.Context, tx *sql.Tx, code string) error {
	p := a.standard.parent
	_, err := tx.ExecContext(ctx, `INSERT INTO ai_mirror_write_events(userID,installationID,operationID,assetID,code,at,attempt,generation) SELECT userID,installationID,operationID,assetID,?,?,attempts,generation FROM ai_mirror_write_steps WHERE userID=? AND installationID=? AND operationID=? AND assetID=? AND (SELECT count(*) FROM ai_mirror_write_events WHERE userID=? AND installationID=? AND operationID=? AND assetID=?)<100`, code, a.standard.store.drafts.results.jobs.now().UnixNano(), p.Plan.Owner, p.Plan.Installation, p.ID, p.Plan.TargetID, p.Plan.Owner, p.Plan.Installation, p.ID, p.Plan.TargetID)
	return err
}
