package main

import (
	"context"
	"testing"
)

func TestAITranslationRetryCannotRegenerateSuccessfulParentLanguage(t *testing.T) {
	f, s, req := translationFixture(t)
	ctx := context.Background()
	parent, err := s.submit(ctx, testUserID, req)
	if err != nil {
		t.Fatal(err)
	}
	selectionSQL(t, f.db, "UPDATE ai_translation_items SET state=CASE WHEN language='en' THEN 'complete' ELSE 'failed' END,text=CASE WHEN language='en' THEN 'Retained success.' ELSE NULL END WHERE runID=?", parent.ID)
	req.Key = "retry"
	req.ParentID = parent.ID
	if _, err = s.submit(ctx, testUserID, req); err == nil {
		t.Fatal("successful language included implicitly in retry")
	}
	req.Languages = []string{"uk"}
	run, err := s.submit(ctx, testUserID, req)
	if err != nil || run.ID == parent.ID || len(run.Items) != 1 || run.Items[0].Language != "uk" {
		t.Fatal("explicit failed-language retry unavailable", run, err)
	}
	var text string
	if err = f.db.db.QueryRow("SELECT text FROM ai_translation_items WHERE runID=? AND language='en'", parent.ID).Scan(&text); err != nil || text != "Retained success." {
		t.Fatal("parent success changed", text, err)
	}
}
