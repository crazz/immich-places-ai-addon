package writepreview

import (
	"math"
	"slices"
)

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
	if len(snapshot.Fields) < 1 || len(snapshot.Fields) > 2 || (len(snapshot.Fields) == 2 && snapshot.Fields[0] == snapshot.Fields[1]) {
		return Failure{Code: "INVALID_PREVIEW"}
	}
	for _, field := range snapshot.Fields {
		if field != "gps" && field != "description" {
			return Failure{Code: "INVALID_PREVIEW"}
		}
	}
	if slices.Contains(snapshot.Fields, "gps") && (snapshot.Camera == nil || !finite(snapshot.Camera.Latitude, 90) || !finite(snapshot.Camera.Longitude, 180)) {
		return Failure{Code: "INVALID_PREVIEW"}
	}
	if slices.Contains(snapshot.Fields, "description") {
		if snapshot.Description == nil || len(snapshot.PolicyID) != 64 {
			return Failure{Code: "INVALID_PREVIEW"}
		}
		if snapshot.DescriptionBaseline == nil || !ValidTextObservation(*snapshot.DescriptionBaseline) {
			return Failure{Code: "BASELINE_REVIEW_REQUIRED"}
		}
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
	if slices.Contains(snapshot.Fields, "gps") && !EqualGPS(snapshot.Baseline, current.GPS) {
		return Failure{Code: "IMMICH_CONFLICT", Conflict: &Conflict{Before: snapshot.Baseline, Current: current.GPS, Proposed: *snapshot.Camera}}
	}
	if slices.Contains(snapshot.Fields, "description") {
		if snapshot.DescriptionBaseline == nil || current.Description == nil || !ValidTextObservation(*current.Description) {
			return Failure{Code: "SOURCE_UNAVAILABLE"}
		}
		if snapshot.DescriptionBaseline.Value != current.Description.Value {
			return Failure{Code: "DESCRIPTION_CONFLICT"}
		}
	}
	return nil
}
