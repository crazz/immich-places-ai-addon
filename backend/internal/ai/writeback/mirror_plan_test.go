package writeback

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/results"
	"immich-places-backend/internal/ai/writepreview"
)

func TestOlderPlanVersionsCannotAcquireMetadataAuthority(t *testing.T) {
	preview, _ := stackPlanFixture(t)
	for _, version := range []string{"gps-preview-v1", "standard-preview-v2", "stack-preview-v3"} {
		plan := preview.Plan
		plan.Version = version
		if version != "stack-preview-v3" {
			plan.Manifest = nil
		}
		if version == "gps-preview-v1" {
			plan.ComparisonPolicy, plan.PolicyID, plan.Fields, plan.Description = "exact-nullable-gps-v1", "", []string{"gps"}, nil
		} else if version == "standard-preview-v2" {
			plan.ComparisonPolicy = "selected-standard-fields-v2"
		}
		original, _ := json.Marshal(plan)
		hash := sha256.Sum256(original)
		if _, err := DecodePlan(original, hex.EncodeToString(hash[:]), plan.Owner, plan.Installation, plan.ID); err != nil {
			t.Fatal("old plan fixture invalid", err)
		}
		plan.Mirror = &writepreview.MirrorPlan{}
		raw, _ := json.Marshal(plan)
		hash = sha256.Sum256(raw)
		if _, err := DecodePlan(raw, hex.EncodeToString(hash[:]), plan.Owner, plan.Installation, plan.ID); err == nil {
			t.Fatal("earlier plan acquired metadata authority", version)
		}
	}
}

func TestMirrorPlanDecodesExactSingleAndStackStandardAuthority(t *testing.T) {
	base, _ := stackPlanFixture(t)
	text := "Reviewed mirror translation"
	d := drafts.Draft{Revision: base.Plan.DraftRevision, FactsRevision: 1, Descriptions: []drafts.Description{{Language: "en", Status: "complete", Text: &text, FactsRevision: 1}}}
	choice := drafts.MirrorSelection{Languages: []string{"en"}}
	const record = "aaf12459-abcd-4321-8421-aaccff110033"
	export, err := writepreview.BuildMirrorExport(d, results.Document{}, choice, writepreview.MirrorProvenance{}, record)
	if err != nil {
		t.Fatal(err)
	}
	for _, stack := range []bool{true, false} {
		preview := base
		if !stack {
			preview.Plan.Manifest = nil
			preview.Plan.Version = "standard-preview-v2"
			preview.Plan.ComparisonPolicy = "selected-standard-fields-v2"
		}
		combined, raw, err := writepreview.AttachMirror(preview, writepreview.MirrorInput{RecordID: record, Export: export, Selection: choice, PolicyID: strings.Repeat("a", 64), Disclosure: writepreview.MirrorDisclosure}, writepreview.MirrorBaseline{})
		if err != nil {
			t.Fatal(err)
		}
		decoded, err := DecodePlan(raw, combined.Digest, combined.Plan.Owner, combined.Plan.Installation, combined.Plan.ID)
		if err != nil || !reflect.DeepEqual(decoded, combined.Plan) {
			t.Fatalf("v4 exact approval failed roundtrip, stack=%t: %v", stack, err)
		}
	}
}
