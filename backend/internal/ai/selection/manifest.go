package selection

import "time"

const PolicyVersion = "selection-v1"

type Exclusion struct {
	AssetID string `json:"assetID"`
	Reason  string `json:"reason"`
}

type Manifest struct {
	SnapshotID     *string     `json:"snapshotID"`
	Mode           string      `json:"mode"`
	Scope          Scope       `json:"scope"`
	PolicyVersion  string      `json:"policyVersion"`
	AssetIDs       []string    `json:"assetIDs"`
	Exclusions     []Exclusion `json:"exclusions"`
	RequestedCount int         `json:"requestedCount"`
	UniqueCount    int         `json:"uniqueCount"`
	DuplicateCount int         `json:"duplicateCount"`
	EligibleCount  int         `json:"eligibleCount"`
	ExcludedCount  int         `json:"excludedCount"`
	CreatedAt      time.Time   `json:"createdAt"`
	ExpiresAt      time.Time   `json:"expiresAt"`
}

func (m Manifest) Current(now time.Time) bool {
	return m.SnapshotID != nil && m.PolicyVersion == PolicyVersion && now.Before(m.ExpiresAt)
}
