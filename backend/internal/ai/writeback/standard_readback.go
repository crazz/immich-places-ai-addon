package writeback

import "immich-places-backend/internal/ai/writepreview"

func standardReadback(op Operation, fresh writepreview.Metadata, completed bool) Decision {
	decision := Decision{}
	verified := 0
	for _, field := range op.Plan.Fields {
		outcome := standardFieldReadback(op.Plan, fresh, field)
		outcome.WasVerified = outcome.Status == "verified"
		for _, previous := range op.Fields {
			if previous.Field == field && (previous.WasVerified || previous.Status == "verified") {
				outcome.WasVerified = true
				if outcome.Status == "baseline" {
					outcome.Status = "conflict"
				}
			}
		}
		decision.Fields = append(decision.Fields, outcome)
		if outcome.Status == "verified" {
			verified++
			decision.GPSVerified = decision.GPSVerified || field == "gps"
		}
	}
	if fresh.ImageIdentity != op.Plan.ImageIdentity {
		decision.Status, decision.Code = "conflict", "SOURCE_CHANGED"
		return decision
	}
	for _, field := range decision.Fields {
		if field.Status == "unavailable" {
			decision.Status, decision.Code = "verifying", "READBACK_UNAVAILABLE"
			return decision
		}
	}
	if verified == len(op.Plan.Fields) && verified > 0 {
		decision.Status, decision.Code, decision.Verified = "succeeded", "STANDARD_FIELDS_VERIFIED", true
		return decision
	}
	if verified > 0 {
		decision.Status, decision.Code = "partial", "PARTIAL_FIELDS"
		if !completed {
			decision.Status, decision.Code = "verifying", "PARTIAL_UNRESOLVED"
		}
		return decision
	}
	decision.Status, decision.Code = "retryable", "RETRY_AVAILABLE"
	for _, field := range decision.Fields {
		if field.Status == "conflict" {
			decision.Status, decision.Code = "conflict", "IMMICH_CONFLICT"
			return decision
		}
	}
	if !completed {
		decision.Status, decision.Code = "verifying", "RECONCILIATION_REQUIRED"
	} else if op.Attempts >= 2 {
		decision.Status, decision.Code = "failed", "ATTEMPTS_EXHAUSTED"
	}
	return decision
}

func standardFieldReadback(plan writepreview.Plan, fresh writepreview.Metadata, field string) FieldOutcome {
	outcome := FieldOutcome{Field: field, Status: "conflict"}
	if fresh.ImageIdentity != plan.ImageIdentity {
		return outcome
	}
	switch field {
	case "gps":
		if fresh.GPSUnavailable {
			outcome.Status = "unavailable"
			return outcome
		}
		outcome.GPS = &fresh.GPS
		if match(fresh.GPS.Latitude, plan.Intended.Latitude) && match(fresh.GPS.Longitude, plan.Intended.Longitude) {
			outcome.Status = "verified"
		} else if writepreview.EqualGPS(fresh.GPS, plan.Before) {
			outcome.Status = "baseline"
		}
	case "description":
		if plan.Description == nil || fresh.Description == nil || !writepreview.ValidTextObservation(*fresh.Description) {
			outcome.Status = "unavailable"
			return outcome
		}
		outcome.Description = fresh.Description
		if fresh.Description.Value == plan.Description.Intended {
			outcome.Status = "verified"
		} else if fresh.Description.Value == plan.Description.Before.Value {
			outcome.Status = "baseline"
		}
	}
	return outcome
}
