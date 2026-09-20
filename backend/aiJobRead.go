package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"immich-places-backend/internal/ai/jobs"
)

func (s *aiJobStore) Get(ctx context.Context, owner, id string) (jobs.Job, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	tx, err := s.db.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return jobs.Job{}, aiJobFailure(ctx, err)
	}
	defer tx.Rollback()
	if err = s.current(ctx, tx); err != nil {
		return jobs.Job{}, err
	}
	job, err := s.readJob(ctx, tx, owner, id)
	if err != nil {
		return jobs.Job{}, aiJobFailure(ctx, err)
	}
	if err = tx.Commit(); err != nil {
		return jobs.Job{}, aiJobFailure(ctx, err)
	}
	return job, nil
}

func (s *aiJobStore) readJob(ctx context.Context, tx *sql.Tx, owner, id string) (jobs.Job, error) {
	job := jobs.Job{ID: id, Items: []jobs.Item{}}
	var data string
	if err := tx.QueryRowContext(ctx, `SELECT requestJSON,model,calls,cancelRequested,createdAt FROM ai_jobs WHERE userID=? AND id=? AND installationID=?`, owner, id, s.binding).Scan(&data, &job.Model, &job.Calls, &job.Canceled, &job.CreatedAt); err != nil {
		return jobs.Job{}, jobs.ErrDenied
	}
	if len(data) > 128<<10 || json.Unmarshal([]byte(data), &job.Input) != nil {
		return jobs.Job{}, jobs.ErrStorage
	}
	rows, err := tx.QueryContext(ctx, `SELECT id,assetID,state,attempts,calls,nextAttemptAt,failure FROM ai_job_items WHERE userID=? AND jobID=? ORDER BY position`, owner, id)
	if err != nil {
		return jobs.Job{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var item jobs.Item
		if err := rows.Scan(&item.ID, &item.Asset, &item.State, &item.Attempts, &item.Calls, &item.NextAttemptAt, &item.Failure); err != nil {
			return jobs.Job{}, err
		}
		job.Items = append(job.Items, item)
	}
	return job, rows.Err()
}
