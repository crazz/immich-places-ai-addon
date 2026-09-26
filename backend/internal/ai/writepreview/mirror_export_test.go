package writepreview

import (
	"encoding/json"
	"strings"
	"testing"

	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/results"
)

func mirrorExportFixture() (drafts.Draft, results.Document, MirrorProvenance) {
	heading, radius := 90.0, 120.0
	text, candidate, uncertainty := "Reviewed description", "candidate", json.Number("15")
	d := drafts.Draft{Revision: 3, FactsRevision: 2, Heading: &heading, HeadingFactsRevision: 2, Radius: &radius, RadiusBasis: "visual_estimate", CandidateID: &candidate, UpdatedAt: "2026-09-26T12:00:00Z", Descriptions: []drafts.Description{{Language: "en", Status: "complete", Text: &text, FactsRevision: 2}}}
	headingUncertainty := 15.0
	d.HeadingMethod, d.HeadingUncertainty = "visual_estimate", &headingUncertainty
	d.Baseline.OwnerID = "PRIVATE_ACCOUNT"
	d.OriginalSourceDigest = "PRIVATE_SOURCE"
	doc := results.Document{Candidates: []results.Candidate{{ID: candidate, PlaceName: "Reviewed place", CameraDirection: &results.Direction{AzimuthDeg: "90", Method: "visual_estimate", UncertaintyDeg: &uncertainty}, SupportSummary: "PRIVATE_HINT"}}, Warnings: []string{"PRIVATE_PROMPT"}, Sources: []results.AnswerSource{{URL: "https://private.invalid/source"}}}
	return d, doc, MirrorProvenance{Mode: "visual", Model: "test-model"}
}

func TestMirrorExportRejectsStaleInvalidOrOversizedSelection(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(*drafts.Draft, *drafts.MirrorSelection, *MirrorProvenance)
	}{
		{"stale direction", func(d *drafts.Draft, _ *drafts.MirrorSelection, _ *MirrorProvenance) { d.HeadingStale = true }},
		{"old direction", func(d *drafts.Draft, _ *drafts.MirrorSelection, _ *MirrorProvenance) { d.HeadingFactsRevision-- }},
		{"stale precision", func(d *drafts.Draft, _ *drafts.MirrorSelection, _ *MirrorProvenance) { d.RadiusStale = true }},
		{"missing place", func(d *drafts.Draft, _ *drafts.MirrorSelection, _ *MirrorProvenance) { d.CandidateID = nil }},
		{"stale text", func(d *drafts.Draft, _ *drafts.MirrorSelection, _ *MirrorProvenance) { d.Descriptions[0].Stale = true }},
		{"old text", func(d *drafts.Draft, _ *drafts.MirrorSelection, _ *MirrorProvenance) {
			d.Descriptions[0].FactsRevision--
		}},
		{"unavailable text", func(d *drafts.Draft, _ *drafts.MirrorSelection, _ *MirrorProvenance) {
			d.Descriptions[0].Status = "unavailable"
		}},
		{"missing language", func(_ *drafts.Draft, c *drafts.MirrorSelection, _ *MirrorProvenance) { c.Languages = []string{"de"} }},
		{"duplicate language", func(_ *drafts.Draft, c *drafts.MirrorSelection, _ *MirrorProvenance) {
			c.Languages = []string{"en", "en"}
		}},
		{"empty choice", func(_ *drafts.Draft, c *drafts.MirrorSelection, _ *MirrorProvenance) { *c = drafts.MirrorSelection{} }},
		{"model without provenance", func(_ *drafts.Draft, c *drafts.MirrorSelection, _ *MirrorProvenance) {
			c.Model = true
			c.Provenance = false
		}},
		{"unknown mode", func(_ *drafts.Draft, _ *drafts.MirrorSelection, p *MirrorProvenance) { p.Mode = "private context" }},
		{"missing review", func(d *drafts.Draft, _ *drafts.MirrorSelection, _ *MirrorProvenance) { d.UpdatedAt = "" }},
		{"invalid revision", func(d *drafts.Draft, _ *drafts.MirrorSelection, _ *MirrorProvenance) { d.Revision = 0 }},
		{"oversized", func(d *drafts.Draft, c *drafts.MirrorSelection, _ *MirrorProvenance) {
			c.Languages = []string{"en", "de", "fr", "es", "it"}
			d.Descriptions = nil
			for _, tag := range c.Languages {
				text := strings.Repeat("é", 8000)
				d.Descriptions = append(d.Descriptions, drafts.Description{Language: tag, Status: "complete", Text: &text, FactsRevision: d.FactsRevision})
			}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d, doc, p := mirrorExportFixture()
			choice := drafts.MirrorSelection{Direction: true, Precision: true, Place: true, Languages: []string{"en"}, Provenance: true}
			tc.change(&d, &choice, &p)
			if raw, err := BuildMirrorExport(d, doc, choice, p, "aaf12459-abcd-4321-8421-aaccff110033"); err == nil || raw != nil {
				t.Fatal("invalid selection silently exported or truncated")
			}
		})
	}
}

func TestMirrorExportContainsOnlyExplicitReviewedContent(t *testing.T) {
	d, doc, p := mirrorExportFixture()
	choice := drafts.MirrorSelection{Direction: true, Precision: true, Place: true, Languages: []string{"en"}, Provenance: true}
	raw, err := BuildMirrorExport(d, doc, choice, p, "aaf12459-abcd-4321-8421-aaccff110033")
	if err != nil {
		t.Fatal(err)
	}
	var value map[string]json.RawMessage
	if json.Unmarshal(raw, &value) != nil || string(value["schemaVersion"]) != "1" || string(value["application"]) != `"immich-places-ai-addon"` || len(value) != 9 {
		t.Fatalf("unexpected export envelope: %s", raw)
	}
	for _, wanted := range []string{"Reviewed description", "Reviewed place", `"heading":90`, `"uncertainty":15`, `"radius":120`, `"draftRevision":3`, `"factsRevision":2`, `"mode":"visual"`} {
		if !strings.Contains(string(raw), wanted) {
			t.Fatalf("selected content omitted: %s", wanted)
		}
	}
	for _, forbidden := range []string{"PRIVATE", "private.invalid", "test-model", "owner", "sourceDigest", "supportSummary"} {
		if strings.Contains(string(raw), forbidden) {
			t.Fatalf("private/unselected content leaked: %s", forbidden)
		}
	}
	minimal, err := BuildMirrorExport(d, doc, drafts.MirrorSelection{Languages: []string{"en"}}, p, "aaf12459-abcd-4321-8421-aaccff110033")
	if err != nil || strings.Contains(string(minimal), "Reviewed place") || strings.Contains(string(minimal), "provenance") || strings.Contains(string(minimal), "direction") || strings.Contains(string(minimal), "precision") {
		t.Fatalf("unselected fields exported: %s %v", minimal, err)
	}
}

func TestMirrorUnknownDirectionNeverBorrowsCandidateOrSubjectBearing(t *testing.T) {
	d, doc, p := mirrorExportFixture()
	d.Heading = nil
	doc.Candidates[0].Subject = &results.Subject{Name: "Visible subject", Location: &results.Coordinate{Latitude: "80", Longitude: "10"}}
	raw, err := BuildMirrorExport(d, doc, drafts.MirrorSelection{Direction: true}, p, "aaf12459-abcd-4321-8421-aaccff110033")
	if err != nil || !strings.Contains(string(raw), `"direction":{"heading":null,"method":"unknown","uncertainty":null}`) {
		t.Fatalf("unknown direction borrowed unrelated geometry: %s %v", raw, err)
	}
}
