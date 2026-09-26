package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/jobs"
	"immich-places-backend/internal/ai/translations"
)

type aiTranslationStore struct {
	drafts      *aiDraftStore
	policies    jobs.ExecutionPolicies
	fingerprint string
	capacity    jobs.Policy
}

type aiTranslationRun struct {
	ID       string               `json:"id"`
	Request  translations.Request `json:"request"`
	Items    []aiTranslationItem  `json:"items"`
	PolicyID string               `json:"policyId"`
	Policy   jobs.ExecutionPolicy `json:"-"`
}

type aiTranslationItem struct {
	Language string  `json:"language"`
	State    string  `json:"state"`
	Text     *string `json:"text"`
	Failure  string  `json:"failure"`
}

func (s *aiTranslationStore) get(ctx context.Context, owner, id string) (aiTranslationRun, error) {
	var run aiTranslationRun
	err := s.drafts.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var err error
		run, err = s.read(ctx, tx, owner, id)
		return err
	})
	return run, err
}

func (s *aiTranslationStore) submit(ctx context.Context, owner string, req translations.Request) (aiTranslationRun, error) {
	req, digest, err := translations.Normalize(req)
	if err != nil {
		return aiTranslationRun{}, drafts.ErrInvalid
	}
	var run aiTranslationRun
	err = s.drafts.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		binding := s.drafts.results.jobs.binding
		var id, storedDigest string
		err := tx.QueryRowContext(ctx, `SELECT id,requestDigest FROM ai_translation_runs WHERE userID=? AND installationID=? AND idempotencyKey=?`, owner, binding, req.Key).Scan(&id, &storedDigest)
		if err == nil {
			if storedDigest != digest {
				return drafts.ErrConflict
			}
			run, err = s.read(ctx, tx, owner, id)
			return err
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return drafts.ErrStorage
		}
		if !s.drafts.results.jobs.enabled {
			return drafts.ErrUnavailable
		}
		current, err := s.drafts.read(ctx, tx, owner, req.DraftID)
		if err != nil {
			return err
		}
		if current.Revision != req.Revision || current.FactsRevision != req.FactsRevision {
			return drafts.ErrConflict
		}
		if err = s.validateParent(ctx, tx, owner, req); err != nil {
			return err
		}
		var active bool
		if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM ai_translation_runs r JOIN ai_translation_items i ON i.userID=r.userID AND i.installationID=r.installationID AND i.runID=r.id WHERE r.userID=? AND r.installationID=? AND r.draftID=? AND i.state IN ('queued','reserved'))`, owner, binding, req.DraftID).Scan(&active); err != nil {
			return drafts.ErrStorage
		}
		if active {
			return drafts.ErrConflict
		}
		var total, owned int
		if tx.QueryRowContext(ctx, `SELECT count(*),COALESCE(sum(userID=?),0) FROM (SELECT i.userID FROM ai_job_items i JOIN ai_jobs j ON j.userID=i.userID AND j.id=i.jobID WHERE j.installationID=? AND i.state IN ('queued','running','retry_wait','blocked') UNION ALL SELECT userID FROM ai_translation_items WHERE installationID=? AND state IN ('queued','reserved'))`, owner, binding, binding).Scan(&total, &owned) != nil {
			return drafts.ErrStorage
		}
		if total+len(req.Languages) > 500 || owned+len(req.Languages) > 100 {
			return drafts.ErrConflict
		}
		id = uuid.NewString()
		policy, policyID, err := s.authority(ctx, tx, owner, req)
		if err != nil {
			return err
		}
		policyJSON, err := json.Marshal(policy)
		if err != nil {
			return drafts.ErrStorage
		}
		raw, err := json.Marshal(req)
		if err != nil {
			return drafts.ErrInvalid
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO ai_translation_runs VALUES(?,?,?,?,?,?,?,?,?,?,?)`, owner, binding, id, req.DraftID, req.Revision, req.Key, digest, string(raw), policyID, string(policyJSON), s.drafts.results.jobs.now().UnixNano())
		if err != nil {
			return drafts.ErrStorage
		}
		for _, tag := range req.Languages {
			if _, err = tx.ExecContext(ctx, `INSERT INTO ai_translation_items(userID,installationID,runID,language,state) VALUES(?,?,?,?,'queued')`, owner, binding, id, tag); err != nil {
				return drafts.ErrStorage
			}
		}
		run, err = s.read(ctx, tx, owner, id)
		return err
	})
	return run, err
}

func (s *aiTranslationStore) read(ctx context.Context, tx *sql.Tx, owner, id string) (aiTranslationRun, error) {
	run := aiTranslationRun{ID: id, Items: []aiTranslationItem{}}
	var raw, policy string
	binding := s.drafts.results.jobs.binding
	err := tx.QueryRowContext(ctx, `SELECT requestJSON,policyID,policyJSON FROM ai_translation_runs WHERE userID=? AND installationID=? AND id=?`, owner, binding, id).Scan(&raw, &run.PolicyID, &policy)
	if errors.Is(err, sql.ErrNoRows) {
		return run, drafts.ErrUnavailable
	}
	if err != nil || json.Unmarshal([]byte(raw), &run.Request) != nil || json.Unmarshal([]byte(policy), &run.Policy) != nil {
		return run, drafts.ErrStorage
	}
	rows, err := tx.QueryContext(ctx, `SELECT language,state,text,failure FROM ai_translation_items WHERE userID=? AND installationID=? AND runID=? ORDER BY language`, owner, binding, id)
	if err != nil {
		return run, drafts.ErrStorage
	}
	defer rows.Close()
	for rows.Next() {
		var item aiTranslationItem
		if rows.Scan(&item.Language, &item.State, &item.Text, &item.Failure) != nil {
			return run, drafts.ErrStorage
		}
		run.Items = append(run.Items, item)
	}
	if rows.Err() != nil {
		return run, drafts.ErrStorage
	}
	return run, nil
}
