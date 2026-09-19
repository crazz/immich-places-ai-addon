package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"

	"immich-places-backend/internal/ai/selection"
)

type aiSelectionCandidate struct {
	selection.Candidate
	CaptureDay   *string  `json:"captureDay,omitempty"`
	OriginalPath string   `json:"originalPath,omitempty"`
	Latitude     *float64 `json:"latitude,omitempty"`
	Longitude    *float64 `json:"longitude,omitempty"`
}

func resolveAISelectionCandidate(ctx context.Context, tx *sql.Tx, owner, id string, scope selection.Scope) (aiSelectionCandidate, error) {
	var candidate aiSelectionCandidate
	var primary *string
	err := tx.QueryRowContext(ctx, `SELECT type,isHidden,stackPrimaryAssetID,latitude,longitude,COALESCE(originalPath,''),`+captureDaySQL("dateTimeOriginal")+`
 FROM assets WHERE userID=? AND immichID=?`+hiddenLibraryFilter, owner, id).Scan(&candidate.Type, &candidate.Hidden, &primary, &candidate.Latitude, &candidate.Longitude, &candidate.OriginalPath, &candidate.CaptureDay)
	if errors.Is(err, sql.ErrNoRows) {
		return candidate, nil
	}
	if err != nil {
		return candidate, err
	}
	candidate.Available = true
	candidate.StackChild = primary != nil
	complete := candidate.Latitude != nil && candidate.Longitude != nil
	candidate.InScope = (scope.GPSFilter == "all" || scope.GPSFilter == "with-gps" && complete || scope.GPSFilter == "no-gps" && !complete) && (scope.HiddenFilter == "all" || scope.HiddenFilter == "hidden" && candidate.Hidden || scope.HiddenFilter == "visible" && !candidate.Hidden)
	if scope.StartDate != "" || scope.EndDate != "" {
		candidate.InScope = candidate.InScope && candidate.CaptureDay != nil && (scope.StartDate == "" || *candidate.CaptureDay >= scope.StartDate) && (scope.EndDate == "" || *candidate.CaptureDay <= scope.EndDate)
	} else {
		candidate.CaptureDay = nil
	}
	if scope.View == "folder" {
		candidate.InScope = candidate.InScope && strings.HasPrefix(candidate.OriginalPath, scope.FolderPath+"/")
	} else {
		candidate.OriginalPath = ""
	}
	if scope.GPSFilter == "all" {
		candidate.Latitude = nil
		candidate.Longitude = nil
	}
	for _, membership := range []struct{ query, reference string }{
		{"SELECT EXISTS(SELECT 1 FROM albumAssets WHERE userID=? AND assetID=? AND albumID=?)", scope.AlbumID},
		{"SELECT EXISTS(SELECT 1 FROM assetTags WHERE userID=? AND assetID=? AND tagID=?)", scope.TagID},
	} {
		if membership.reference == "" {
			continue
		}
		var matches bool
		if err := tx.QueryRowContext(ctx, membership.query, owner, id, membership.reference).Scan(&matches); err != nil {
			return candidate, err
		}
		candidate.InScope = candidate.InScope && matches
	}
	return candidate, nil
}

func aiSelectionFacts(candidate aiSelectionCandidate) (string, error) {
	data, err := json.Marshal(candidate)
	return string(data), err
}

func validateAISelectionScope(ctx context.Context, tx *sql.Tx, owner string, scope selection.Scope) error {
	for _, reference := range []struct{ table, id string }{{"albums", scope.AlbumID}, {"tags", scope.TagID}} {
		if reference.id == "" {
			continue
		}
		var exists bool
		if err := tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM "+reference.table+" WHERE userID=? AND immichID=?)", owner, reference.id).Scan(&exists); err != nil {
			return err
		}
		if !exists {
			return selection.ErrInvalid
		}
	}
	return nil
}
