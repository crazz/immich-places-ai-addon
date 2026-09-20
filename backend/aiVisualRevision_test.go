package main

import (
	"context"
	"testing"
)

func TestAIVisualExplicitJSONAndHistoricalRevisionRemainExact(t *testing.T) {
	for _, mode := range []string{"json", "historical"} {
		t.Run(mode, func(t *testing.T) {
			f := newAIVisualFixture(t)
			if mode == "json" {
				f.request.Format = "json"
				f.request.AllowJSON = true
			} else {
				for _, query := range []string{`INSERT INTO ai_provider_versions (userID,profileID,revision,name,baseURL,model,secretCiphertext) SELECT userID,profileID,2,'New',baseURL,'different-model',secretCiphertext FROM ai_provider_versions WHERE revision=1`, `UPDATE ai_provider_profiles SET activeRevision=2`} {
					if _, err := f.image.db.db.Exec(query); err != nil {
						t.Fatal(err)
					}
				}
			}
			result, err := f.analyzer.analyze(context.Background(), f.request)
			if err != nil || result == nil || result.Info().Model != "bound-model" || result.Info().Revision != 1 || f.hits.Load() != 1 {
				t.Fatal("bound revision was changed", err)
			}
			if _, err := f.image.db.db.Exec(`UPDATE ai_provider_profiles SET enabled=0`); err != nil {
				t.Fatal(err)
			}
			if result, err := f.analyzer.analyze(context.Background(), f.request); err == nil || result != nil || f.hits.Load() != 1 {
				t.Fatal("disablement did not apply")
			}
		})
	}
}

func TestAIVisualRejectsJSONWithoutApplicableObservation(t *testing.T) {
	f := newAIVisualFixture(t)
	f.request.Format = "json"
	f.request.AllowJSON = true
	f.exec(`UPDATE ai_provider_capability_checks SET observationsJSON='{"image":{"status":"supported"},"json":{"status":"unverified"},"strict":{"status":"supported"}}'`)
	result, err := f.analyzer.analyze(context.Background(), f.request)
	if err == nil || result != nil || f.hits.Load() != 0 || f.reservations.Load() != 0 {
		t.Fatal("unobserved JSON format dispatched")
	}
}
