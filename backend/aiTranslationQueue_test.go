package main

import (
	"context"
	"immich-places-backend/internal/ai/drafts"
	"testing"
)

func TestAITranslationAdmissionBoundsPendingWorkIncludingAnalysis(t *testing.T) {
	f, s, req := translationFixture(t)
	selectionSQL(t, f.db, `WITH RECURSIVE n(x) AS (SELECT 1 UNION ALL SELECT x+1 FROM n WHERE x<100) INSERT INTO ai_job_items(userID,jobID,id,assetID,position) SELECT userID,id,'pending-'||x,'asset-'||x,1000+x FROM ai_jobs,n LIMIT 100`)
	if _, err := s.submit(context.Background(), testUserID, req); err != drafts.ErrConflict {
		t.Fatal("queue cap bypassed", err)
	}
	var count int
	if err := f.db.db.QueryRow("SELECT count(*) FROM ai_translation_runs").Scan(&count); err != nil || count != 0 {
		t.Fatal("partial admission", count, err)
	}
}
