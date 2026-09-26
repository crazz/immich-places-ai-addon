package writeback

import (
	"encoding/hex"
	"reflect"
	"slices"

	"github.com/google/uuid"
	"immich-places-backend/internal/ai/writepreview"
)

func validStackPlan(p writepreview.Plan) bool {
	policy, err := hex.DecodeString(p.PolicyID)
	if err != nil || len(policy) != 32 || p.PolicyID != hex.EncodeToString(policy) || p.ComparisonPolicy != "selected-stack-targets-v3" || p.Manifest == nil || len(p.Manifest.Targets) < 2 || len(p.Manifest.Targets) > writepreview.MaxStackTargets || !slices.Contains(p.Fields, "gps") {
		return false
	}
	if len(p.Fields) == 1 {
		if p.Description != nil {
			return false
		}
	} else {
		standard := p
		standard.ComparisonPolicy = "selected-standard-fields-v2"
		if !validStandardPlan(standard) {
			return false
		}
	}
	for _, id := range []string{p.Manifest.StackID, p.Manifest.ReviewID} {
		if parsed, err := uuid.Parse(id); err != nil || parsed.String() != id {
			return false
		}
	}
	ids := make([]string, 0, len(p.Manifest.Targets))
	for _, target := range p.Manifest.Targets {
		ids = append(ids, target.AssetID)
		if target.ImageIdentity == "" || !nullable(target.Before.Latitude, 90) || !nullable(target.Before.Longitude, 180) || target.Intended != p.Intended {
			return false
		}
		if target.AssetID == p.TargetID {
			if target.ImageIdentity != p.ImageIdentity || !slices.Equal(target.Fields, p.Fields) || !writepreview.EqualGPS(target.Before, p.Before) || !reflect.DeepEqual(target.Description, p.Description) {
				return false
			}
		} else if len(target.Fields) != 1 || target.Fields[0] != "gps" || target.Description != nil {
			return false
		}
	}
	selected, err := writepreview.SelectStackTargets(writepreview.Snapshot{AssetID: p.TargetID, Fields: p.Fields, Camera: &p.Intended}, ids, ids)
	return err == nil && slices.Equal(selected, ids)
}
