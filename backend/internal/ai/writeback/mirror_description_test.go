package writeback

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/results"
	"immich-places-backend/internal/ai/writepreview"
)

func TestDescriptionOnlyMirrorRoundtripIgnoresUnselectedGPS(t *testing.T) {
	base, _ := stackPlanFixture(t)
	base.Plan.Manifest = nil
	base.Plan.Fields = []string{"description"}
	base.Plan.Description = &writepreview.DescriptionPlan{Language: "en", Policy: "replace", Before: writepreview.TextObservation{Presence: "value", Value: "same"}, Intended: "same"}
	base.Plan.Intended = writepreview.Point{Latitude: 21, Longitude: 32}
	lat, lon := 1.0, 2.0
	base.Plan.Before = writepreview.GPS{Latitude: &lat, Longitude: &lon}
	text := "same"
	choice := drafts.MirrorSelection{Languages: []string{"en"}}
	d := drafts.Draft{Revision: base.Plan.DraftRevision, FactsRevision: 1, Descriptions: []drafts.Description{{Language: "en", Status: "complete", Text: &text, FactsRevision: 1}}}
	const record = "aaf12459-abcd-4321-8421-aaccff110033"
	exported, err := writepreview.BuildMirrorExport(d, results.Document{}, choice, writepreview.MirrorProvenance{}, record)
	if err != nil {
		t.Fatal(err)
	}
	preview, raw, err := writepreview.AttachMirror(base, writepreview.MirrorInput{RecordID: record, Export: exported, Owned: exported, Selection: choice, PolicyID: strings.Repeat("a", 64), Disclosure: writepreview.MirrorDisclosure}, writepreview.MirrorBaseline{Present: true, Value: json.RawMessage(exported)})
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodePlan(raw, preview.Digest, preview.Plan.Owner, preview.Plan.Installation, preview.Plan.ID)
	if err != nil || !reflect.DeepEqual(decoded, preview.Plan) {
		t.Fatal("description-only mirror cannot reload", err)
	}
	if preview.Diff != "unchanged" {
		t.Fatal("unselected GPS affected comparison", preview.Diff)
	}
}
