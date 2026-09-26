package writepreview

import (
	"encoding/json"
	"testing"

	"immich-places-backend/internal/ai/drafts"
)

func TestMirrorPlanRejectsUnsupportedGeometryAndEvidence(t *testing.T) {
	d, doc, p := mirrorExportFixture()
	const id = "aaf12459-abcd-4321-8421-aaccff110033"
	choice := drafts.MirrorSelection{Direction: true, Precision: true}
	raw, err := BuildMirrorExport(d, doc, choice, p, id)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name   string
		change func(*MirrorExport)
	}{
		{"negative heading", func(v *MirrorExport) { n := -1.0; v.Direction.Heading = &n }},
		{"full circle", func(v *MirrorExport) { n := 360.0; v.Direction.Heading = &n }},
		{"unknown method", func(v *MirrorExport) { v.Direction.Method = "private hint" }},
		{"invalid uncertainty", func(v *MirrorExport) { n := 181.0; v.Direction.Uncertainty = &n }},
		{"negative radius", func(v *MirrorExport) { n := -1.0; v.Precision.Radius = &n }},
		{"invalid precision basis", func(v *MirrorExport) { v.Precision.Basis = "private prompt" }},
		{"unknown bearing evidence", func(v *MirrorExport) { v.Direction.Heading = nil }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var value MirrorExport
			if err := json.Unmarshal(raw, &value); err != nil {
				t.Fatal(err)
			}
			tc.change(&value)
			changed, _ := json.Marshal(value)
			plan := MirrorPlan{Key: MirrorNamespace, RecordID: id, Value: changed, Selection: choice, Disclosure: MirrorDisclosure}
			if ValidMirrorPlan(plan, d.Revision) {
				t.Fatal("unsupported exported geometry admitted")
			}
		})
	}
}
