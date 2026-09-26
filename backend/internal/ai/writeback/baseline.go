package writeback

import (
	"slices"

	"immich-places-backend/internal/ai/writepreview"
)

func CompareBefore(plan writepreview.Plan, fresh writepreview.Metadata) string {
	if fresh.ImageIdentity != plan.ImageIdentity {
		return "SOURCE_CHANGED"
	}
	if slices.Contains(plan.Fields, "gps") && fresh.GPSUnavailable {
		return "SOURCE_UNAVAILABLE"
	}
	if slices.Contains(plan.Fields, "gps") && !writepreview.EqualGPS(fresh.GPS, plan.Before) {
		return "IMMICH_CONFLICT"
	}
	if slices.Contains(plan.Fields, "description") {
		if plan.Description == nil || fresh.Description == nil || !writepreview.ValidTextObservation(*fresh.Description) {
			return "SOURCE_UNAVAILABLE"
		}
		if plan.Description.Before.Value != fresh.Description.Value {
			return "DESCRIPTION_CONFLICT"
		}
	}
	return ""
}
