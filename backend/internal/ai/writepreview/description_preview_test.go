package writepreview

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestDescriptionPreviewDoesNotRequireOrApproveGPS(t *testing.T) {
	before := TextObservation{Presence: "null"}
	snapshot := Snapshot{Owner: "owner", Installation: "installation", ID: "draft", Revision: 3, AnalysisID: "analysis", AssetID: "photo", State: "staged", Fields: []string{"description"}, BaselineReviewed: true, ImageIdentity: "image", DescriptionBaseline: &before, Description: &DescriptionInput{Text: "A scene without known location.", Language: "en", Policy: "replace"}, PolicyID: strings.Repeat("a", 64)}
	if err := Validate(snapshot, 3); err != nil {
		t.Fatal("description-only preview rejected", err)
	}
	lat, lon := 47.0, 18.0
	current := Metadata{ImageIdentity: "image", GPS: GPS{Latitude: &lat, Longitude: &lon}, Description: &before}
	if err := Compare(snapshot, current); err != nil {
		t.Fatal("unselected GPS conflicted", err)
	}
	preview, raw, err := Build(snapshot, current, "preview", time.Unix(1000, 0))
	if err != nil || preview.Plan.Version != "standard-preview-v2" || preview.Plan.Description == nil || preview.Plan.Description.Intended != snapshot.Description.Text || len(preview.Plan.Fields) != 1 || preview.Plan.Fields[0] != "description" {
		t.Fatalf("bad description preview: %+v %v", preview, err)
	}
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil || fields["before"] != nil || fields["intended"] != nil {
		t.Fatal("unselected GPS was serialized", string(raw))
	}
}
