package writeback

import (
	"immich-places-backend/internal/ai/writepreview"
	"math"
)

type Decision struct {
	Status, Code string
	Verified     bool
}

func Readback(op Operation, fresh writepreview.Metadata, completed bool) Decision {
	if fresh.ImageIdentity != op.Plan.ImageIdentity {
		return Decision{Status: "conflict", Code: "SOURCE_CHANGED"}
	}
	if match(fresh.GPS.Latitude, op.Plan.Intended.Latitude) && match(fresh.GPS.Longitude, op.Plan.Intended.Longitude) {
		return Decision{Status: "succeeded", Code: "GPS_VERIFIED", Verified: true}
	}
	if !writepreview.EqualGPS(fresh.GPS, op.Plan.Before) {
		return Decision{Status: "conflict", Code: "IMMICH_CONFLICT"}
	}
	if completed {
		if op.Attempts >= 2 {
			return Decision{Status: "failed", Code: "ATTEMPTS_EXHAUSTED"}
		}
		return Decision{Status: "retryable", Code: "RETRY_AVAILABLE"}
	}
	return Decision{Status: "verifying", Code: "RECONCILIATION_REQUIRED"}
}
func match(actual *float64, intended float64) bool {
	return actual != nil && coordinate(*actual, 180) && math.Abs(*actual-intended) <= 1e-7
}
