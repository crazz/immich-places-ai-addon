package main

import (
	"context"
	"database/sql"
	"strings"

	"immich-places-backend/internal/ai/selection"
)

// enumerateAISelectionMatching streams distinct catalog rows inside the caller's
// transaction. Discovery deliberately reuses gallery suppression and filters.
func enumerateAISelectionMatching(ctx context.Context, tx *sql.Tx, owner string, scope selection.Scope, yield func(selection.Match) error) error {
	filter := buildAssetFilter(owner, scope.AlbumID, scope.TagID, scope.GPSFilter, scope.HiddenFilter, scope.StartDate, scope.EndDate)
	prefix := ""
	if filter.aliased {
		prefix = "a."
	}
	if scope.View == "folder" {
		filter.fromClause += " AND " + prefix + "originalPath >= ? AND " + prefix + "originalPath < ?"
		filter.args = append(filter.args, scope.FolderPath+"/", scope.FolderPath+"0")
	}
	columns := []string{prefix + "immichID", prefix + "type", prefix + "isHidden", "NULL", "NULL", "''", "NULL"}
	if scope.GPSFilter != "all" {
		columns[3] = prefix + "latitude"
		columns[4] = prefix + "longitude"
	}
	if scope.View == "folder" {
		columns[5] = "COALESCE(" + prefix + "originalPath,'')"
	}
	if scope.StartDate != "" || scope.EndDate != "" {
		columns[6] = captureDaySQL(prefix + "dateTimeOriginal")
	}
	rows, err := tx.QueryContext(ctx, "SELECT DISTINCT "+strings.Join(columns, ",")+" "+filter.fromClause+" ORDER BY "+prefix+"fileCreatedAt DESC,"+prefix+"immichID DESC", filter.args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		if err := ctx.Err(); err != nil {
			return err
		}
		var id string
		candidate := aiSelectionCandidate{Candidate: selection.Candidate{Available: true, InScope: true}}
		if err := rows.Scan(&id, &candidate.Type, &candidate.Hidden, &candidate.Latitude, &candidate.Longitude, &candidate.OriginalPath, &candidate.CaptureDay); err != nil {
			return err
		}
		facts := ""
		if selection.ExclusionReason(candidate.Candidate) == "" {
			facts, err = aiSelectionFacts(candidate)
			if err != nil {
				return err
			}
		}
		if err := yield(selection.Match{AssetID: id, Candidate: candidate.Candidate, Facts: facts}); err != nil {
			return err
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return ctx.Err()
}
