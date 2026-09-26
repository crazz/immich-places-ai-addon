package drafts

import (
	"encoding/json"
	"testing"
)

func TestStageReviewedDescriptionWithoutCamera(t *testing.T) {
	text := "A bridge, location unknown."
	current := Draft{State: "draft", Revision: 1, FactsRevision: 1, Fields: []string{}, Descriptions: []Description{{Language: "en", Status: "complete", Text: &text, Basis: "scene_only", FactsRevision: 1}}}
	var edit Edit
	if err := json.Unmarshal([]byte(`{"fields":["description"],"primaryLanguage":"en","descriptionPolicy":"replace","state":"staged"}`), &edit); err != nil {
		t.Fatal(err)
	}
	next, err := Apply(current, edit, nil)
	if err != nil || next.State != "staged" || next.Camera != nil || next.Revision != 2 || len(next.Fields) != 1 || next.Fields[0] != "description" {
		t.Fatalf("description-only staging failed: %+v, %v", next, err)
	}
	raw, err := json.Marshal(next)
	var saved map[string]any
	if err != nil || json.Unmarshal(raw, &saved) != nil || saved["primaryLanguage"] != "en" || saved["descriptionPolicy"] != "replace" {
		t.Fatalf("selected language/policy not retained: %s", raw)
	}
}

func TestDescriptionSelectionRejectsUnreadyTextAndScope(t *testing.T) {
	text, empty := "Exact text", ""
	base := Draft{Fields: []string{"description"}, PrimaryLanguage: "en", DescriptionPolicy: "replace", FactsRevision: 2, Descriptions: []Description{{Language: "en", Status: "complete", Text: &text, Basis: "candidate", FactsRevision: 2}}}
	for _, change := range []func(*Draft){
		func(d *Draft) { d.Descriptions[0].Stale = true },
		func(d *Draft) { d.Descriptions[0].FactsRevision = 1 },
		func(d *Draft) { d.Descriptions[0].Status = "unavailable" },
		func(d *Draft) { d.Descriptions[0].Text = nil },
		func(d *Draft) { d.Descriptions[0].Text = &empty },
		func(d *Draft) { d.PrimaryLanguage = "uk" },
		func(d *Draft) { d.DescriptionPolicy = "preserve" },
		func(d *Draft) { d.Fields = []string{} },
		func(d *Draft) { d.Fields = []string{"description", "description"} },
		func(d *Draft) { d.Fields = []string{"heading"} },
		func(d *Draft) { d.Fields = []string{"gps", "description"} },
	} {
		current := base
		current.Descriptions = append([]Description{}, base.Descriptions...)
		change(&current)
		if Ready(current) {
			t.Fatalf("unready selection accepted: %+v", current)
		}
	}
	base.Fields, base.Camera = []string{"gps"}, &Point{}
	base.Descriptions[0].Stale = true
	if !Ready(base) {
		t.Fatal("stale unselected text blocked zero GPS")
	}
}
