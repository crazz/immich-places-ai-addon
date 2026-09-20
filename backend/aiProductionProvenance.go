package main

import (
	"context"
	"database/sql"
	"unicode/utf8"

	"immich-places-backend/internal/ai/selection"
)

func (p *aiProductionJobs) retainLaunch(ctx context.Context, tx *sql.Tx, owner, id string, manifest selection.Manifest) error {
	var albumID, label *string
	if manifest.Scope.AlbumID != "" {
		value := manifest.Scope.AlbumID
		albumID = &value
		if err := tx.QueryRowContext(ctx, `SELECT CASE WHEN length(CAST(albumName AS BLOB))<=4096 THEN albumName END FROM albums WHERE userID=? AND immichID=?`, owner, value).Scan(&label); err != nil {
			return err
		}
		if label != nil && !utf8.ValidString(*label) {
			label = nil
		}
	}
	for _, asset := range manifest.AssetIDs {
		var capture *string
		if err := tx.QueryRowContext(ctx, `SELECT CASE WHEN length(CAST(dateTimeOriginal AS BLOB))<=128 THEN NULLIF(dateTimeOriginal,'') END FROM assets WHERE userID=? AND immichID=?`, owner, asset).Scan(&capture); err != nil {
			return err
		}
		if capture != nil && !utf8.ValidString(*capture) {
			capture = nil
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO ai_job_launch VALUES(?,?,?,?,?,?)`, owner, id, asset, capture, albumID, label); err != nil {
			return err
		}
	}
	return nil
}
