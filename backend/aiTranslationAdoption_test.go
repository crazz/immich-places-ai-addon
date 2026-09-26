package main

import (
	"context"
	"immich-places-backend/internal/ai/drafts"
	"reflect"
	"testing"
)

func TestAITranslationAdoptsOnlySelectedSuccessfulLanguage(t *testing.T) {
	f, s, req := translationFixture(t)
	ctx := context.Background()
	before, err := s.drafts.get(ctx, testUserID, req.DraftID)
	if err != nil {
		t.Fatal(err)
	}
	run, err := s.submit(ctx, testUserID, req)
	if err != nil {
		t.Fatal(err)
	}
	selectionSQL(t, f.db, "UPDATE ai_translation_items SET state='complete',text='Selected translation.' WHERE runID=?", run.ID)
	got, err := s.adopt(ctx, testUserID, run.ID, before.Revision, []string{"uk"})
	if err != nil || got.Revision != before.Revision+1 || got.FactsRevision != before.FactsRevision || !reflect.DeepEqual(got.Camera, before.Camera) || len(got.Descriptions) != len(before.Descriptions)+1 {
		t.Fatal("selected adoption unavailable or altered geometry", got, err)
	}
	if !reflect.DeepEqual(got.Descriptions[:len(before.Descriptions)], before.Descriptions) {
		t.Fatal("unselected text overwritten")
	}
	added := got.Descriptions[len(got.Descriptions)-1]
	if added.Language != "uk" || added.Text == nil || *added.Text != "Selected translation." || added.Stale || added.FactsRevision != got.FactsRevision || added.Basis != "scene_only" {
		t.Fatal("incorrect adopted suggestion", added)
	}
}

func TestAITranslationRejectsStaleOrUnusableAdoption(t *testing.T) {
	for _, reason := range []string{"revision", "edited", "camera", "empty", "duplicate", "unknown", "failed", "canceled", "no text"} {
		t.Run(reason, func(t *testing.T) {
			f, s, req := translationFixture(t)
			ctx := context.Background()
			run, err := s.submit(ctx, testUserID, req)
			if err != nil {
				t.Fatal(err)
			}
			selectionSQL(t, f.db, "UPDATE ai_translation_items SET state='complete',text='Translated.' WHERE runID=?", run.ID)
			langs := []string{"en"}
			revision := req.Revision
			switch reason {
			case "revision":
				revision++
			case "edited":
				text := "Manual text"
				_, err = s.drafts.edit(ctx, testUserID, req.DraftID, revision, drafts.Edit{Descriptions: map[string]drafts.DescriptionEdit{"en": {Text: &text}}})
			case "camera":
				_, err = s.drafts.edit(ctx, testUserID, req.DraftID, revision, drafts.Edit{Camera: []byte(`{"latitude":0,"longitude":12}`)})
			case "empty":
				langs = nil
			case "duplicate":
				langs = []string{"en", "en"}
			case "unknown":
				langs = []string{"fr"}
			case "failed", "canceled":
				selectionSQL(t, f.db, "UPDATE ai_translation_items SET state=?,text=NULL WHERE runID=?", reason, run.ID)
			case "no text":
				selectionSQL(t, f.db, "UPDATE ai_translation_items SET text=NULL WHERE runID=?", run.ID)
			}
			if err != nil {
				t.Fatal(err)
			}
			before, err := s.drafts.get(ctx, testUserID, req.DraftID)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = s.adopt(ctx, testUserID, run.ID, revision, langs); err == nil {
				t.Fatal("invalid suggestion adopted")
			}
			after, err := s.drafts.get(ctx, testUserID, req.DraftID)
			if err != nil || !reflect.DeepEqual(before, after) {
				t.Fatal("conflict overwrote draft", err)
			}
		})
	}
}
