package results

import (
	"strings"
	"testing"
)

func firstDescription(m map[string]any) map[string]any {
	return m["descriptions"].([]any)[0].(map[string]any)
}

func TestDescriptionsCoverOnlyTheRequestedLanguagesAndStatuses(t *testing.T) {
	v := testValidator(t)
	for _, change := range []func(map[string]any){
		func(m map[string]any) { m["descriptions"] = []any{} },
		func(m map[string]any) { firstDescription(m)["language"] = "uk" },
		func(m map[string]any) { firstDescription(m)["language"] = "en_US" },
		func(m map[string]any) { m["descriptions"] = append(m["descriptions"].([]any), firstDescription(m)) },
		func(m map[string]any) { firstDescription(m)["text"] = " " },
		func(m map[string]any) { firstDescription(m)["unavailable_reason"] = "unexpected" },
		func(m map[string]any) { firstDescription(m)["status"] = "unavailable" },
		func(m map[string]any) {
			firstDescription(m)["status"] = "unavailable"
			firstDescription(m)["text"] = nil
			firstDescription(m)["unavailable_reason"] = " "
		},
		func(m map[string]any) { firstDescription(m)["basis"] = "candidate" },
	} {
		requireFailure(t, v, mutateFixture(t, "unknown", change), visualContext("en"), "semantic_violation")
	}
	data := mutateFixture(t, "synthetic", func(m map[string]any) { firstDescription(m)["basis"] = "scene_only" })
	requireFailure(t, v, data, visualContext("en", "uk"), "semantic_violation")
}

func TestNonEnglishStatusesAndNormalizedTagsArePreserved(t *testing.T) {
	ctx := visualContext("uk", "pt-br")
	ctx.PrimaryLanguage = "UK"
	data := mutateFixture(t, "unknown", func(m map[string]any) {
		firstDescription(m)["language"] = "UK"
		firstDescription(m)["text"] = "  Сцена без визначеної локації.  "
		m["descriptions"] = append(m["descriptions"].([]any), map[string]any{"language": "pt-br", "status": "unavailable", "text": nil, "basis": "scene_only", "candidate_id": nil, "unavailable_reason": "Synthetic translation unavailable"})
	})
	v := testValidator(t)
	proposal, err := v.Validate(data, ctx)
	if err != nil {
		t.Fatal(err)
	}
	doc, err := proposal.Data()
	if err != nil {
		t.Fatal(err)
	}
	if doc.Descriptions[0].Language != "uk" || doc.Descriptions[1].Language != "pt-BR" || doc.Descriptions[1].Status != "unavailable" || doc.Descriptions[1].Text != nil || *doc.Descriptions[0].Text != "  Сцена без визначеної локації.  " {
		t.Fatalf("status/tag/prose changed: %+v", doc.Descriptions)
	}
}

func TestLanguageAliasesCannotDuplicateOutputCoverage(t *testing.T) {
	ctx := visualContext("he")
	good := mutateFixture(t, "unknown", func(m map[string]any) { firstDescription(m)["language"] = "iw" })
	v := testValidator(t)
	proposal, err := v.Validate(good, ctx)
	if err != nil {
		t.Fatal(err)
	}
	doc, err := proposal.Data()
	if err != nil || doc.Descriptions[0].Language != "he" {
		t.Fatal("alias not normalized", err)
	}
	duplicated := mutateFixture(t, "unknown", func(m map[string]any) {
		firstDescription(m)["language"] = "iw"
		other := map[string]any{}
		for key, value := range firstDescription(m) {
			other[key] = value
		}
		other["language"] = "HE"
		m["descriptions"] = append(m["descriptions"].([]any), other)
	})
	requireFailure(t, v, duplicated, ctx, "semantic_violation")
}

func TestNormalizedLanguageTagsRespectTheByteCeiling(t *testing.T) {
	exact := "en-x-" + strings.Repeat("abcdefgh-", 13) + "abcdef"
	if tag, ok := normalizeLanguage(exact); !ok || len(tag) != 128 {
		t.Fatalf("valid exact tag rejected: length=%d accepted=%v", len(tag), ok)
	}
	expanding := "sh-x-" + strings.Repeat("abcdefgh-", 13) + "abcdef"
	if tag, ok := normalizeLanguage(expanding); ok && len(tag) > 128 {
		t.Fatalf("normalization exceeded tag bound: %d", len(tag))
	}
}
