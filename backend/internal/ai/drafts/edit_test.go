package drafts

import "testing"

func TestReviewOneLanguageRetainsOtherStaleAndUnavailableContent(t *testing.T) {
	heading := 40.0
	text := "Inherited description"
	current := Draft{State: "draft", Revision: 2, FactsRevision: 2, Camera: &Point{}, Fields: []string{"gps"}, Heading: &heading, HeadingStale: true, Descriptions: []Description{
		{Language: "en", Status: "complete", Text: &text, Basis: "candidate", Stale: true, FactsRevision: 1},
		{Language: "uk", Status: "complete", Text: &text, Basis: "candidate", Stale: true, FactsRevision: 1},
		{Language: "fr", Status: "unavailable", Basis: "candidate", Stale: true, FactsRevision: 1},
	}}
	next, err := Apply(current, Edit{Descriptions: map[string]DescriptionEdit{"en": {Review: true}}, State: "staged"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if next.State != "staged" || next.Descriptions[0].Stale || next.Descriptions[0].FactsRevision != 2 || !next.Descriptions[1].Stale || next.Descriptions[1].FactsRevision != 1 || next.Descriptions[2].Status != "unavailable" || next.Descriptions[2].Text != nil || !next.HeadingStale {
		t.Fatal("review scope expanded", next)
	}
	if !current.Descriptions[0].Stale {
		t.Fatal("previous revision mutated")
	}
}
