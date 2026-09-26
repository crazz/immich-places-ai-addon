package main

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func TestAIStackPartialHistoryAndAuditSurviveCleanupAndDisabledReopen(t *testing.T) {
	f := stackWriteFixture(t)
	op := approveStackFixture(t, f, f.members)
	last := op.Plan.Manifest.Targets[2].AssetID
	f.metadata[last]["exifInfo"].(map[string]any)["latitude"] = 44
	m := stackMutationServer(t, f)
	ctx := context.Background()
	for range 3 {
		if err := f.writer.runOne(ctx, testUserID, op.ID); err != nil {
			t.Fatal(err)
		}
	}
	before, err := f.writer.get(ctx, testUserID, op.ID, false)
	if err != nil || before.Status != "partial" {
		t.Fatal("fixture did not reach partial", err)
	}
	analysis, err := f.f.store.ReadAnalysis(ctx, testUserID, f.draft.AnalysisID)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.f.store.Cancel(ctx, testUserID, analysis.JobID); err != nil {
		t.Fatal(err)
	}
	f.f.now = f.f.now.Add(48 * time.Hour)
	selectionSQL(t, f.f.db, `DELETE FROM assets WHERE userID=?`, testUserID)
	if purged, err := f.f.store.PurgeBefore(ctx, testUserID, f.f.now.Add(-time.Hour), 100); err != nil || purged != 0 {
		t.Fatal("retained stack audit purged", purged, err)
	}
	f.f.reopen(t)
	f.writer.drafts.results.jobs = f.f.store
	f.writer.enabled = func() bool { return false }
	after, err := f.writer.get(ctx, testUserID, op.ID, false)
	if err != nil || !reflect.DeepEqual(before, after) {
		t.Fatal("private stack audit changed", err)
	}
	h := newAIResultHandler(f.writer.drafts.results, f.image.service)
	rec := aiRequest(h, "GET", "/ai/write-operations?draftId="+f.draft.ID, "", "", true)
	var page aiWriteHistory
	if rec.Code != 200 || json.Unmarshal(rec.Body.Bytes(), &page) != nil || len(page.Items) != 1 || page.Items[0].ID != op.ID || page.Items[0].Status != "partial" || rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("history missing projected partial operation", rec.Code, rec.Body.String())
	}
	if err := f.writer.runOne(ctx, testUserID, op.ID); err != nil {
		t.Fatal(err)
	}
	if m.sent(last) != 0 || m.sent(op.Plan.Manifest.Targets[0].AssetID) != 1 || m.sent(op.Plan.Manifest.Targets[1].AssetID) != 1 {
		t.Fatal("history recovery resent completed work")
	}
}
