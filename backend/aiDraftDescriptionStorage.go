package main

import (
	"context"
	"database/sql"
	"errors"
	"unicode/utf8"

	"immich-places-backend/internal/ai/drafts"
)

func aiScanDescription(row *sql.Row) (*drafts.DescriptionBaseline, error) {
	var value drafts.DescriptionBaseline
	err := row.Scan(&value.Presence, &value.Value)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil || !utf8.ValidString(value.Value) || len(value.Value) > 64<<10 ||
		(value.Presence != "value" && (value.Value != "" || (value.Presence != "absent" && value.Presence != "null"))) {
		return nil, drafts.ErrStorage
	}
	return &value, nil
}

func (s *aiDraftStore) saveDescriptionBaseline(ctx context.Context, tx *sql.Tx, owner string, value drafts.Draft) error {
	if value.Baseline.Description == nil {
		return nil
	}
	_, err := tx.ExecContext(ctx, `INSERT INTO ai_draft_description_baselines(userID,installationID,draftID,revision,presence,value) VALUES(?,?,?,?,?,?)`,
		owner, s.results.jobs.binding, value.ID, value.Revision, value.Baseline.Description.Presence, value.Baseline.Description.Value)
	if err != nil {
		return drafts.ErrStorage
	}
	return nil
}

func (s *aiDraftStore) saveDescriptionObservation(ctx context.Context, tx *sql.Tx, owner string, value drafts.Observation) error {
	if value.Baseline.Description == nil {
		return nil
	}
	_, err := tx.ExecContext(ctx, `INSERT INTO ai_draft_description_observations(userID,installationID,draftID,observationID,presence,value) VALUES(?,?,?,?,?,?)`,
		owner, s.results.jobs.binding, value.DraftID, value.ID, value.Baseline.Description.Presence, value.Baseline.Description.Value)
	if err != nil {
		return drafts.ErrStorage
	}
	return nil
}
