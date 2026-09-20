package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"immich-places-backend/internal/ai/jobs"
)

type aiJobPage struct {
	Items      []aiJobProgress `json:"items"`
	NextCursor string          `json:"nextCursor,omitempty"`
}
type aiJobCursor struct {
	Kind, Owner, Installation, ID string
	CreatedAt                     int64
	Limit                         int
}

func (p *aiProductionJobs) list(ctx context.Context, owner string, limit int, cursor string) (aiJobPage, error) {
	if limit < 1 || limit > 100 || len(cursor) > 2048 {
		return aiJobPage{}, jobs.ErrInvalid
	}
	continuation := aiJobCursor{Kind: "jobs-v1", Owner: owner, Installation: p.store.binding, Limit: limit}
	if cursor != "" {
		if !strings.HasPrefix(cursor, encryptedPrefix) {
			return aiJobPage{}, jobs.ErrInvalid
		}
		raw, err := decryptValue(p.store.db.encryptionKey, cursor)
		if err != nil || json.Unmarshal([]byte(raw), &continuation) != nil || continuation.Kind != "jobs-v1" || continuation.Owner != owner || continuation.Installation != p.store.binding || continuation.Limit != limit || continuation.CreatedAt <= 0 {
			return aiJobPage{}, jobs.ErrInvalid
		}
		if _, err := uuid.Parse(continuation.ID); err != nil {
			return aiJobPage{}, jobs.ErrInvalid
		}
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	tx, err := p.store.db.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return aiJobPage{}, jobs.ErrStorage
	}
	defer tx.Rollback()
	if err = p.store.currentInstallation(ctx, tx); err != nil {
		return aiJobPage{}, err
	}
	rows, err := tx.QueryContext(ctx, `SELECT j.id FROM ai_jobs j JOIN ai_job_admissions a ON a.userID=j.userID AND a.jobID=j.id WHERE j.userID=? AND j.installationID=? AND (?='' OR j.createdAt<? OR (j.createdAt=? AND j.id<?)) ORDER BY j.createdAt DESC,j.id DESC LIMIT ?`, owner, p.store.binding, continuation.ID, continuation.CreatedAt, continuation.CreatedAt, continuation.ID, limit+1)
	if err != nil {
		return aiJobPage{}, jobs.ErrStorage
	}
	ids := []string{}
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			break
		}
		ids = append(ids, id)
	}
	rowErr := rows.Err()
	rows.Close()
	if err != nil || rowErr != nil {
		return aiJobPage{}, jobs.ErrStorage
	}
	more := len(ids) > limit
	if more {
		ids = ids[:limit]
	}
	page := aiJobPage{Items: []aiJobProgress{}}
	for _, id := range ids {
		item, err := p.readProgress(ctx, tx, owner, id)
		if err != nil {
			return aiJobPage{}, aiJobFailure(ctx, err)
		}
		item.Items = nil
		page.Items = append(page.Items, item)
	}
	if more {
		last := page.Items[len(page.Items)-1]
		continuation.ID = last.ID
		continuation.CreatedAt = last.CreatedAt
		raw, _ := json.Marshal(continuation)
		page.NextCursor, err = encryptValue(p.store.db.encryptionKey, string(raw))
		if err != nil {
			return aiJobPage{}, jobs.ErrStorage
		}
	}
	if err = tx.Commit(); err != nil {
		return aiJobPage{}, jobs.ErrStorage
	}
	return page, nil
}
