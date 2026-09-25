package writepreview

import "math"

type Conflict struct {
	Before   GPS   `json:"before"`
	Current  GPS   `json:"current"`
	Proposed Point `json:"proposed"`
}

type Failure struct {
	Code     string
	Conflict *Conflict
}

func (f Failure) Error() string { return f.Code }

func Validate(snapshot Snapshot, revision int) error {
	if snapshot.Revision != revision {
		return Failure{Code: "DRAFT_CONFLICT"}
	}
	if snapshot.State != "staged" {
		return Failure{Code: "DRAFT_NOT_STAGED"}
	}
	if snapshot.Camera == nil || !finite(snapshot.Camera.Latitude, 90) || !finite(snapshot.Camera.Longitude, 180) || len(snapshot.Fields) != 1 || snapshot.Fields[0] != "gps" {
		return Failure{Code: "INVALID_PREVIEW"}
	}
	if !snapshot.BaselineReviewed || snapshot.ImageIdentity == "" {
		return Failure{Code: "BASELINE_REVIEW_REQUIRED"}
	}
	return nil
}

func finite(value, limit float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= -limit && value <= limit
}

func EqualGPS(a, b GPS) bool {
	return equalCoordinate(a.Latitude, b.Latitude) && equalCoordinate(a.Longitude, b.Longitude)
}
func equalCoordinate(a, b *float64) bool {
	return (a == nil && b == nil) || (a != nil && b != nil && *a == *b)
}

func Compare(snapshot Snapshot, current Metadata) error {
	if snapshot.ImageIdentity != current.ImageIdentity {
		return Failure{Code: "SOURCE_CHANGED"}
	}
	if !EqualGPS(snapshot.Baseline, current.GPS) {
		return Failure{Code: "IMMICH_CONFLICT", Conflict: &Conflict{Before: snapshot.Baseline, Current: current.GPS, Proposed: *snapshot.Camera}}
	}
	return nil
}
