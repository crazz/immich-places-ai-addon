package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"immich-places-backend/internal/ai/jobs"
)

func (s *aiJobStore) ReadAnalysis(ctx context.Context, owner, id string) (jobs.AnalysisRecord, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	tx, err := s.db.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return jobs.AnalysisRecord{}, aiJobFailure(ctx, err)
	}
	defer tx.Rollback()
	if err = s.current(ctx, tx); err != nil {
		return jobs.AnalysisRecord{}, err
	}
	record := jobs.AnalysisRecord{ID: id}
	var metadata string
	err = tx.QueryRowContext(ctx, `SELECT a.jobID,a.itemID,a.assetID,a.outcome,a.payload,a.metadata,a.createdAt FROM ai_analyses a JOIN ai_jobs j ON j.userID=a.userID AND j.id=a.jobID WHERE a.userID=? AND a.id=? AND j.installationID=?`, owner, id, s.binding).Scan(&record.JobID, &record.ItemID, &record.Asset, &record.Outcome, &record.Payload, &metadata, &record.CreatedAt)
	if err == sql.ErrNoRows {
		return jobs.AnalysisRecord{}, jobs.ErrDenied
	}
	if err != nil {
		return jobs.AnalysisRecord{}, aiJobFailure(ctx, err)
	}
	if len(record.Payload) > 1<<20 || len(metadata) > 128<<10 || json.Unmarshal([]byte(metadata), &record.Metadata) != nil {
		return jobs.AnalysisRecord{}, jobs.ErrStorage
	}
	if err = tx.Commit(); err != nil {
		return jobs.AnalysisRecord{}, aiJobFailure(ctx, err)
	}
	return record, nil
}
