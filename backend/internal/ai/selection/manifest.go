package selection

import (
	"encoding/json"
	"time"
)

const PolicyVersion = "selection-v1"

type Exclusion struct {
	AssetID string `json:"assetID"`
	Reason  string `json:"reason"`
}

type QuerySummary struct {
	MatchedCount    int            `json:"matchedCount"`
	ExclusionCounts map[string]int `json:"exclusionCounts"`
}

type Manifest struct {
	*QuerySummary
	ContextPreview *ContextPreview `json:"contextPreview,omitempty"`
	SnapshotID     *string         `json:"snapshotID"`
	Mode           string          `json:"mode"`
	Scope          Scope           `json:"scope"`
	PolicyVersion  string          `json:"policyVersion"`
	AssetIDs       []string        `json:"assetIDs"`
	Exclusions     []Exclusion     `json:"exclusions"`
	RequestedCount int             `json:"requestedCount"`
	UniqueCount    int             `json:"uniqueCount"`
	DuplicateCount int             `json:"duplicateCount"`
	EligibleCount  int             `json:"eligibleCount"`
	ExcludedCount  int             `json:"excludedCount"`
	CreatedAt      time.Time       `json:"createdAt"`
	ExpiresAt      time.Time       `json:"expiresAt"`
}

type ContextPreview struct {
	AlbumLabel   *string           `json:"albumLabel"`
	CaptureTimes map[string]string `json:"captureTimes"`
}

func (m Manifest) Current(now time.Time) bool {
	return m.SnapshotID != nil && m.PolicyVersion == PolicyVersion && now.Before(m.ExpiresAt)
}

// Query responses contain aggregate reasons only. Explicit responses retain the
// original per-ID exclusions, including an empty array for compatible clients.
func (m Manifest) MarshalJSON() ([]byte, error) {
	type plain Manifest
	if m.Mode != "all-matching" {
		return json.Marshal(plain(m))
	}
	return json.Marshal(struct {
		plain
		Exclusions *struct{} `json:"exclusions,omitempty"`
	}{plain: plain(m)})
}
