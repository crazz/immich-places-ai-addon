package main

import (
	"context"
	"testing"
	"time"

	"immich-places-backend/internal/ai/jobs"
	"immich-places-backend/internal/ai/selection"
)

func productionFixture(t *testing.T) (*aiVisualFixture, *aiProductionJobs, jobs.Admission) {
	t.Helper()
	f := newAIVisualFixture(t)
	p := &aiProductionJobs{store: newAIJobStore(f.image.db, f.image.store.binding, true, time.Now), selections: f.image.store, policies: attestedExecutionPolicy(t, f), fingerprint: policyFingerprint(f.analyzer.dispatcher.Policy)}
	snapshot, err := p.selections.preview(context.Background(), selection.Input{Mode: "explicit", AssetIDs: []string{selectionA}, Scope: &selection.Scope{View: "all"}}, testUserID)
	if err != nil {
		t.Fatal(err)
	}
	_, policyID, ok := p.policies.Find(jobs.ExecutionBinding{Owner: testUserID, Installation: p.store.binding, Profile: f.request.ProfileID, Revision: 1, Model: "bound-model", EgressFingerprint: p.fingerprint})
	if !ok {
		t.Fatal("missing fixture policy")
	}
	cfg := jobs.Configuration{SelectionToken: *snapshot.SnapshotID, ProfileID: f.request.ProfileID, Revision: 1, Mode: "visual", Format: "strict", Languages: []string{"en"}, PrimaryLanguage: "en", PolicyID: policyID, Limits: jobs.Limits{MaxCalls: 1, MaxTokens: 104000, OutputTokens: 4000}}
	return f, p, jobs.Admission{Configuration: cfg, Consent: jobs.ImageConsent{Version: "image-consent-v1", Image: true, Configuration: cfg}, IdempotencyKey: "production-one"}
}
func TestAIProductionAdmissionIsDurableWithoutUpstreamWork(t *testing.T) {
	f, p, req := productionFixture(t)
	before, _ := f.image.counts()
	job, err := p.submit(context.Background(), testUserID, req)
	if err != nil || job.ID == "" {
		t.Fatalf("admission unavailable: %v", err)
	}
	if len(job.Items) != 1 || job.Items[0].Asset != selectionA || job.Items[0].State != "queued" || job.Input.SelectionDigest == "" {
		t.Fatal("membership/configuration missing", job)
	}
	var count int
	if err := f.image.db.db.QueryRow("SELECT count(*) FROM ai_job_admissions WHERE userID=? AND jobID=?", testUserID, job.ID).Scan(&count); err != nil || count != 1 {
		t.Fatal("missing durable admission", count, err)
	}
	after, _ := f.image.counts()
	if f.hits.Load() != 0 || after != before {
		t.Fatal("admission contacted upstream")
	}
}
