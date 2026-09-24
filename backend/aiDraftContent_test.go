package main

import (
	"encoding/json"
	"testing"

	"immich-places-backend/internal/ai/results"
)

func TestAIDraftCameraMoveInvalidatesOnlyDependentContent(t *testing.T) {
	f, id := draftProposalFixture(t, func(d *results.Document) { d.Descriptions[1].Basis = "scene_only"; d.Descriptions[1].CandidateID = nil })
	value := acceptedDraft(t, f, id)
	rec := draftPatch(f, value.ID, `"1"`, `{"camera":{"latitude":0,"longitude":0}}`)
	if rec.Code != 200 {
		t.Fatal(rec.Code, rec.Body.String())
	}
	var saved struct {
		Heading      *float64 `json:"heading"`
		HeadingStale bool     `json:"headingStale"`
		Radius       *float64 `json:"radius"`
		RadiusStale  bool     `json:"radiusStale"`
		Descriptions []struct {
			Language      string `json:"language"`
			Stale         bool   `json:"stale"`
			FactsRevision int    `json:"factsRevision"`
		} `json:"descriptions"`
	}
	if json.Unmarshal(rec.Body.Bytes(), &saved) != nil || saved.Heading == nil || *saved.Heading != 40 || !saved.HeadingStale || saved.Radius == nil || *saved.Radius != 250 || !saved.RadiusStale || len(saved.Descriptions) != 2 {
		t.Fatal("dependent values not retained as stale", saved)
	}
	if !saved.Descriptions[0].Stale || saved.Descriptions[1].Stale || saved.Descriptions[1].FactsRevision != 1 {
		t.Fatal("scene basis changed", saved.Descriptions)
	}
}

func TestAIDraftReviewsOneLanguageWithoutApprovingOtherFields(t *testing.T) {
	f, id := draftProposalFixture(t, func(d *results.Document) {})
	value := acceptedDraft(t, f, id)
	if rec := draftPatch(f, value.ID, `"1"`, `{"camera":{"latitude":1,"longitude":2}}`); rec.Code != 200 {
		t.Fatal(rec.Code)
	}
	rec := draftPatch(f, value.ID, `"2"`, `{"descriptions":{"en":{"text":"User correction"}},"state":"staged","fields":["gps"]}`)
	if rec.Code != 200 {
		t.Fatalf("review language: %d %s", rec.Code, rec.Body.String())
	}
	if json.Unmarshal(rec.Body.Bytes(), &value) != nil || value.Descriptions[0].Stale || value.Descriptions[0].FactsRevision != 2 || !value.Descriptions[0].UserSupplied || *value.Descriptions[0].Text != "User correction" || !value.Descriptions[1].Stale || !value.HeadingStale || value.State != "staged" {
		t.Fatal("review expanded", value)
	}
}

func TestAIDraftGeometryAndScopeValidationIsAtomic(t *testing.T) {
	f, id := draftFixture(t)
	value := acceptedDraft(t, f, id)
	rec := draftPatch(f, value.ID, `"1"`, `{"heading":0}`)
	if rec.Code != 200 {
		t.Fatalf("valid heading: %d %s", rec.Code, rec.Body.String())
	}
	for _, body := range []string{
		`{"camera":{"latitude":91,"longitude":0}}`, `{"camera":{"latitude":0}}`,
		`{"camera":{"latitude":0,"longitude":181}}`, `{"heading":360}`, `{"heading":-1}`,
		`{"fields":["description"]}`, `{"fields":["gps","gps"]}`, `{"assetId":"other"}`,
		`{"descriptions":{"fr":{"text":"unrequested"}}}`, `{"camera":{"latitude":null,"longitude":0}}`,
	} {
		rec = draftPatch(f, value.ID, `"2"`, body)
		if rec.Code != 400 {
			t.Fatal("invalid input accepted", body, rec.Code)
		}
	}
	rec = draftPatch(f, value.ID, `"2"`, `{"heading":null}`)
	if rec.Code != 200 || json.Unmarshal(rec.Body.Bytes(), &value) != nil || value.Revision != 3 || value.Heading != nil {
		t.Fatal("validation changed revision or lost null", rec.Code, value)
	}
}
