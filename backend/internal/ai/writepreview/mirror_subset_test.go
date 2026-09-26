package writepreview

import (
	"encoding/json"
	"strings"
	"testing"

	"immich-places-backend/internal/ai/drafts"
)

func TestMirrorSelectedCurrentSubsetExcludesStaleAndUnselectedContent(t *testing.T) {
	d, doc, p := mirrorExportFixture()
	stale, unselected := "STALE PRIVATE TRANSLATION", "UNSELECTED TRANSLATION"
	d.Descriptions = append(d.Descriptions,
		drafts.Description{Language: "fr", Status: "complete", Text: &stale, FactsRevision: 1, Stale: true},
		drafts.Description{Language: "de", Status: "complete", Text: &unselected, FactsRevision: d.FactsRevision},
	)
	d.HeadingStale, d.RadiusStale = true, true
	before, _ := json.Marshal(d)
	raw, err := BuildMirrorExport(d, doc, drafts.MirrorSelection{Languages: []string{"en"}}, p, "aaf12459-abcd-4321-8421-aaccff110033")
	if err != nil || !strings.Contains(string(raw), "Reviewed description") {
		t.Fatal("stale unselected fields blocked current selected subset", err)
	}
	for _, excluded := range []string{stale, unselected, `"direction":`, `"precision":`, `"place":`, `"provenance":`, "private.invalid"} {
		if strings.Contains(string(raw), excluded) {
			t.Fatal("unselected data leaked", excluded)
		}
	}
	after, _ := json.Marshal(d)
	if string(before) != string(after) {
		t.Fatal("export mutated local reviewed content")
	}
}

func TestMirrorExportAllowsEightLanguagesAndRejectsNineBeforeSizeLimit(t *testing.T) {
	d, doc, p := mirrorExportFixture()
	tags := []string{"en", "de", "fr", "es", "it", "pt", "uk", "pl", "ja"}
	d.Descriptions = nil
	for _, tag := range tags {
		text := "Reviewed " + tag
		d.Descriptions = append(d.Descriptions, drafts.Description{Language: tag, Text: &text, Status: "complete", FactsRevision: d.FactsRevision})
	}
	raw, err := BuildMirrorExport(d, doc, drafts.MirrorSelection{Languages: tags[:8]}, p, "aaf12459-abcd-4321-8421-aaccff110033")
	var exported MirrorExport
	if err != nil || json.Unmarshal(raw, &exported) != nil || len(exported.Descriptions) != 8 {
		t.Fatal("eight current languages not exported", err)
	}
	if raw, err := BuildMirrorExport(d, doc, drafts.MirrorSelection{Languages: tags}, p, "aaf12459-abcd-4321-8421-aaccff110033"); err == nil || raw != nil {
		t.Fatal("ninth language admitted below byte limit")
	}
}
