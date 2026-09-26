package drafts

import (
	"encoding/json"
	"testing"
)

func TestMirrorSelectionIsExplicitRevisionedAndRemovable(t *testing.T) {
	current := Draft{State: "staged", Revision: 2, FactsRevision: 1, Camera: &Point{}, Fields: []string{"gps"}}
	var edit Edit
	if err := json.Unmarshal([]byte(`{"mirror":{"direction":true,"languages":["en"]}}`), &edit); err != nil {
		t.Fatal(err)
	}
	next, err := Apply(current, edit, nil)
	if err != nil || next.Mirror == nil || !next.Mirror.Direction || len(next.Mirror.Languages) != 1 || next.Revision != 3 || next.State != "draft" || len(next.Fields) != 1 || next.Fields[0] != "gps" || current.Mirror != nil {
		t.Fatalf("mirror selection not independently revisioned: %+v %v", next, err)
	}
	next.State = "staged"
	var remove Edit
	if err := json.Unmarshal([]byte(`{"mirror":null}`), &remove); err != nil {
		t.Fatal(err)
	}
	removed, err := Apply(next, remove, nil)
	if err != nil || removed.Mirror != nil || removed.Revision != 4 || removed.State != "draft" {
		t.Fatalf("mirror deselection kept old approval usable: %+v %v", removed, err)
	}
	for _, invalid := range []string{`{}`, `{"direction":true,"assetId":"sibling"}`, `{"direction":true,"direction":false}`, `{"languages":["en","en"]}`, `{"model":true}`, `[]`} {
		if _, err := Apply(current, Edit{Mirror: json.RawMessage(invalid)}, nil); err == nil {
			t.Fatalf("invalid mirror selection accepted: %s", invalid)
		}
	}
}

func TestReviewPrecisionIndependentlyWithoutRevivingOtherStaleValues(t *testing.T) {
	radius, heading := 200.0, 40.0
	current := Draft{State: "staged", Revision: 4, FactsRevision: 2, Camera: &Point{}, Fields: []string{"gps"}, Radius: &radius, RadiusStale: true, Heading: &heading, HeadingStale: true, Descriptions: []Description{{Language: "en", Stale: true}}}
	var edit Edit
	if err := json.Unmarshal([]byte(`{"reviewPrecision":true}`), &edit); err != nil {
		t.Fatal(err)
	}
	next, err := Apply(current, edit, nil)
	if err != nil || next.RadiusStale || !next.HeadingStale || !next.Descriptions[0].Stale || next.State != "draft" || next.Revision != 5 || *next.Radius != radius {
		t.Fatalf("precision review expanded or lost scope: %+v %v", next, err)
	}
}
