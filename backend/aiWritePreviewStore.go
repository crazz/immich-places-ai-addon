package main

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/writeback"
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
		if (snapshot.Description != nil || preview.Plan.Manifest != nil) && snapshot.PolicyID != s.policyIdentity() {
			return writepreview.Failure{Code: "POLICY_CHANGED"}
		}
		if err := s.drafts.checkAuthority(ctx, tx, snapshot.Owner, snapshot.AssetID, s.authority); err != nil {
			return err
		}
		if preview.Plan.Manifest != nil {
			if len(s.targetAuthorities) != len(preview.Plan.Manifest.Targets) {
				return drafts.ErrUnavailable
			}
			for _, target := range preview.Plan.Manifest.Targets {
				if err := s.drafts.checkAuthority(ctx, tx, snapshot.Owner, target.AssetID, s.targetAuthorities[target.AssetID]); err != nil {
					return err
				}
			}
		}
		now := s.drafts.results.jobs.now().UnixNano()
		for _, table := range []string{"ai_write_previews", "ai_standard_write_previews", "ai_stack_write_previews"} {
			if _, err = tx.ExecContext(ctx, `DELETE FROM `+table+` WHERE rowid IN (SELECT rowid FROM `+table+` WHERE userID=? AND installationID=? AND expiresAt<=? AND protected=0 ORDER BY expiresAt LIMIT 100)`, snapshot.Owner, snapshot.Installation, now); err != nil {
				return drafts.ErrStorage
			}
		}
		var active int
		if err = tx.QueryRowContext(ctx, `SELECT count(*) FROM ai_all_write_previews WHERE userID=? AND installationID=? AND draftID=? AND revision=? AND expiresAt>? AND invalidated=0`, snapshot.Owner, snapshot.Installation, snapshot.ID, snapshot.Revision, now).Scan(&active); err != nil {
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
		table := "ai_write_previews"
		if preview.Plan.Version == "standard-preview-v2" {
			table = "ai_standard_write_previews"
		}
		if preview.Plan.Version == "stack-preview-v3" || preview.Plan.Version == "mirror-preview-v4" {
			table = "ai_stack_write_previews"
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO `+table+`(userID,installationID,id,draftID,revision,version,payload,digest,createdAt,expiresAt) VALUES(?,?,?,?,?,?,?,?,?,?)`, snapshot.Owner, snapshot.Installation, preview.Plan.ID, snapshot.ID, snapshot.Revision, preview.Plan.Version, string(raw), preview.Digest, created.UnixNano(), expires.UnixNano())
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
		err := tx.QueryRowContext(ctx, `SELECT payload,digest,expiresAt,invalidated FROM ai_all_write_previews WHERE userID=? AND installationID=? AND id=?`, owner, s.drafts.results.jobs.binding, id).Scan(&raw, &preview.Digest, &expires, &invalidated)
		if errors.Is(err, sql.ErrNoRows) {
			return drafts.ErrUnavailable
		}
		if err != nil {
			return drafts.ErrStorage
		}
		preview.Plan, err = writeback.DecodePlan(raw, preview.Digest, owner, s.drafts.results.jobs.binding, id)
		if err != nil {
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
