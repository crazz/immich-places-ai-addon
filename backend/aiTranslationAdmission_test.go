package main

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/translations"
)

func translationFixture(t *testing.T) (*aiJobFixture, *aiTranslationStore, translations.Request) {
	t.Helper()
	f, analysis := draftFixture(t)
	draft := acceptedDraft(t, f, analysis)
	s := &aiTranslationStore{drafts: &aiDraftStore{results: &aiResultStore{jobs: f.store}}, fingerprint: strings.Repeat("a", 64)}
	req := translations.Request{Key: "translation-key", DraftID: draft.ID, Revision: draft.Revision,
		FactsRevision: draft.FactsRevision, ProfileID: "profile", ProfileRevision: 1,
		Basis: "A bridge with uncertain location.", BasisKind: "scene", Languages: []string{"en", "uk"}, Confirmed: true}
	return f, s, req
}

func TestAITranslationAdmissionIsDurableAndPreservesDraft(t *testing.T) {
	f, s, req := translationFixture(t)
	ctx := context.Background()
	before, err := s.drafts.get(ctx, testUserID, req.DraftID)
	if err != nil {
		t.Fatal(err)
	}
	run, err := s.submit(ctx, testUserID, req)
	if err != nil || run.ID == "" || len(run.Items) != 2 || run.Items[0].State != "queued" || run.Request.Basis != req.Basis {
		t.Fatalf("translation admission unavailable: %#v %v", run, err)
	}
	f.reopen(t)
	s.drafts.results.jobs = f.store
	f.store.enabled = false
	again, err := s.submit(ctx, testUserID, req)
	if err != nil || !reflect.DeepEqual(run, again) {
		t.Fatal("idempotent recovery after disable/reopen failed", err)
	}
	after, err := s.drafts.get(ctx, testUserID, req.DraftID)
	if err != nil || !reflect.DeepEqual(before, after) {
		t.Fatal("generation altered draft", err)
	}
}

func TestAITranslationAdmissionRejectsSubstitutionAndConcurrentRun(t *testing.T) {
	_, s, req := translationFixture(t)
	ctx := context.Background()
	first, err := s.submit(ctx, testUserID, req)
	if err != nil {
		t.Fatal(err)
	}
	req.Basis = "Another basis."
	if _, err = s.submit(ctx, testUserID, req); err != drafts.ErrConflict {
		t.Fatal("substitution accepted", err)
	}
	req.Key = "another-key"
	if _, err = s.submit(ctx, testUserID, req); err != drafts.ErrConflict {
		t.Fatal("concurrent draft run accepted", err)
	}
	if first.ID == "" {
		t.Fatal("missing original run")
	}
}

func TestAITranslationAdmissionRequiresCurrentPrivateAuthority(t *testing.T) {
	for _, reason := range []string{"disabled", "profile", "revision", "draft revision", "facts", "foreign", "installation", "policy"} {
		t.Run(reason, func(t *testing.T) {
			f, s, req := translationFixture(t)
			owner := testUserID
			switch reason {
			case "disabled":
				selectionSQL(t, f.db, "UPDATE ai_provider_profiles SET enabled=0")
			case "profile":
				req.ProfileID = "foreign-profile"
			case "revision":
				req.ProfileRevision++
			case "draft revision":
				req.Revision++
			case "facts":
				req.FactsRevision++
			case "foreign":
				owner = "other"
			case "installation":
				s.drafts.results.jobs.binding = "obsolete"
			case "policy":
				s.fingerprint = "invalid"
			}
			if _, err := s.submit(context.Background(), owner, req); err == nil {
				t.Fatal("invalid authority admitted")
			}
			var count int
			if err := f.db.db.QueryRow("SELECT count(*) FROM ai_translation_runs").Scan(&count); err != nil || count != 0 {
				t.Fatal("retained unauthorized run", count, err)
			}
		})
	}
}
