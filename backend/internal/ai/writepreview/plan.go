package writepreview

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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
	Owner, Installation, ID, AnalysisID, AssetID, State string
	Revision                                            int
	Camera                                              *Point
	Fields                                              []string
	BaselineReviewed                                    bool
	ImageIdentity                                       string
	Baseline                                            GPS
}

type Metadata struct {
	ImageIdentity string
	GPS           GPS
}

type Plan struct {
	Version          string   `json:"version"`
	ComparisonPolicy string   `json:"comparisonPolicy"`
	ID               string   `json:"id"`
	Owner            string   `json:"owner"`
	Installation     string   `json:"installation"`
	DraftID          string   `json:"draftId"`
	DraftRevision    int      `json:"draftRevision"`
	AnalysisID       string   `json:"analysisId"`
	ImageIdentity    string   `json:"imageIdentity"`
	TargetID         string   `json:"targetId"`
	Fields           []string `json:"fields"`
	Before           GPS      `json:"before"`
	Intended         Point    `json:"intended"`
	ObservedAt       string   `json:"observedAt"`
	CreatedAt        string   `json:"createdAt"`
	ExpiresAt        string   `json:"expiresAt"`
}

type Preview struct {
	Plan   Plan   `json:"plan"`
	Digest string `json:"digest"`
	Status string `json:"status"`
	Diff   string `json:"diff"`
}

func Build(snapshot Snapshot, current Metadata, id string, now time.Time) (Preview, []byte, error) {
	current.GPS = GPS{Latitude: normalizeCoordinate(current.GPS.Latitude), Longitude: normalizeCoordinate(current.GPS.Longitude)}
	plan := Plan{Version: "gps-preview-v1", ComparisonPolicy: "exact-nullable-gps-v1", ID: id, Owner: snapshot.Owner, Installation: snapshot.Installation, DraftID: snapshot.ID, DraftRevision: snapshot.Revision, AnalysisID: snapshot.AnalysisID, ImageIdentity: current.ImageIdentity, TargetID: snapshot.AssetID, Fields: []string{"gps"}, Before: current.GPS, Intended: *snapshot.Camera, ObservedAt: now.UTC().Format(time.RFC3339Nano), CreatedAt: now.UTC().Format(time.RFC3339Nano), ExpiresAt: now.Add(5 * time.Minute).UTC().Format(time.RFC3339Nano)}
	plan.Intended.Latitude = *normalizeCoordinate(&plan.Intended.Latitude)
	plan.Intended.Longitude = *normalizeCoordinate(&plan.Intended.Longitude)
	raw, err := json.Marshal(plan)
	if err != nil {
		return Preview{}, nil, err
	}
	if len(raw) > 16<<10 {
		return Preview{}, nil, Failure{Code: "INVALID_PREVIEW"}
	}
	sum := sha256.Sum256(raw)
	return Preview{Plan: plan, Digest: hex.EncodeToString(sum[:]), Status: "usable", Diff: Diff(plan)}, raw, nil
}

func Diff(plan Plan) string {
	if EqualGPS(plan.Before, GPS{Latitude: &plan.Intended.Latitude, Longitude: &plan.Intended.Longitude}) {
		return "unchanged"
	}
	return "changed"
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
