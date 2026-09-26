package writepreview

import (
	"slices"
	"time"

	"github.com/google/uuid"
)

func validateStackReview(snapshot Snapshot, review StackReview, current map[string]Metadata, now time.Time) error {
	if err := Validate(snapshot, snapshot.Revision); err != nil {
		return err
	}
	expires, err := time.Parse(time.RFC3339Nano, review.ExpiresAt)
	if err != nil || !now.Before(expires) || review.Owner != snapshot.Owner || review.Installation != snapshot.Installation || review.DraftID != snapshot.ID || review.DraftRevision != snapshot.Revision || review.AnalyzedID != snapshot.AssetID || review.ImageIdentity != snapshot.ImageIdentity || len(snapshot.PolicyID) != 64 {
		return Failure{Code: "TARGET_REVIEW_STALE"}
	}
	for _, id := range []string{review.ID, review.StackID} {
		if parsed, err := uuid.Parse(id); err != nil || parsed.String() != id {
			return Failure{Code: "INVALID_TARGET_SELECTION"}
		}
	}
	ids := make([]string, 0, len(review.Targets))
	for _, target := range review.Targets {
		ids = append(ids, target.AssetID)
	}
	selected, err := SelectStackTargets(snapshot, ids, ids)
	if err != nil || len(ids) < 2 || !slices.Equal(selected, ids) {
		return Failure{Code: "INVALID_TARGET_SELECTION"}
	}
	if err := Compare(snapshot, current[snapshot.AssetID]); err != nil {
		return err
	}
	for _, target := range review.Targets {
		metadata, ok := current[target.AssetID]
		if !ok || metadata.GPSUnavailable || !validObservedGPS(metadata.GPS) {
			return Failure{Code: "SOURCE_UNAVAILABLE"}
		}
		if target.ImageIdentity == "" || metadata.ImageIdentity != target.ImageIdentity {
			return Failure{Code: "SOURCE_CHANGED"}
		}
		if !EqualGPS(target.Before, metadata.GPS) {
			return Failure{Code: "TARGET_GPS_CONFLICT"}
		}
	}
	return nil
}

func validObservedGPS(gps GPS) bool {
	return (gps.Latitude == nil || finite(*gps.Latitude, 90)) && (gps.Longitude == nil || finite(*gps.Longitude, 180))
}
