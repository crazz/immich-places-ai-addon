package main

import (
	"context"
	"database/sql"
	"unicode/utf8"

	"immich-places-backend/internal/ai/selection"
)

func aiSelectionContextPreview(ctx context.Context, tx *sql.Tx, owner string, manifest selection.Manifest) (*selection.ContextPreview, error) {
	preview := &selection.ContextPreview{CaptureTimes: map[string]string{}}
	if manifest.Scope.AlbumID != "" {
		if err := tx.QueryRowContext(ctx, `SELECT CASE WHEN length(CAST(albumName AS BLOB))<=256 THEN albumName END FROM albums WHERE userID=? AND immichID=?`, owner, manifest.Scope.AlbumID).Scan(&preview.AlbumLabel); err != nil {
			return nil, err
		}
		if preview.AlbumLabel != nil && !utf8.ValidString(*preview.AlbumLabel) {
			preview.AlbumLabel = nil
		}
	}
	for _, asset := range manifest.AssetIDs {
		var capture *string
		if err := tx.QueryRowContext(ctx, `SELECT CASE WHEN length(CAST(dateTimeOriginal AS BLOB))<=128 THEN NULLIF(dateTimeOriginal,'') END FROM assets WHERE userID=? AND immichID=?`, owner, asset).Scan(&capture); err != nil {
			return nil, err
		}
		if capture != nil && utf8.ValidString(*capture) {
			preview.CaptureTimes[asset] = *capture
		}
	}
	return preview, nil
}
