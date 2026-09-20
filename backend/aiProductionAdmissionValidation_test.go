package main

import (
	"context"
	"testing"

	"immich-places-backend/internal/ai/jobs"
)

func TestAIProductionRejectsIncompleteConsentAndAuthorityAtomically(t *testing.T) {
	cases := map[string]func(*aiVisualFixture, *aiProductionJobs, *jobs.Admission){
		"no-consent":      func(_ *aiVisualFixture, _ *aiProductionJobs, r *jobs.Admission) { r.Consent.Image = false },
		"consent-version": func(_ *aiVisualFixture, _ *aiProductionJobs, r *jobs.Admission) { r.Consent.Version = "future" },
		"changed-consent": func(_ *aiVisualFixture, _ *aiProductionJobs, r *jobs.Admission) {
			r.Consent.Configuration.Format = "json"
		},
		"missing-token-limit": func(_ *aiVisualFixture, _ *aiProductionJobs, r *jobs.Admission) {
			r.Configuration.Limits.MaxTokens = 0
			r.Consent.Configuration = r.Configuration
		},
		"output-limit": func(_ *aiVisualFixture, _ *aiProductionJobs, r *jobs.Admission) {
			r.Configuration.Limits.OutputTokens = 4001
			r.Consent.Configuration = r.Configuration
		},
		"token-budget": func(_ *aiVisualFixture, _ *aiProductionJobs, r *jobs.Admission) {
			r.Configuration.Limits.MaxTokens = 100000
			r.Consent.Configuration = r.Configuration
		},
		"cost-unknown": func(_ *aiVisualFixture, _ *aiProductionJobs, r *jobs.Admission) {
			n := int64(1)
			r.Configuration.Limits.MaxEstimatedMicros = &n
			r.Consent.Configuration = r.Configuration
		},
		"json-opt-in": func(_ *aiVisualFixture, _ *aiProductionJobs, r *jobs.Admission) {
			r.Configuration.Format = "json"
			r.Consent.Configuration = r.Configuration
		},
		"context": func(_ *aiVisualFixture, _ *aiProductionJobs, r *jobs.Admission) {
			r.Configuration.Mode = "context"
			r.Consent.Configuration = r.Configuration
		},
		"unsupported": func(f *aiVisualFixture, _ *aiProductionJobs, _ *jobs.Admission) {
			selectionSQL(t, f.image.db, "UPDATE ai_provider_capability_checks SET observationsJSON='{}'")
		},
		"revoked": func(f *aiVisualFixture, _ *aiProductionJobs, _ *jobs.Admission) {
			selectionSQL(t, f.image.db, "UPDATE ai_provider_profiles SET enabled=0")
		},
		"stale": func(f *aiVisualFixture, _ *aiProductionJobs, _ *jobs.Admission) {
			selectionSQL(t, f.image.db, "UPDATE assets SET isHidden=1")
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			f, p, req := productionFixture(t)
			mutate(f, p, &req)
			if job, err := p.submit(context.Background(), testUserID, req); err == nil || job.ID != "" {
				t.Fatal("invalid request admitted")
			}
			var n int
			if err := f.image.db.db.QueryRow("SELECT count(*) FROM ai_jobs").Scan(&n); err != nil || n != 0 {
				t.Fatal("partial admission", n, err)
			}
			if f.hits.Load() != 0 {
				t.Fatal("provider contacted")
			}
		})
	}
}
