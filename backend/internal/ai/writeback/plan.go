package writeback

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"
	"slices"
	"time"

	"immich-places-backend/internal/ai/writepreview"
)

func DecodePlan(raw []byte, digest, owner, installation, id string) (writepreview.Plan, error) {
	var p writepreview.Plan
	sum := sha256.Sum256(raw)
	if len(raw) > 1<<20 || hex.EncodeToString(sum[:]) != digest || json.Unmarshal(raw, &p) != nil {
		return p, Failure("WRITE_UNAVAILABLE")
	}
	canonical, err := json.Marshal(p)
	created, e1 := time.Parse(time.RFC3339Nano, p.CreatedAt)
	expires, e2 := time.Parse(time.RFC3339Nano, p.ExpiresAt)
	if err != nil || !bytes.Equal(raw, canonical) || p.Owner != owner || p.Installation != installation || p.ID != id || p.DraftID == "" || p.AnalysisID == "" || p.TargetID == "" || p.DraftRevision < 1 || p.ImageIdentity == "" || e1 != nil || e2 != nil || expires.Sub(created) != 5*time.Minute || p.ObservedAt != p.CreatedAt {
		return p, Failure("WRITE_UNAVAILABLE")
	}
	if p.Version != "mirror-preview-v4" && p.Mirror != nil {
		return p, Failure("WRITE_UNAVAILABLE")
	}
	switch p.Version {
	case "gps-preview-v1":
		if p.Manifest != nil || len(raw) > 16<<10 || p.ComparisonPolicy != "exact-nullable-gps-v1" || len(p.Fields) != 1 || p.Fields[0] != "gps" || p.Description != nil || p.PolicyID != "" {
			return p, Failure("WRITE_UNAVAILABLE")
		}
	case "standard-preview-v2":
		if p.Manifest != nil || !validStandardPlan(p) {
			return p, Failure("WRITE_UNAVAILABLE")
		}
	case "stack-preview-v3":
		if !validStackPlan(p) {
			return p, Failure("WRITE_UNAVAILABLE")
		}
	case "mirror-preview-v4":
		if !validMirrorPlan(p) {
			return p, Failure("WRITE_UNAVAILABLE")
		}
	default:
		return p, Failure("WRITE_UNAVAILABLE")
	}
	if slices.Contains(p.Fields, "gps") && (!coordinate(p.Intended.Latitude, 90) || !coordinate(p.Intended.Longitude, 180) || !nullable(p.Before.Latitude, 90) || !nullable(p.Before.Longitude, 180)) {
		return p, Failure("WRITE_UNAVAILABLE")
	}
	return p, nil
}

func validStandardPlan(p writepreview.Plan) bool {
	policy, err := hex.DecodeString(p.PolicyID)
	if err != nil || len(policy) != 32 || p.PolicyID != hex.EncodeToString(policy) || p.ComparisonPolicy != "selected-standard-fields-v2" || !slices.Contains(p.Fields, "description") || len(p.Fields) > 2 || (len(p.Fields) == 2 && p.Fields[0] == p.Fields[1]) {
		return false
	}
	for _, field := range p.Fields {
		if field != "gps" && field != "description" {
			return false
		}
	}
	return p.Description != nil && writepreview.ValidDescriptionPlan(*p.Description)
}

func coordinate(v, limit float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0) && v >= -limit && v <= limit
}
func nullable(v *float64, limit float64) bool { return v == nil || coordinate(*v, limit) }
