package writepreview

import (
	"encoding/json"
	"strings"
	"testing"

	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/results"
)

func TestMirrorRetainedHeadingKeepsItsOwnMethodAndUncertainty(t *testing.T) {
	a, b := "first", "second"
	uncertaintyA, uncertaintyB := json.Number("10"), json.Number("70")
	doc := results.Document{SelectedCandidateID: &a, Candidates: []results.Candidate{
		{ID: a, CameraLocation: &results.CameraLocation{Coordinate: results.Coordinate{Latitude: "0", Longitude: "0"}}, CameraDirection: &results.Direction{AzimuthDeg: "90", Method: "visual_estimate", UncertaintyDeg: &uncertaintyA}},
		{ID: b, CameraLocation: &results.CameraLocation{Coordinate: results.Coordinate{Latitude: "10", Longitude: "20"}}, CameraDirection: &results.Direction{AzimuthDeg: "90", Method: "known_viewpoint_alignment", UncertaintyDeg: &uncertaintyB}},
	}}
	d, err := drafts.FromProposal(drafts.Draft{Revision: 1}, doc, nil)
	if err != nil {
		t.Fatal(err)
	}
	d, err = drafts.Apply(d, drafts.Edit{CandidateID: &b}, &doc)
	if err != nil {
		t.Fatal(err)
	}
	d, err = drafts.Apply(d, drafts.Edit{ReviewHeading: true}, nil)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := BuildMirrorExport(d, doc, drafts.MirrorSelection{Direction: true}, MirrorProvenance{}, "aaf12459-abcd-4321-8421-aaccff110033")
	if err != nil || !strings.Contains(string(raw), `"method":"visual_estimate","uncertainty":10`) || strings.Contains(string(raw), "known_viewpoint_alignment") {
		t.Fatalf("retained heading borrowed new candidate evidence: %s %v", raw, err)
	}
}
