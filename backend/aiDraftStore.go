package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"immich-places-backend/internal/ai/drafts"
)

type aiDraftStore struct{ results *aiResultStore }

func (s *aiDraftStore) write(ctx context.Context, operation func(context.Context, *sql.Tx) error) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	tx, err := s.results.jobs.db.db.BeginTx(ctx, nil)
	if err != nil {
		return drafts.ErrStorage
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, "UPDATE ai_drafts SET revision=revision WHERE 0"); err != nil {
		return drafts.ErrStorage
	}
	if s.results.jobs.currentInstallation(ctx, tx) != nil {
		return drafts.ErrUnavailable
	}
	if err = operation(ctx, tx); err != nil {
		return err
	}
	if tx.Commit() != nil {
		return drafts.ErrStorage
	}
	return nil
}

func (s *aiDraftStore) read(ctx context.Context, tx *sql.Tx, owner, id string) (drafts.Draft, error) {
	var raw []byte
	err := tx.QueryRowContext(ctx, `SELECT r.content FROM ai_drafts d JOIN ai_draft_revisions r ON
 r.userID=d.userID AND r.installationID=d.installationID AND r.draftID=d.id AND r.revision=d.revision
 WHERE d.userID=? AND d.installationID=? AND d.id=?`, owner, s.results.jobs.binding, id).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return drafts.Draft{}, drafts.ErrUnavailable
	}
	var value drafts.Draft
	if err != nil || len(raw) > 128<<10 || json.Unmarshal(raw, &value) != nil {
		return value, drafts.ErrStorage
	}
	return value, nil
}

func (s *aiDraftStore) get(ctx context.Context, owner, id string) (drafts.Draft, error) {
	var value drafts.Draft
	err := s.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var err error
		value, err = s.read(ctx, tx, owner, id)
		return err
	})
	return value, err
}

func (s *aiDraftStore) accept(ctx context.Context, owner, analysis string, candidate *string) (drafts.Draft, error) {
	detail, err := s.results.detail(ctx, owner, analysis, "", "")
	if err != nil || detail.Proposal == nil || detail.Provenance == nil {
		return drafts.Draft{}, drafts.ErrUnavailable
	}
	var value drafts.Draft
	err = s.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var id string
		err := tx.QueryRowContext(ctx, `SELECT id FROM ai_drafts WHERE userID=? AND installationID=? AND analysisID=?`, owner, s.results.jobs.binding, analysis).Scan(&id)
		if err == nil {
			value, err = s.read(ctx, tx, owner, id)
			return err
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return drafts.ErrStorage
		}
		var exists int
		if tx.QueryRowContext(ctx, `SELECT 1 FROM ai_analyses a JOIN ai_jobs j ON j.userID=a.userID AND j.id=a.jobID WHERE a.userID=? AND a.id=? AND j.installationID=?`, owner, analysis, s.results.jobs.binding).Scan(&exists) != nil {
			return drafts.ErrUnavailable
		}
		value = drafts.Draft{ID: uuid.NewString(), AnalysisID: analysis, AssetID: detail.Entry.AssetID, Revision: 1, State: "draft", Fields: []string{}, OriginalSourceDigest: detail.Provenance.SourceDigest, Baseline: drafts.Baseline{Status: "unavailable"}, UpdatedAt: s.results.jobs.now().UTC().Format(time.RFC3339Nano)}
		value, err = drafts.FromProposal(value, *detail.Proposal, candidate)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO ai_drafts(userID,installationID,id,analysisID,revision,state) VALUES(?,?,?,?,?,?)`, owner, s.results.jobs.binding, value.ID, analysis, value.Revision, value.State); err != nil {
			return drafts.ErrStorage
		}
		return s.snapshot(ctx, tx, owner, value)
	})
	return value, err
}

func (s *aiDraftStore) snapshot(ctx context.Context, tx *sql.Tx, owner string, value drafts.Draft) error {
	if err := s.guardWriteRevision(ctx, tx, owner, value.ID); err != nil {
		return err
	}
	data, err := json.Marshal(value)
	if err != nil || len(data) > 128<<10 {
		return drafts.ErrInvalid
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO ai_draft_revisions(userID,installationID,draftID,revision,content) VALUES(?,?,?,?,?)`, owner, s.results.jobs.binding, value.ID, value.Revision, string(data))
	if err != nil {
		return drafts.ErrStorage
	}
	return nil
}
