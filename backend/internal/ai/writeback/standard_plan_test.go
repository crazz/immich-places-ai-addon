package writeback

import (
	"strings"
	"testing"
	"time"

	"immich-places-backend/internal/ai/writepreview"
)

func TestDecodeStandardPlanRetainsExactDescriptionAuthority(t *testing.T) {
	before := writepreview.TextObservation{Presence: "value", Value: "Original\r\n"}
	snapshot := writepreview.Snapshot{Owner: "owner", Installation: "installation", ID: "draft", Revision: 3, AnalysisID: "analysis", AssetID: "photo", State: "staged", Fields: []string{"description"}, BaselineReviewed: true, ImageIdentity: "image", DescriptionBaseline: &before, Description: &writepreview.DescriptionInput{Text: "A scene without known location.", Language: "en", Policy: "replace"}, PolicyID: strings.Repeat("a", 64)}
	preview, raw, err := writepreview.Build(snapshot, writepreview.Metadata{ImageIdentity: "image", Description: &before}, "preview", time.Unix(1000, 0))
	if err != nil {
		t.Fatal(err)
	}
	plan, err := DecodePlan(raw, preview.Digest, "owner", "installation", "preview")
	if err != nil || plan.Description == nil || plan.Description.Intended != snapshot.Description.Text || plan.Description.Before != before || len(plan.Fields) != 1 || plan.Fields[0] != "description" {
		t.Fatalf("new plan lost authority: %+v %v", plan, err)
	}
}
