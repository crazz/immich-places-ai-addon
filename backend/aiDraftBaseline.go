package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"reflect"
	"time"

	"github.com/google/uuid"
	"immich-places-backend/internal/ai/drafts"
)

func (s *aiDraftStore) observe(ctx context.Context, owner, id string, revision int, reader *aiImagePreparer) (drafts.Observation, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	value, err := s.get(ctx, owner, id)
	if err != nil {
		return drafts.Observation{}, err
	}
	if value.Revision != revision {
		return drafts.Observation{}, drafts.ErrConflict
	}
	baseline, authority, err := s.source(ctx, owner, value.AssetID, reader)
	if err != nil {
		return drafts.Observation{}, err
	}
	now := s.results.jobs.now()
	baseline.ObservedAt = now.UTC().Format(time.RFC3339Nano)
	observation := drafts.Observation{ID: uuid.NewString(), DraftID: id, Revision: revision, Baseline: baseline, ExpiresAt: now.Add(5 * time.Minute).UTC().Format(time.RFC3339Nano), OriginalSourceMatches: baseline.SourceDigest == value.OriginalSourceDigest}
	detail, err := s.results.detail(ctx, owner, value.AnalysisID, "", "")
	if err != nil {
		return drafts.Observation{}, drafts.ErrUnavailable
	}
	observation.PreviewURL = "/ai/jobs/" + detail.Entry.JobID + "/items/" + detail.Entry.ID + "/thumbnail"
	err = s.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		current, err := s.read(ctx, tx, owner, id)
		if err != nil {
			return err
		}
		if current.Revision != revision {
			return drafts.ErrConflict
		}
		if err := s.checkAuthority(ctx, tx, owner, value.AssetID, authority); err != nil {
			return drafts.ErrUnavailable
		}
		if _, err = tx.ExecContext(ctx, `DELETE FROM ai_draft_baseline_observations WHERE rowid IN (SELECT rowid FROM ai_draft_baseline_observations WHERE userID=? AND installationID=? AND expiresAt<=? ORDER BY expiresAt LIMIT 100)`, owner, s.results.jobs.binding, now.UnixNano()); err != nil {
			return drafts.ErrStorage
		}
		var active int
		if err = tx.QueryRowContext(ctx, `SELECT count(*) FROM ai_draft_baseline_observations WHERE userID=? AND installationID=? AND draftID=? AND expiresAt>?`, owner, s.results.jobs.binding, id, now.UnixNano()).Scan(&active); err != nil {
			return drafts.ErrStorage
		}
		if active >= 10 {
			return drafts.ErrConflict
		}
		stored := observation
		stored.Baseline.Description = nil
		data, err := json.Marshal(stored)
		if err != nil {
			return drafts.ErrStorage
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO ai_draft_baseline_observations(userID,installationID,draftID,revision,id,expiresAt,content) VALUES(?,?,?,?,?,?,?)`, owner, s.results.jobs.binding, id, revision, observation.ID, now.Add(5*time.Minute).UnixNano(), string(data))
		if err != nil {
			return drafts.ErrStorage
		}
		return s.saveDescriptionObservation(ctx, tx, owner, observation)
	})
	return observation, err
}

func (s *aiDraftStore) acknowledge(ctx context.Context, owner, id string, revision int, observationID string, reader *aiImagePreparer) (drafts.Draft, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	var observation drafts.Observation
	value, err := s.get(ctx, owner, id)
	if err != nil {
		return value, err
	}
	if value.Revision != revision {
		return drafts.Draft{}, drafts.ErrConflict
	}
	var data []byte
	var expires int64
	err = s.results.jobs.db.db.QueryRowContext(ctx, `SELECT content,expiresAt FROM ai_draft_baseline_observations WHERE userID=? AND installationID=? AND draftID=? AND revision=? AND id=?`, owner, s.results.jobs.binding, id, revision, observationID).Scan(&data, &expires)
	if err != nil || json.Unmarshal(data, &observation) != nil || expires <= s.results.jobs.now().UnixNano() {
		return drafts.Draft{}, drafts.ErrConflict
	}
	observation.Baseline.Description, err = aiScanDescription(s.results.jobs.db.db.QueryRowContext(ctx, `SELECT presence,value FROM ai_draft_description_observations WHERE userID=? AND installationID=? AND draftID=? AND observationID=?`, owner, s.results.jobs.binding, id, observationID))
	if err != nil {
		return drafts.Draft{}, err
	}
	baseline, authority, err := s.source(ctx, owner, value.AssetID, reader)
	if err != nil {
		return drafts.Draft{}, err
	}
	baseline.ObservedAt = observation.Baseline.ObservedAt
	if !reflect.DeepEqual(baseline, observation.Baseline) {
		return drafts.Draft{}, drafts.ErrConflict
	}
	err = s.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		current, err := s.read(ctx, tx, owner, id)
		if err != nil {
			return err
		}
		if current.Revision != revision || expires <= s.results.jobs.now().UnixNano() {
			return drafts.ErrConflict
		}
		if err := s.checkAuthority(ctx, tx, owner, value.AssetID, authority); err != nil {
			return drafts.ErrUnavailable
		}
		value = current
		value.Revision++
		value.Baseline = baseline
		if value.State == "staged" {
			value.State = "draft"
		}
		value.UpdatedAt = s.results.jobs.now().UTC().Format(time.RFC3339Nano)
		if err = s.snapshot(ctx, tx, owner, value); err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, `UPDATE ai_drafts SET revision=?,state=? WHERE userID=? AND installationID=? AND id=?`, value.Revision, value.State, owner, s.results.jobs.binding, id); err != nil {
			return drafts.ErrStorage
		}
		return nil
	})
	return value, err
}
