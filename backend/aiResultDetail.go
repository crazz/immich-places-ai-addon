package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"reflect"
	"time"

	"github.com/google/uuid"
	"immich-places-backend/internal/ai/jobs"
	"immich-places-backend/internal/ai/results"
	"immich-places-backend/internal/ai/review"
)

func (s *aiResultStore) detail(ctx context.Context, owner, analysisID, jobID, itemID string) (review.Detail, error) {
	ids := []string{jobID, itemID}
	if analysisID != "" {
		ids = []string{analysisID}
	}
	for _, id := range ids {
		if _, err := uuid.Parse(id); err != nil {
			return review.Detail{}, review.ErrUnavailable
		}
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	tx, err := s.jobs.db.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return review.Detail{}, review.ErrUnavailable
	}
	defer tx.Rollback()
	if s.jobs.currentInstallation(ctx, tx) != nil {
		return review.Detail{}, review.ErrUnavailable
	}
	where := ` WHERE h.userID=? AND h.installationID=?`
	args := []any{owner, s.jobs.binding}
	if analysisID != "" {
		where += ` AND a.id=?`
		args = append(args, analysisID)
	} else {
		where += ` AND h.jobID=? AND h.itemID=?`
		args = append(args, jobID, itemID)
	}
	entry, err := scanAIResultEntry(tx.QueryRowContext(ctx, aiResultSummarySQL+where, args...))
	if err != nil {
		return review.Detail{}, review.ErrUnavailable
	}
	detail := review.Detail{Entry: entry}
	job, err := s.jobs.readJob(ctx, tx, owner, entry.JobID)
	if err != nil {
		return review.Detail{}, review.ErrUnavailable
	}
	if (entry.ExecutionState == "succeeded") != (entry.AnalysisID != nil) {
		return review.Detail{}, review.ErrUnavailable
	}
	if entry.AnalysisID != nil {
		var payload, metadata []byte
		err = tx.QueryRowContext(ctx, `SELECT CASE WHEN length(CAST(payload AS BLOB))<=1048576 THEN payload END,CASE WHEN length(CAST(metadata AS BLOB))<=131072 THEN metadata END FROM ai_analyses WHERE userID=? AND id=? AND jobID=? AND itemID=?`, owner, *entry.AnalysisID, entry.JobID, entry.ID).Scan(&payload, &metadata)
		if err != nil || len(payload) == 0 || len(metadata) == 0 {
			return review.Detail{}, review.ErrUnavailable
		}
		var provenance jobs.ResultMetadata
		if json.Unmarshal(metadata, &provenance) != nil || entry.ProposalOutcome == nil {
			return review.Detail{}, review.ErrUnavailable
		}
		document, err := review.ValidateRecord(payload, provenance, job, entry.AssetID, *entry.ProposalOutcome)
		if err != nil {
			return review.Detail{}, err
		}
		if provenance.Context != nil {
			lease := jobs.Lease{Owner: owner, Installation: s.jobs.binding, JobID: entry.JobID, ItemID: entry.ID, Asset: entry.AssetID}
			_, stored, err := aiJobResultContext(ctx, tx, job, lease, jobs.Completion{SourceDigest: provenance.SourceDigest, PromptVersion: provenance.PromptVersion})
			if err != nil || !reflect.DeepEqual(stored, provenance.Context) {
				return review.Detail{}, review.ErrUnavailable
			}
		}
		detail.Proposal = &document
		detail.Provenance = &provenance
	} else {
		detail.Provenance = &jobs.ResultMetadata{Mode: results.Mode(entry.Mode), Installation: job.Input.Installation, Profile: job.Input.Profile, Model: job.Model, Revision: job.Input.Revision, Languages: job.Input.Languages, PrimaryLanguage: job.Input.PrimaryLanguage, SelectionDigest: job.Input.SelectionDigest}
	}
	if tx.Commit() != nil {
		return review.Detail{}, review.ErrUnavailable
	}
	return detail, nil
}
