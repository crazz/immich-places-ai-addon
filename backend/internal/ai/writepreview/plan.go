package writepreview

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"slices"
	"time"
)

type GPS struct {
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
}

type Point struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Snapshot struct {
	Mirror                                              *MirrorInput
	Description                                         *DescriptionInput
	DescriptionBaseline                                 *TextObservation
	PolicyID                                            string
	Owner, Installation, ID, AnalysisID, AssetID, State string
	Revision                                            int
	Camera                                              *Point
	Fields                                              []string
	BaselineReviewed                                    bool
	ImageIdentity                                       string
	Baseline                                            GPS
}

type Metadata struct {
	Mirror         *MirrorBaseline
	GPSUnavailable bool
	Description    *TextObservation
	ImageIdentity  string
	GPS            GPS
}

type Plan struct {
	Manifest         *TargetManifest  `json:"manifest,omitempty"`
	Version          string           `json:"version"`
	ComparisonPolicy string           `json:"comparisonPolicy"`
	ID               string           `json:"id"`
	Owner            string           `json:"owner"`
	Installation     string           `json:"installation"`
	DraftID          string           `json:"draftId"`
	DraftRevision    int              `json:"draftRevision"`
	AnalysisID       string           `json:"analysisId"`
	ImageIdentity    string           `json:"imageIdentity"`
	TargetID         string           `json:"targetId"`
	Fields           []string         `json:"fields"`
	Before           GPS              `json:"before"`
	Intended         Point            `json:"intended"`
	ObservedAt       string           `json:"observedAt"`
	CreatedAt        string           `json:"createdAt"`
	ExpiresAt        string           `json:"expiresAt"`
	Description      *DescriptionPlan `json:"description,omitempty"`
	PolicyID         string           `json:"policyId,omitempty"`
	Mirror           *MirrorPlan      `json:"mirror,omitempty"`
}

type Preview struct {
	Plan   Plan   `json:"plan"`
	Digest string `json:"digest"`
	Status string `json:"status"`
	Diff   string `json:"diff"`
}

func Build(snapshot Snapshot, current Metadata, id string, now time.Time) (Preview, []byte, error) {
	current.GPS = GPS{Latitude: normalizeCoordinate(current.GPS.Latitude), Longitude: normalizeCoordinate(current.GPS.Longitude)}
	plan := Plan{Version: "gps-preview-v1", ComparisonPolicy: "exact-nullable-gps-v1", ID: id, Owner: snapshot.Owner, Installation: snapshot.Installation, DraftID: snapshot.ID, DraftRevision: snapshot.Revision, AnalysisID: snapshot.AnalysisID, ImageIdentity: current.ImageIdentity, TargetID: snapshot.AssetID, Fields: append([]string{}, snapshot.Fields...), Before: current.GPS, ObservedAt: now.UTC().Format(time.RFC3339Nano), CreatedAt: now.UTC().Format(time.RFC3339Nano), ExpiresAt: now.Add(5 * time.Minute).UTC().Format(time.RFC3339Nano)}
	if snapshot.Camera != nil {
		plan.Intended = *snapshot.Camera
	}
	plan.Intended.Latitude = *normalizeCoordinate(&plan.Intended.Latitude)
	plan.Intended.Longitude = *normalizeCoordinate(&plan.Intended.Longitude)
	limit := 16 << 10
	if slices.Contains(snapshot.Fields, "description") {
		if snapshot.Description == nil || current.Description == nil {
			return Preview{}, nil, Failure{Code: "INVALID_PREVIEW"}
		}
		var err error
		plan.Description, err = PlanDescription(*current.Description, *snapshot.Description)
		if err != nil {
			return Preview{}, nil, err
		}
		plan.Version, plan.ComparisonPolicy, plan.PolicyID = "standard-preview-v2", "selected-standard-fields-v2", snapshot.PolicyID
		limit = 1 << 20
	}
	raw, err := json.Marshal(plan)
	if err != nil {
		return Preview{}, nil, err
	}
	if len(raw) > limit {
		return Preview{}, nil, Failure{Code: "INVALID_PREVIEW"}
	}
	sum := sha256.Sum256(raw)
	preview := Preview{Plan: plan, Digest: hex.EncodeToString(sum[:]), Status: "usable", Diff: Diff(plan)}
	if snapshot.Mirror != nil {
		if current.Mirror == nil {
			return Preview{}, nil, Failure{Code: "METADATA_UNAVAILABLE"}
		}
		return AttachMirror(preview, *snapshot.Mirror, *current.Mirror)
	}
	return preview, raw, nil
}

func Diff(plan Plan) string {
	if plan.Mirror != nil && (!plan.Mirror.Before.Present || !EqualMirrorValue(plan.Mirror.Before.Value, plan.Mirror.Value)) {
		return "changed"
	}
	if plan.Manifest != nil {
		for _, target := range plan.Manifest.Targets {
			if (slices.Contains(target.Fields, "gps") && !EqualGPS(target.Before, GPS{Latitude: &target.Intended.Latitude, Longitude: &target.Intended.Longitude})) || (target.Description != nil && target.Description.Before.Value != target.Description.Intended) {
				return "changed"
			}
		}
		return "unchanged"
	}
	if slices.Contains(plan.Fields, "gps") && !EqualGPS(plan.Before, GPS{Latitude: &plan.Intended.Latitude, Longitude: &plan.Intended.Longitude}) {
		return "changed"
	}
	if plan.Description != nil && plan.Description.Before.Value != plan.Description.Intended {
		return "changed"
	}
	return "unchanged"
}

func normalizeCoordinate(value *float64) *float64 {
	if value == nil {
		return nil
	}
	normalized := *value
	if normalized == 0 {
		normalized = 0
	}
	return &normalized
}
