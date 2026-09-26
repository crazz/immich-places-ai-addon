package main

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writeback"
)

func (s *aiWriteStore) reconcileMirror(ctx context.Context, owner, id, asset string, generation int) (writeback.Operation, error) {
	var parent writeback.Operation
	err := s.drafts.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var err error
		parent, err = s.readStackOperation(ctx, tx, owner, id, false)
		if err != nil {
			return drafts.ErrUnavailable
		}
		m := parent.Mirror
		if m == nil || m.AssetID != asset || generation < 1 || m.Generation != generation || (m.Status != "writing" && m.Status != "verifying") {
			return writeback.Failure("RECONCILIATION_UNAVAILABLE")
		}
		now := s.drafts.results.jobs.now()
		result, err := tx.ExecContext(ctx, `UPDATE ai_mirror_write_steps SET reads=0,dueAt=?,generation=generation+1,leaseToken=?,leaseUntil=0 WHERE userID=? AND installationID=? AND operationID=? AND assetID=? AND generation=? AND reads=3 AND (senderActive=0 OR leaseUntil<=?)`, now.Add(time.Second).UnixNano(), uuid.NewString(), owner, parent.Plan.Installation, id, asset, generation, now.UnixNano())
		if err != nil {
			return drafts.ErrStorage
		}
		if n, err := result.RowsAffected(); err != nil || n != 1 {
			return writeback.Failure("RECONCILIATION_UNAVAILABLE")
		}
		a := &aiMirrorAttempt{standard: &aiStackAttempt{store: s, parent: parent, assetID: asset}}
		if err := a.event(ctx, tx, "EXPLICIT_RECONCILIATION"); err != nil {
			return err
		}
		parent, err = s.readStackOperation(ctx, tx, owner, id, false)
		return err
	})
	return parent, err
}
