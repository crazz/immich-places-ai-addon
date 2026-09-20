package main

import (
	"context"
	"testing"
)

func TestAIVisualDenialsNeverDispatchPrivateImage(t *testing.T) {
	cases := map[string]func(*aiVisualFixture){
		"AI disabled":          func(f *aiVisualFixture) { f.image.store.enabled = false },
		"dispatcher disabled":  func(f *aiVisualFixture) { f.analyzer.dispatcher.Enabled = false },
		"foreign owner":        func(f *aiVisualFixture) { f.request.Owner = "other" },
		"missing revision":     func(f *aiVisualFixture) { f.request.Revision = 999 },
		"profile disabled":     func(f *aiVisualFixture) { f.exec(`UPDATE ai_provider_profiles SET enabled=0`) },
		"missing observations": func(f *aiVisualFixture) { f.exec(`DELETE FROM ai_provider_capability_checks`) },
		"stale policy": func(f *aiVisualFixture) {
			f.exec(`UPDATE ai_provider_capability_checks SET policyFingerprint='stale'`)
		},
		"stale protocol": func(f *aiVisualFixture) {
			f.exec(`UPDATE ai_provider_capability_checks SET protocolVersion='old'`)
		},
		"wrong model": func(f *aiVisualFixture) {
			f.exec(`UPDATE ai_provider_capability_checks SET requestedModel='other'`)
		},
		"running observation": func(f *aiVisualFixture) {
			f.exec(`UPDATE ai_provider_capability_checks SET lifecycle='running'`)
		},
		"unsupported image": func(f *aiVisualFixture) {
			f.exec(`UPDATE ai_provider_capability_checks SET observationsJSON='{"image":{"status":"unsupported"},"json":{"status":"supported"},"strict":{"status":"supported"}}'`)
		},
		"unsupported strict": func(f *aiVisualFixture) {
			f.exec(`UPDATE ai_provider_capability_checks SET observationsJSON='{"image":{"status":"supported"},"json":{"status":"supported"},"strict":{"status":"unsupported"}}'`)
		},
		"local hidden": func(f *aiVisualFixture) { f.exec(`UPDATE assets SET isHidden=1`) },
		"credential removed": func(f *aiVisualFixture) {
			if err := f.image.db.updateImmichAPIKey(context.Background(), testUserID, nil); err != nil {
				t.Fatal(err)
			}
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			f := newAIVisualFixture(t)
			mutate(f)
			result, err := f.analyzer.analyze(context.Background(), f.request)
			if err == nil || result != nil || f.hits.Load() != 0 || f.reservations.Load() != 0 {
				t.Fatal("denied input dispatched", err, f.hits.Load())
			}
		})
	}
}
