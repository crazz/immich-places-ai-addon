package main

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var errAISelectionOwnerQuota = errors.New("owner selection quota exhausted")
var errAISelectionCapacity = errors.New("selection capacity exhausted")

func aiSelectionQuota(ctx context.Context, tx *sql.Tx, owner string, now time.Time) error {
	var owned int
	if err := tx.QueryRowContext(ctx, "SELECT count(*) FROM ai_selection_snapshots WHERE userID=? AND expiresAt>?", owner, now.UnixNano()).Scan(&owned); err != nil {
		return err
	}
	if owned >= 20 {
		return errAISelectionOwnerQuota
	}
	var retained int
	if err := tx.QueryRowContext(ctx, "SELECT count(*) FROM ai_selection_snapshots").Scan(&retained); err != nil {
		return err
	}
	if retained >= 1000 {
		return errAISelectionCapacity
	}
	return nil
}

func (s *aiSelectionStore) cleanup(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	_, err := s.db.db.ExecContext(ctx, `DELETE FROM ai_selection_snapshots WHERE rowid IN
 (SELECT rowid FROM ai_selection_snapshots WHERE expiresAt<=? ORDER BY expiresAt LIMIT 100)`, s.now().UnixNano())
	return err
}

func (s *aiSelectionStore) runCleanup(ctx context.Context, ticks <-chan time.Time, failed func()) {
	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-ticks:
			if !ok {
				return
			}
			if err := s.cleanup(ctx); err != nil && ctx.Err() == nil {
				failed()
			}
		}
	}
}
