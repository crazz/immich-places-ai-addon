package writeback

import (
	"encoding/hex"
	"reflect"
	"slices"

	"github.com/google/uuid"
	"immich-places-backend/internal/ai/writepreview"
)

func validMirrorPlan(p writepreview.Plan) bool {
	policy, err := hex.DecodeString(p.PolicyID)
	if err != nil || len(policy) != 32 || hex.EncodeToString(policy) != p.PolicyID || p.ComparisonPolicy != "standard-then-metadata-v4" || p.Mirror == nil || !writepreview.ValidMirrorPlan(*p.Mirror, p.DraftRevision) || p.Manifest == nil || len(p.Manifest.Targets) < 1 {
		return false
	}
	standard := p
	standard.Mirror = nil
	if len(p.Manifest.Targets) > 1 {
		standard.Version, standard.ComparisonPolicy = "stack-preview-v3", "selected-stack-targets-v3"
		return validStackPlan(standard)
	}
	if p.Manifest.StackID != "" || p.Manifest.ReviewID != "" {
		return false
	}
	if id, err := uuid.Parse(p.TargetID); err != nil || id.String() != p.TargetID {
		return false
	}
	if slices.Contains(p.Fields, "description") {
		standard.ComparisonPolicy = "selected-standard-fields-v2"
		if !validStandardPlan(standard) {
			return false
		}
	} else if len(p.Fields) != 1 || p.Fields[0] != "gps" || p.Description != nil {
		return false
	}
	target := p.Manifest.Targets[0]
	return target.AssetID == p.TargetID && target.ImageIdentity == p.ImageIdentity && slices.Equal(target.Fields, p.Fields) && writepreview.EqualGPS(target.Before, p.Before) && target.Intended == p.Intended && reflect.DeepEqual(target.Description, p.Description)
}
