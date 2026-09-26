package immichwrite

import (
	"context"
	"encoding/json"
	"math"

	"github.com/google/uuid"
	"immich-places-backend/internal/ai/writepreview"
)

func (t *Transport) SendStandard(ctx context.Context, key string, plan writepreview.Plan) Outcome {
	if _, err := uuid.Parse(plan.TargetID); err != nil || (plan.Version != "standard-preview-v2" && plan.Version != "stack-preview-v3") || plan.Manifest != nil || plan.Description == nil || !writepreview.ValidDescriptionPlan(*plan.Description) || len(plan.Fields) < 1 || len(plan.Fields) > 2 {
		return Outcome{CompletionKnown: true, Code: "not_sent"}
	}
	var payload struct {
		Latitude    *float64 `json:"latitude,omitempty"`
		Longitude   *float64 `json:"longitude,omitempty"`
		Description *string  `json:"description,omitempty"`
	}
	for _, field := range plan.Fields {
		switch field {
		case "description":
			if payload.Description != nil {
				return Outcome{CompletionKnown: true, Code: "not_sent"}
			}
			payload.Description = &plan.Description.Intended
		case "gps":
			if payload.Latitude != nil || math.IsNaN(plan.Intended.Latitude) || math.IsNaN(plan.Intended.Longitude) || math.Abs(plan.Intended.Latitude) > 90 || math.Abs(plan.Intended.Longitude) > 180 {
				return Outcome{CompletionKnown: true, Code: "not_sent"}
			}
			payload.Latitude, payload.Longitude = &plan.Intended.Latitude, &plan.Intended.Longitude
		default:
			return Outcome{CompletionKnown: true, Code: "not_sent"}
		}
	}
	if payload.Description == nil {
		return Outcome{CompletionKnown: true, Code: "not_sent"}
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return Outcome{CompletionKnown: true, Code: "not_sent"}
	}
	return t.send(ctx, key, plan.TargetID, raw)
}
