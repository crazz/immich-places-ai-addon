package main

import (
	"context"
	"net/http"
	"testing"
	"time"

	"immich-places-backend/internal/ai/jobs"
)

func TestAIProductionFailuresKeepRetryAndBlockScopeExplicit(t *testing.T) {
	cases := []struct {
		name        string
		status      int
		body, state string
		image       bool
	}{
		{"provider-500", 500, `{}`, "retry_wait", false},
		{"provider-429", 429, `{}`, "retry_wait", false},
		{"provider-auth", 401, `{}`, "blocked", false},
		{"provider-model", 404, `{}`, "blocked", false},
		{"provider-request", 400, `{}`, "failed", false},
		{"invalid-output", 200, `{"choices":[]}`, "failed", false},
		{"image-transient", 500, `{}`, "retry_wait", true},
		{"image-inaccessible", 404, `{}`, "failed", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f, p, req := productionFixture(t)
			ctx := context.Background()
			req.Configuration.Limits.MaxCalls = 3
			req.Configuration.Limits.MaxTokens = 312000
			req.Consent.Configuration = req.Configuration
			handler := func(w http.ResponseWriter, r *http.Request) bool {
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
				return true
			}
			if tc.image {
				f.image.handle = handler
			} else {
				f.handle = handler
			}
			job, err := p.submit(ctx, testUserID, req)
			if err != nil {
				t.Fatal(err)
			}
			worker := jobs.Worker{Store: &aiProductionWorkerStore{aiJobStore: p.store, production: p}, Policy: jobs.DefaultPolicy(), Execute: p.executor(f.analyzer)}
			if _, err = worker.RunOne(ctx, make(chan time.Time)); err != nil {
				t.Fatal(err)
			}
			progress, err := p.progress(ctx, testUserID, job.ID)
			if err != nil || progress.Counts[tc.state] != 1 || progress.Blocked != (tc.state == "blocked") {
				t.Fatal("wrong failure scope", progress, err)
			}
			expected := int32(1)
			if tc.image {
				expected = 0
			}
			if f.hits.Load() != expected {
				t.Fatal("hidden retry", f.hits.Load())
			}
		})
	}
}
