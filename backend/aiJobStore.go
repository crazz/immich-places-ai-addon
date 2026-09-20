package main

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"immich-places-backend/internal/ai/jobs"
)

type aiJobStore struct {
	db      *Database
	binding string
	enabled bool
	now     func() time.Time
}

func newAIJobStore(db *Database, binding string, enabled bool, now func() time.Time) *aiJobStore {
	return &aiJobStore{db: db, binding: binding, enabled: enabled, now: now}
}

func (s *aiJobStore) write(ctx context.Context, operation func(context.Context, *sql.Tx) error) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	tx, err := s.db.db.BeginTx(ctx, nil)
	if err != nil {
		return aiJobFailure(ctx, err)
	}
	defer tx.Rollback()
	// Acquire SQLite's write lock before a read/decide/update sequence.
	if _, err = tx.ExecContext(ctx, "UPDATE ai_jobs SET calls=calls WHERE 0"); err != nil {
		return aiJobFailure(ctx, err)
	}
	if err = operation(ctx, tx); err != nil {
		return aiJobFailure(ctx, err)
	}
	return aiJobFailure(ctx, tx.Commit())
}

func aiJobFailure(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	for _, safe := range []error{jobs.ErrInvalid, jobs.ErrDenied, jobs.ErrConflict, jobs.ErrBudget, jobs.ErrLease} {
		if errors.Is(err, safe) {
			return safe
		}
	}
	return jobs.ErrStorage
}

func (s *aiJobStore) current(ctx context.Context, tx *sql.Tx) error {
	if !s.enabled {
		return jobs.ErrDenied
	}
	return s.currentInstallation(ctx, tx)
}

func (s *aiJobStore) currentInstallation(ctx context.Context, tx *sql.Tx) error {
	var current string
	if s.binding == "" {
		return jobs.ErrDenied
	}
	if err := tx.QueryRowContext(ctx, "SELECT id FROM ai_installation_identity WHERE singleton=1").Scan(&current); err != nil || current != s.binding {
		return jobs.ErrDenied
	}
	return nil
}
