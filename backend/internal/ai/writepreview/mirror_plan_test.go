package writepreview

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"immich-places-backend/internal/ai/drafts"
)

func TestMirrorPlanBindsExactExportDisclosureAndAnalyzedStandardStep(t *testing.T) {
	d, doc, provenance := mirrorExportFixture()
	choice := drafts.MirrorSelection{Languages: []string{"en"}}
	const record = "aaf12459-abcd-4321-8421-aaccff110033"
	export, err := BuildMirrorExport(d, doc, choice, provenance, record)
	if err != nil {
		t.Fatal(err)
	}
	snapshot := Snapshot{Owner: "owner", Installation: "installation", ID: "draft", AnalysisID: "analysis", AssetID: "asset", Revision: d.Revision, State: "staged", Fields: []string{"gps"}, Camera: &Point{}, BaselineReviewed: true, ImageIdentity: "v1:image"}
	base, baseRaw, err := Build(snapshot, Metadata{ImageIdentity: snapshot.ImageIdentity}, "preview", time.Date(2026, 9, 26, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	input := MirrorInput{RecordID: record, Export: export, Selection: choice, PolicyID: strings.Repeat("a", 64), Disclosure: MirrorDisclosure}
	combined, raw, err := AttachMirror(base, input, MirrorBaseline{})
	if err != nil || combined.Plan.Version != "mirror-preview-v4" || combined.Plan.Mirror == nil || combined.Plan.Manifest == nil || len(combined.Plan.Manifest.Targets) != 1 {
		t.Fatalf("combined plan unavailable: %+v %v", combined, err)
	}
	m := combined.Plan.Mirror
	if m.Key != MirrorNamespace || m.RecordID != record || m.Disclosure != MirrorDisclosure || m.Before.Present || string(m.Value) != string(export) || combined.Plan.Manifest.Targets[0].AssetID != "asset" || combined.Plan.Manifest.Targets[0].Fields[0] != "gps" || combined.Plan.PolicyID != input.PolicyID {
		t.Fatal("combined plan expanded or omitted authority")
	}
	hash := sha256.Sum256(raw)
	if combined.Digest != hex.EncodeToString(hash[:]) || combined.Digest == base.Digest || combined.Diff != "changed" {
		t.Fatal("digest/diff omitted metadata authority")
	}
	after, _ := json.Marshal(base.Plan)
	if string(after) != string(baseRaw) || base.Plan.Mirror != nil || base.Plan.Manifest != nil {
		t.Fatal("old plan bytes mutated")
	}
}

func TestMirrorPlanRejectsMissingDisclosureExpandedScopeAndUnownedBaseline(t *testing.T) {
	d, doc, p := mirrorExportFixture()
	choice := drafts.MirrorSelection{Languages: []string{"en"}}
	const record = "aaf12459-abcd-4321-8421-aaccff110033"
	export, err := BuildMirrorExport(d, doc, choice, p, record)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name   string
		change func(*Preview, *MirrorInput, *MirrorBaseline)
	}{
		{"missing disclosure", func(_ *Preview, i *MirrorInput, _ *MirrorBaseline) { i.Disclosure = "" }},
		{"wrong disclosure", func(_ *Preview, i *MirrorInput, _ *MirrorBaseline) { i.Disclosure = "other" }},
		{"no standard field", func(p *Preview, _ *MirrorInput, _ *MirrorBaseline) { p.Plan.Fields = nil }},
		{"unsupported field", func(p *Preview, _ *MirrorInput, _ *MirrorBaseline) { p.Plan.Fields = []string{"metadata"} }},
		{"missing capability", func(_ *Preview, i *MirrorInput, _ *MirrorBaseline) { i.PolicyID = "" }},
		{"different record", func(_ *Preview, i *MirrorInput, _ *MirrorBaseline) {
			i.RecordID = "baf12459-abcd-4321-8421-aaccff110033"
		}},
		{"missing export", func(_ *Preview, i *MirrorInput, _ *MirrorBaseline) { i.Export = nil }},
		{"wrong revision", func(p *Preview, _ *MirrorInput, _ *MirrorBaseline) { p.Plan.DraftRevision++ }},
		{"different selection", func(_ *Preview, i *MirrorInput, _ *MirrorBaseline) { i.Selection.Languages = []string{"de"} }},
		{"unapproved content", func(_ *Preview, i *MirrorInput, _ *MirrorBaseline) {
			i.Export = json.RawMessage(strings.TrimSuffix(string(export), "}") + `,"prompt":"private"}`)
		}},
		{"foreign namespace", func(_ *Preview, _ *MirrorInput, b *MirrorBaseline) {
			b.Present = true
			b.Value = []byte(`{"foreign":true}`)
		}},
		{"duplicate attachment", func(p *Preview, _ *MirrorInput, _ *MirrorBaseline) { p.Plan.Mirror = &MirrorPlan{} }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			base := Preview{Plan: Plan{Version: "gps-preview-v1", TargetID: "asset", Fields: []string{"gps"}, DraftRevision: d.Revision}}
			input := MirrorInput{RecordID: record, Export: export, Selection: choice, PolicyID: strings.Repeat("a", 64), Disclosure: MirrorDisclosure}
			before := MirrorBaseline{}
			tc.change(&base, &input, &before)
			if _, raw, err := AttachMirror(base, input, before); err == nil || raw != nil {
				t.Fatal("unapproved mirror scope admitted")
			}
		})
	}
}
