package writepreview

import (
	"math"
	"strings"
	"testing"
	"time"
)

func TestCanonicalGPSPlanNormalizesValuesAndBoundsPayload(t *testing.T) {
	zero := math.Copysign(0, -1)
	snapshot := Snapshot{Owner: "owner", Installation: "installation", ID: "draft", AnalysisID: "analysis", AssetID: "asset", Revision: 2, State: "staged", Fields: []string{"gps"}, Camera: &Point{Latitude: zero, Longitude: 0}, BaselineReviewed: true, ImageIdentity: "v1:image", Baseline: GPS{Latitude: &zero}}
	current := Metadata{ImageIdentity: snapshot.ImageIdentity, GPS: snapshot.Baseline}
	now := time.Date(2026, 9, 25, 0, 0, 0, 0, time.UTC)
	first, raw, err := Build(snapshot, current, "preview", now)
	if err != nil || strings.Contains(string(raw), ":-0") {
		t.Fatal("noncanonical zero", string(raw), err)
	}
	second, again, err := Build(snapshot, current, "preview", now)
	if err != nil || string(again) != string(raw) || second.Digest != first.Digest {
		t.Fatal("unstable canonical encoding")
	}
	if !math.Signbit(*current.GPS.Latitude) {
		t.Fatal("canonicalization mutated its input")
	}
	snapshot.Owner = strings.Repeat("x", 16<<10)
	if _, _, err = Build(snapshot, current, "preview", now); err == nil {
		t.Fatal("oversized canonical plan accepted")
	}
}
