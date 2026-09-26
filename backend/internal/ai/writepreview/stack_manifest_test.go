package writepreview

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestBuildStackManifestFreezesExactMemberFieldMatrix(t *testing.T) {
	now := time.Unix(1000, 0)
	primary, sibling := "00000000-0000-4000-8000-000000000001", "00000000-0000-4000-8000-000000000002"
	lat, lon := 0.0, 12.0
	text := TextObservation{Presence: "value", Value: "exact\r\né "}
	baseline := GPS{Latitude: &lat, Longitude: &lon}
	snapshot := Snapshot{Owner: "owner", Installation: "installation", ID: "draft", AnalysisID: "analysis", AssetID: primary, Revision: 2, State: "staged", Fields: []string{"gps", "description"}, Camera: &Point{Latitude: lat, Longitude: lon}, BaselineReviewed: true, ImageIdentity: "image1", Baseline: baseline, DescriptionBaseline: &text, Description: &DescriptionInput{Text: text.Value, Language: "en", Policy: "replace"}, PolicyID: strings.Repeat("a", 64)}
	review := StackReview{ID: "00000000-0000-4000-8000-000000000004", Owner: snapshot.Owner, Installation: snapshot.Installation, DraftID: snapshot.ID, DraftRevision: 2, AnalyzedID: primary, ImageIdentity: "image1", StackID: "00000000-0000-4000-8000-000000000003", ExpiresAt: now.Add(time.Minute).Format(time.RFC3339Nano), Targets: []ReviewedTarget{{AssetID: primary, ImageIdentity: "image1", Before: baseline}, {AssetID: sibling, ImageIdentity: "image2", Before: GPS{}}}}
	current := map[string]Metadata{primary: {ImageIdentity: "image1", GPS: baseline, Description: &text}, sibling: {ImageIdentity: "image2", GPS: GPS{}, Description: &TextObservation{Presence: "value", Value: "sibling text"}}}
	preview, raw, err := BuildStack(snapshot, review, current, "preview", now)
	if err != nil {
		t.Fatal(err)
	}
	if preview.Plan.Version != "stack-preview-v3" || preview.Plan.Manifest == nil || len(preview.Plan.Manifest.Targets) != 2 || preview.Diff != "changed" || Diff(preview.Plan) != "changed" {
		t.Fatalf("bad stack manifest: %+v", preview)
	}
	targets := preview.Plan.Manifest.Targets
	if targets[0].Description == nil || targets[0].Description.Intended != text.Value || targets[1].Description != nil || len(targets[1].Fields) != 1 || targets[1].Fields[0] != "gps" || targets[1].Intended != *snapshot.Camera || targets[1].Before.Latitude != nil {
		t.Fatalf("expanded or changed fields: %+v", targets)
	}
	again, rawAgain, err := BuildStack(snapshot, review, current, "preview", now)
	if err != nil || !bytes.Equal(raw, rawAgain) || again.Digest != preview.Digest || bytes.Contains(raw, []byte("sibling text")) {
		t.Fatal("unstable or excessive manifest", err)
	}
	review.Targets[1].ImageIdentity = "later"
	snapshot.Fields[0] = "changed"
	if targets[1].ImageIdentity != "image2" || targets[0].Fields[0] != "gps" {
		t.Fatal("manifest aliases mutable inputs")
	}
}
