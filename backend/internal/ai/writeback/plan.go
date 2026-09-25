package writeback

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"
	"time"

	"immich-places-backend/internal/ai/writepreview"
)

func DecodePlan(raw []byte, digest, owner, installation, id string) (writepreview.Plan, error) {
	var p writepreview.Plan
	sum := sha256.Sum256(raw)
	if len(raw) > 16<<10 || hex.EncodeToString(sum[:]) != digest || json.Unmarshal(raw, &p) != nil {
		return p, Failure("WRITE_UNAVAILABLE")
	}
	canonical, err := json.Marshal(p)
	created, e1 := time.Parse(time.RFC3339Nano, p.CreatedAt)
	expires, e2 := time.Parse(time.RFC3339Nano, p.ExpiresAt)
	if err != nil || !bytes.Equal(raw, canonical) || p.Owner != owner || p.Installation != installation || p.ID != id || p.Version != "gps-preview-v1" || p.ComparisonPolicy != "exact-nullable-gps-v1" || p.DraftID == "" || p.AnalysisID == "" || p.TargetID == "" || p.DraftRevision < 1 || len(p.Fields) != 1 || p.Fields[0] != "gps" || p.ImageIdentity == "" || !coordinate(p.Intended.Latitude, 90) || !coordinate(p.Intended.Longitude, 180) || !nullable(p.Before.Latitude, 90) || !nullable(p.Before.Longitude, 180) || e1 != nil || e2 != nil || expires.Sub(created) != 5*time.Minute || p.ObservedAt != p.CreatedAt {
		return p, Failure("WRITE_UNAVAILABLE")
	}
	return p, nil
}

func coordinate(v, limit float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0) && v >= -limit && v <= limit
}
func nullable(v *float64, limit float64) bool { return v == nil || coordinate(*v, limit) }
