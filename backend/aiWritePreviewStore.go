package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writepreview"
)

func (s *aiWritePreviewSession) Publish(ctx context.Context, snapshot writepreview.Snapshot, preview writepreview.Preview, raw []byte) error {
	return s.drafts.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		current, err := s.drafts.read(ctx, tx, snapshot.Owner, snapshot.ID)
		if err != nil {
			return err
		}
		if current.Revision != snapshot.Revision || current.State != "staged" {
			return writepreview.Failure{Code: "DRAFT_CONFLICT"}
		}
		if err := s.drafts.checkAuthority(ctx, tx, snapshot.Owner, snapshot.AssetID, s.authority); err != nil {
			return err
		}
		now := s.drafts.results.jobs.now().UnixNano()
		if _, err = tx.ExecContext(ctx, `DELETE FROM ai_write_previews WHERE rowid IN (SELECT rowid FROM ai_write_previews WHERE userID=? AND installationID=? AND expiresAt<=? AND protected=0 ORDER BY expiresAt LIMIT 100)`, snapshot.Owner, snapshot.Installation, now); err != nil {
			return drafts.ErrStorage
		}
		var active int
		if err = tx.QueryRowContext(ctx, `SELECT count(*) FROM ai_write_previews WHERE userID=? AND installationID=? AND draftID=? AND revision=? AND expiresAt>? AND invalidated=0`, snapshot.Owner, snapshot.Installation, snapshot.ID, snapshot.Revision, now).Scan(&active); err != nil {
			return drafts.ErrStorage
		}
		if active >= 10 {
			return writepreview.Failure{Code: "PREVIEW_CAPACITY"}
		}
		created, err := time.Parse(time.RFC3339Nano, preview.Plan.CreatedAt)
		if err != nil {
			return drafts.ErrStorage
		}
		expires, err := time.Parse(time.RFC3339Nano, preview.Plan.ExpiresAt)
		if err != nil {
			return drafts.ErrStorage
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO ai_write_previews(userID,installationID,id,draftID,revision,version,payload,digest,createdAt,expiresAt) VALUES(?,?,?,?,?,?,?,?,?,?)`, snapshot.Owner, snapshot.Installation, preview.Plan.ID, snapshot.ID, snapshot.Revision, preview.Plan.Version, string(raw), preview.Digest, created.UnixNano(), expires.UnixNano())
		if err != nil {
			return drafts.ErrStorage
		}
		return nil
	})
}

func (s *aiWritePreviewSession) get(ctx context.Context, owner, id string) (writepreview.Preview, error) {
	var preview writepreview.Preview
	err := s.drafts.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var raw []byte
		var expires int64
		var invalidated bool
		err := tx.QueryRowContext(ctx, `SELECT payload,digest,expiresAt,invalidated FROM ai_write_previews WHERE userID=? AND installationID=? AND id=?`, owner, s.drafts.results.jobs.binding, id).Scan(&raw, &preview.Digest, &expires, &invalidated)
		if errors.Is(err, sql.ErrNoRows) {
			return drafts.ErrUnavailable
		}
		if err != nil || len(raw) > 16<<10 || json.Unmarshal(raw, &preview.Plan) != nil {
			return drafts.ErrStorage
		}
		preview.Status, preview.Diff = "usable", writepreview.Diff(preview.Plan)
		current, err := s.drafts.read(ctx, tx, owner, preview.Plan.DraftID)
		if err != nil {
			return err
		}
		if invalidated || current.Revision != preview.Plan.DraftRevision || current.State != "staged" {
			preview.Status = "stale"
		}
		if expires <= s.drafts.results.jobs.now().UnixNano() {
			preview.Status = "expired"
		}
		return nil
	})
	return preview, err
}
