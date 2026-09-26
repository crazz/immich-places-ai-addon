package writepreview

import (
	"strings"
	"testing"
)

func TestReplaceUsesExactTextAndValidAbsentOrNullBaseline(t *testing.T) {
	text := "  Cafe\u0301\r\nНевідоме місце.  "
	for _, before := range []TextObservation{{Presence: "absent"}, {Presence: "null"}, {Presence: "value", Value: ""}, {Presence: "value", Value: strings.Repeat("x", 64<<10)}} {
		plan, err := PlanDescription(before, DescriptionInput{Text: text, Language: "uk", Policy: "replace"})
		if err != nil || plan == nil || plan.Intended != text || plan.Before != before || plan.Lineage != nil {
			t.Fatalf("exact replacement failed: %+v %v", plan, err)
		}
	}
}

func TestManagedAppendRejectsUnownedTamperedAndOverLimitText(t *testing.T) {
	input := DescriptionInput{Text: "Approved", Language: "en", Policy: "managed_append", NewLineageID: "2dd67e75-96d5-4491-bb52-7150d65d0947"}
	first, err := PlanDescription(TextObservation{Presence: "value", Value: "Original"}, input)
	if err != nil {
		t.Fatal(err)
	}
	for name, value := range map[string]string{
		"missing": "Original", "edited": strings.Replace(first.Intended, "Approved", "Edited", 1),
		"duplicate":    first.Intended + first.Lineage.Block,
		"nested":       strings.Replace(first.Intended, "Approved", "[[Immich Places AI broken]]", 1),
		"unowned":      first.Intended,
		"malformed":    "[[Immich Places AI v1:broken",
		"oversized":    strings.Repeat("a", 64<<10),
		"invalid-utf8": string([]byte{0xff}),
	} {
		t.Run(name, func(t *testing.T) {
			request := input
			if name == "missing" || name == "edited" || name == "duplicate" || name == "nested" {
				request.Owned = first.Lineage
			}
			if plan, err := PlanDescription(TextObservation{Presence: "value", Value: value}, request); err == nil || plan != nil {
				t.Fatal("invalid append accepted", name)
			}
		})
	}
	for _, change := range []func(*DescriptionInput){
		func(i *DescriptionInput) { i.Text = "" }, func(i *DescriptionInput) { i.Text = strings.Repeat("é", 8193) },
		func(i *DescriptionInput) { i.Text = "[[/Immich Places AI v1:injected]]" },
		func(i *DescriptionInput) { i.Language = "en\ninjected" }, func(i *DescriptionInput) { i.NewLineageID = "untrusted" },
		func(i *DescriptionInput) { i.Policy = "clear" },
		func(i *DescriptionInput) {
			owned := *first.Lineage
			owned.Hash = strings.Repeat("0", 64)
			i.Owned = &owned
		},
	} {
		request := input
		change(&request)
		if _, err := PlanDescription(first.Before, request); err == nil {
			t.Fatal("invalid request accepted")
		}
	}
	for _, before := range []TextObservation{{Presence: "unavailable"}, {Presence: "null", Value: "fabricated"}} {
		if _, err := PlanDescription(before, input); err == nil {
			t.Fatal("unavailable baseline fabricated")
		}
	}
}

func TestRepeatedAppendChangesLanguageWithinSameOwnedBlock(t *testing.T) {
	first, err := PlanDescription(TextObservation{Presence: "value", Value: "Original\n"}, DescriptionInput{Text: "First description", Language: "en", Policy: "managed_append", NewLineageID: "2dd67e75-96d5-4491-bb52-7150d65d0947"})
	if err != nil {
		t.Fatal(err)
	}
	before := TextObservation{Presence: "value", Value: first.Intended + "\nAfter the block."}
	updated, err := PlanDescription(before, DescriptionInput{Text: "Оновлений опис", Language: "uk", Policy: "managed_append", NewLineageID: "fbe793c1-f6ac-4cee-a854-c4a4c397066c", Owned: first.Lineage})
	want := "Original\n\n\n[[Immich Places AI v1:2dd67e75-96d5-4491-bb52-7150d65d0947]]\nLanguage: uk\nОновлений опис\n[[/Immich Places AI v1:2dd67e75-96d5-4491-bb52-7150d65d0947]]\nAfter the block."
	if err != nil || updated == nil || updated.Intended != want || updated.Lineage.ID != first.Lineage.ID || updated.Lineage.Hash == first.Lineage.Hash {
		t.Fatalf("repeat append duplicated or altered user text: %+v %v", updated, err)
	}
}

func TestFirstManagedAppendPreservesExactUserText(t *testing.T) {
	before := TextObservation{Presence: "value", Value: "  Original café\r\nKeep this line.  "}
	plan, err := PlanDescription(before, DescriptionInput{Text: "Міст із невідомим місцем.", Language: "uk", Policy: "managed_append", NewLineageID: "2dd67e75-96d5-4491-bb52-7150d65d0947"})
	want := "  Original café\r\nKeep this line.  \n\n[[Immich Places AI v1:2dd67e75-96d5-4491-bb52-7150d65d0947]]\nLanguage: uk\nМіст із невідомим місцем.\n[[/Immich Places AI v1:2dd67e75-96d5-4491-bb52-7150d65d0947]]"
	if err != nil || plan == nil || plan.Intended != want || plan.Before != before || plan.Lineage == nil || len(plan.Lineage.Hash) != 64 {
		t.Fatalf("append did not preserve exact text: %+v %v", plan, err)
	}
}
