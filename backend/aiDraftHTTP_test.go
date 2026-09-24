package main

import (
	"strings"
	"testing"
)

func TestAIDraftStrictBodiesPreconditionsAndSafeStorageFailure(t *testing.T) {
	f, id := draftFixture(t)
	value := acceptedDraft(t, f, id)
	for index, body := range []string{`{"state":"draft","state":"rejected"}`, `{"camera":{"latitude":0,"longitude":0,"longitude":1}}`, `{"state":"` + strings.Repeat("x", 33<<10) + `"}`, `{} {}`, `{"heading":1e999}`, `{"fields":null}`} {
		rec := draftPatch(f, value.ID, `"1"`, body)
		if rec.Code != 400 {
			t.Fatal("invalid body accepted", index, rec.Code)
		}
	}
	for _, revision := range []string{`1`, `"01"`, `W/"1"`, `"0"`, `"-1"`, `"1", "2"`} {
		if rec := draftPatch(f, value.ID, revision, `{"state":"rejected"}`); rec.Code != 400 {
			t.Fatal("invalid revision", rec.Code)
		}
	}
	selectionSQL(t, f.db, `CREATE TRIGGER draft_storage_failure BEFORE INSERT ON ai_draft_revisions BEGIN SELECT RAISE(ABORT,'private storage detail'); END`)
	rec := draftPatch(f, value.ID, `"1"`, `{"state":"rejected"}`)
	if rec.Code != 503 || strings.Contains(rec.Body.String(), "private storage detail") {
		t.Fatal("unsafe storage response", rec.Code, rec.Body.String())
	}
	latest := acceptedDraft(t, f, id)
	if latest.Revision != 1 || latest.State != "draft" {
		t.Fatal("failed request changed state", latest)
	}
}
