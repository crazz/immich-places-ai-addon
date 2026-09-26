package writeback

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"immich-places-backend/internal/ai/writepreview"
)

func stackPlanFixture(t *testing.T) (writepreview.Preview, []byte) {
	t.Helper()
	now := time.Unix(1000, 0)
	primary, sibling := "00000000-0000-4000-8000-000000000001", "00000000-0000-4000-8000-000000000002"
	text := writepreview.TextObservation{Presence: "null"}
	snapshot := writepreview.Snapshot{Owner: "owner", Installation: "installation", ID: "draft", Revision: 2, AssetID: primary, AnalysisID: "analysis", State: "staged", Fields: []string{"gps", "description"}, Camera: &writepreview.Point{Latitude: 0, Longitude: 12}, BaselineReviewed: true, ImageIdentity: "image1", DescriptionBaseline: &text, Description: &writepreview.DescriptionInput{Text: "Exact é\r\ntext ", Language: "en", Policy: "replace"}, PolicyID: strings.Repeat("a", 64)}
	review := writepreview.StackReview{ID: "00000000-0000-4000-8000-000000000004", Owner: snapshot.Owner, Installation: snapshot.Installation, DraftID: snapshot.ID, DraftRevision: 2, AnalyzedID: primary, ImageIdentity: "image1", StackID: "00000000-0000-4000-8000-000000000003", ExpiresAt: now.Add(time.Minute).Format(time.RFC3339Nano), Targets: []writepreview.ReviewedTarget{{AssetID: primary, ImageIdentity: "image1"}, {AssetID: sibling, ImageIdentity: "image2"}}}
	current := map[string]writepreview.Metadata{primary: {ImageIdentity: "image1", Description: &text}, sibling: {ImageIdentity: "image2"}}
	preview, raw, err := writepreview.BuildStack(snapshot, review, current, "preview", now)
	if err != nil {
		t.Fatal(err)
	}
	return preview, raw
}

func TestDecodeStackPlanRetainsExactTargetAuthority(t *testing.T) {
	preview, raw := stackPlanFixture(t)
	plan, err := DecodePlan(raw, preview.Digest, "owner", "installation", "preview")
	if err != nil || !reflect.DeepEqual(plan, preview.Plan) {
		t.Fatalf("v3 authority failed roundtrip: %v", err)
	}
}

func TestLegacyPlanCannotAcquireStackManifest(t *testing.T) {
	preview, _ := stackPlanFixture(t)
	for _, version := range []string{"gps-preview-v1", "standard-preview-v2"} {
		t.Run(version, func(t *testing.T) {
			plan := preview.Plan
			plan.Version = version
			if version == "gps-preview-v1" {
				plan.ComparisonPolicy = "exact-nullable-gps-v1"
				plan.Fields = []string{"gps"}
				plan.Description = nil
				plan.PolicyID = ""
			} else {
				plan.ComparisonPolicy = "selected-standard-fields-v2"
			}
			raw, err := json.Marshal(plan)
			if err != nil {
				t.Fatal(err)
			}
			sum := sha256.Sum256(raw)
			if _, err = DecodePlan(raw, hex.EncodeToString(sum[:]), "owner", "installation", "preview"); err == nil {
				t.Fatal("older version acquired stack authority")
			}
		})
	}
}

func TestStackDecoderRejectsAlteredFieldMatrixAndScope(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*writepreview.Plan)
	}{
		{"unknown version", func(p *writepreview.Plan) { p.Version = "stack-preview-v4" }},
		{"unknown policy", func(p *writepreview.Plan) { p.ComparisonPolicy = "other" }},
		{"invalid capability identity", func(p *writepreview.Plan) { p.PolicyID = "invalid" }},
		{"no manifest", func(p *writepreview.Plan) { p.Manifest = nil }},
		{"duplicate target", func(p *writepreview.Plan) { p.Manifest.Targets[1] = p.Manifest.Targets[0] }},
		{"unordered targets", func(p *writepreview.Plan) {
			p.Manifest.Targets[0], p.Manifest.Targets[1] = p.Manifest.Targets[1], p.Manifest.Targets[0]
		}},
		{"no primary", func(p *writepreview.Plan) { p.TargetID = "00000000-0000-4000-8000-000000000099" }},
		{"stack substitution", func(p *writepreview.Plan) { p.Manifest.StackID = "invalid" }},
		{"empty evidence", func(p *writepreview.Plan) { p.Manifest.ReviewID = "" }},
		{"primary mismatch", func(p *writepreview.Plan) { p.Manifest.Targets[0].ImageIdentity = "different" }},
		{"sibling description", func(p *writepreview.Plan) {
			p.Manifest.Targets[1].Description = p.Description
			p.Manifest.Targets[1].Fields = []string{"gps", "description"}
		}},
		{"other coordinates", func(p *writepreview.Plan) { p.Manifest.Targets[1].Intended.Longitude++ }},
		{"invalid baseline", func(p *writepreview.Plan) { v := 91.0; p.Manifest.Targets[1].Before.Latitude = &v }},
		{"no source", func(p *writepreview.Plan) { p.Manifest.Targets[1].ImageIdentity = "" }},
		{"unknown field", func(p *writepreview.Plan) { p.Manifest.Targets[1].Fields = []string{"heading"} }},
		{"over limit", func(p *writepreview.Plan) {
			for len(p.Manifest.Targets) < 51 {
				p.Manifest.Targets = append(p.Manifest.Targets, p.Manifest.Targets[1])
			}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			preview, _ := stackPlanFixture(t)
			tc.mutate(&preview.Plan)
			raw, err := json.Marshal(preview.Plan)
			if err != nil {
				t.Fatal(err)
			}
			sum := sha256.Sum256(raw)
			if _, err = DecodePlan(raw, hex.EncodeToString(sum[:]), "owner", "installation", "preview"); err == nil {
				t.Fatal("altered scope accepted")
			}
		})
	}
}
