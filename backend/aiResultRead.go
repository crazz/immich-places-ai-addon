package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"immich-places-backend/internal/ai/jobs"
	"immich-places-backend/internal/ai/review"
	"time"
)

type aiResultStore struct{ jobs *aiJobStore }

func (s *aiResultStore) list(ctx context.Context, owner string, q review.Query) (review.Page, error) {
	if q.Limit < 1 || q.Limit > 100 {
		return review.Page{}, review.ErrInvalid
	}
	continuation, err := s.cursor(owner, q)
	if err != nil {
		return review.Page{}, err
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	tx, err := s.jobs.db.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return review.Page{}, review.ErrUnavailable
	}
	defer tx.Rollback()
	if err = s.jobs.currentInstallation(ctx, tx); err != nil {
		return review.Page{}, review.ErrUnavailable
	}
	if q.Cursor == "" {
		if err = tx.QueryRowContext(ctx, `SELECT COALESCE(max(sequence),0) FROM ai_result_history WHERE userID=? AND installationID=?`, owner, s.jobs.binding).Scan(&continuation.Watermark); err != nil {
			return review.Page{}, review.ErrUnavailable
		}
	}
	where := ` WHERE h.userID=? AND h.installationID=? AND h.sequence<=?`
	args := []any{owner, s.jobs.binding, continuation.Watermark}
	filters, filterArgs := aiResultFilters(q)
	where += filters
	args = append(args, filterArgs...)
	if continuation.Item != "" {
		where += ` AND (h.terminalAt,h.jobID,h.itemID)<(?,?,?)`
		args = append(args, continuation.TerminalAt, continuation.Job, continuation.Item)
	}
	args = append(args, q.Limit+1)
	rows, err := tx.QueryContext(ctx, aiResultSummarySQL+where+` ORDER BY h.terminalAt DESC,h.jobID DESC,h.itemID DESC LIMIT ?`, args...)
	if err != nil {
		return review.Page{}, review.ErrUnavailable
	}
	defer rows.Close()
	page := review.Page{Items: []review.Entry{}}
	for rows.Next() {
		entry, err := scanAIResultEntry(rows)
		if err != nil {
			return review.Page{}, review.ErrUnavailable
		}
		page.Items = append(page.Items, entry)
	}
	if rows.Err() != nil {
		return review.Page{}, review.ErrUnavailable
	}
	if len(page.Items) > q.Limit {
		page.Items = page.Items[:q.Limit]
		last := page.Items[len(page.Items)-1]
		stamp, err := time.Parse(time.RFC3339Nano, last.TerminalAt)
		if err != nil {
			return review.Page{}, review.ErrUnavailable
		}
		continuation.TerminalAt, continuation.Job, continuation.Item = stamp.UnixNano(), last.JobID, last.ID
		page.NextCursor, err = s.encodeCursor(continuation)
		if err != nil {
			return review.Page{}, err
		}
	}
	if err = tx.Commit(); err != nil {
		return review.Page{}, review.ErrUnavailable
	}
	return page, nil
}

const aiResultSummarySQL = `SELECT h.itemID,h.jobID,h.assetID,h.terminalAt,h.executionState,h.captureDay,h.albumID,a.id,a.outcome,j.model,j.requestJSON,i.failure,l.albumLabel,h.label,
 EXISTS(SELECT 1 FROM assets src WHERE src.userID=h.userID AND src.immichID=h.assetID AND src.isHidden=0 AND src.type='IMAGE' AND (src.libraryID IS NULL OR NOT EXISTS(SELECT 1 FROM libraries lib WHERE lib.libraryID=src.libraryID AND lib.isHidden=1)))
 FROM ai_result_history h JOIN ai_jobs j ON j.userID=h.userID AND j.id=h.jobID AND j.installationID=h.installationID
 JOIN ai_job_items i ON i.userID=h.userID AND i.jobID=h.jobID AND i.id=h.itemID
 LEFT JOIN ai_analyses a ON a.userID=h.userID AND a.jobID=h.jobID AND a.itemID=h.itemID
 LEFT JOIN ai_job_launch l ON l.userID=h.userID AND l.jobID=h.jobID AND l.assetID=h.assetID`

type aiResultScanner interface{ Scan(...any) error }

func scanAIResultEntry(row aiResultScanner) (review.Entry, error) {
	var e review.Entry
	var terminal int64
	var raw string
	err := row.Scan(&e.ID, &e.JobID, &e.AssetID, &terminal, &e.ExecutionState, &e.CaptureDay, &e.AlbumID, &e.AnalysisID, &e.ProposalOutcome, &e.Model, &raw, &e.Failure, &e.AlbumLabel, &e.Label, &e.SourceAvailable)
	if err != nil {
		return e, err
	}
	var input jobs.Submission
	if len(raw) > 128<<10 || json.Unmarshal([]byte(raw), &input) != nil {
		return e, review.ErrUnavailable
	}
	e.Mode = input.Mode
	if e.Mode == "" {
		e.Mode = "visual"
	}
	e.TerminalAt = time.Unix(0, terminal).UTC().Format(time.RFC3339Nano)
	e.ReviewState = "unreviewed"
	e.WriteState = "not_requested"
	return e, nil
}
