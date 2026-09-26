package writepreview

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"slices"
	"time"
)

type ReviewedTarget struct {
	AssetID       string `json:"assetId"`
	ImageIdentity string `json:"imageIdentity"`
	Before        GPS    `json:"before"`
}

type StackReview struct {
	Candidates    []string         `json:"candidates"`
	ID            string           `json:"id"`
	Owner         string           `json:"owner"`
	Installation  string           `json:"installation"`
	DraftID       string           `json:"draftId"`
	DraftRevision int              `json:"draftRevision"`
	AnalyzedID    string           `json:"analyzedId"`
	ImageIdentity string           `json:"imageIdentity"`
	StackID       string           `json:"stackId"`
	ExpiresAt     string           `json:"expiresAt"`
	Targets       []ReviewedTarget `json:"targets"`
}

type Target struct {
	AssetID       string           `json:"assetId"`
	ImageIdentity string           `json:"imageIdentity"`
	Fields        []string         `json:"fields"`
	Before        GPS              `json:"before"`
	Intended      Point            `json:"intended"`
	Description   *DescriptionPlan `json:"description,omitempty"`
}

type TargetManifest struct {
	StackID  string   `json:"stackId"`
	ReviewID string   `json:"reviewId"`
	Targets  []Target `json:"targets"`
}

func BuildStack(snapshot Snapshot, review StackReview, current map[string]Metadata, id string, now time.Time) (Preview, []byte, error) {
	if err := validateStackReview(snapshot, review, current, now); err != nil {
		return Preview{}, nil, err
	}
	standard := snapshot
	standard.Mirror = nil
	preview, _, err := Build(standard, current[snapshot.AssetID], id, now)
	if err != nil {
		return Preview{}, nil, err
	}
	plan := &preview.Plan
	plan.Version, plan.ComparisonPolicy, plan.PolicyID = "stack-preview-v3", "selected-stack-targets-v3", snapshot.PolicyID
	plan.Manifest = &TargetManifest{StackID: review.StackID, ReviewID: review.ID, Targets: make([]Target, 0, len(review.Targets))}
	for _, item := range review.Targets {
		target := Target{AssetID: item.AssetID, ImageIdentity: item.ImageIdentity, Before: GPS{Latitude: normalizeCoordinate(item.Before.Latitude), Longitude: normalizeCoordinate(item.Before.Longitude)}, Intended: plan.Intended, Fields: []string{"gps"}}
		if item.AssetID == snapshot.AssetID {
			target.Fields = slices.Clone(plan.Fields)
			target.Description = plan.Description
		}
		plan.Manifest.Targets = append(plan.Manifest.Targets, target)
	}
	raw, err := json.Marshal(plan)
	if err != nil {
		return Preview{}, nil, err
	}
	if len(raw) > 1<<20 {
		return Preview{}, nil, Failure{Code: "INVALID_PREVIEW"}
	}
	sum := sha256.Sum256(raw)
	preview.Digest, preview.Diff = hex.EncodeToString(sum[:]), Diff(*plan)
	if snapshot.Mirror != nil {
		if current[snapshot.AssetID].Mirror == nil {
			return Preview{}, nil, Failure{Code: "METADATA_UNAVAILABLE"}
		}
		return AttachMirror(preview, *snapshot.Mirror, *current[snapshot.AssetID].Mirror)
	}
	return preview, raw, nil
}
