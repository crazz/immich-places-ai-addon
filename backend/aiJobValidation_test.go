package main

import (
	"context"
	"fmt"
	"testing"
)

func TestAIJobInvalidSubmissionLeavesNoPartialWork(t *testing.T) {
	cases := map[string]func(*aiJobFixture){
		"missing consent":  func(f *aiJobFixture) { f.input.ConsentVersion = "" },
		"missing key":      func(f *aiJobFixture) { f.input.Key = "" },
		"invalid digest":   func(f *aiJobFixture) { f.input.SelectionDigest = "wrong" },
		"empty membership": func(f *aiJobFixture) { f.input.AssetIDs = nil },
		"bad asset":        func(f *aiJobFixture) { f.input.AssetIDs = []string{"bad-id"} },
		"asset cap": func(f *aiJobFixture) {
			for i := 0; i < 501; i++ {
				f.input.AssetIDs = append(f.input.AssetIDs, fmt.Sprintf("00000000-0000-4000-8000-%012d", i))
			}
		},
		"zero call cap":      func(f *aiJobFixture) { f.input.MaxCalls = 0 },
		"excess call cap":    func(f *aiJobFixture) { f.input.MaxCalls = 7 },
		"invalid languages":  func(f *aiJobFixture) { f.input.Languages = []string{"en", "EN"} },
		"foreign owner":      func(f *aiJobFixture) { f.input.Owner = "foreign" },
		"foreign revision":   func(f *aiJobFixture) { f.input.Revision = 10 },
		"stale installation": func(f *aiJobFixture) { f.input.Installation = "other" },
		"disabled AI":        func(f *aiJobFixture) { f.store.enabled = false },
		"disabled profile":   func(f *aiJobFixture) { selectionSQL(t, f.db, "UPDATE ai_provider_profiles SET enabled=0") },
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			f := newAIJobFixture(t)
			change(f)
			if job, err := f.store.Submit(context.Background(), f.input); err == nil || job.ID != "" {
				t.Fatal("invalid job accepted")
			}
			var count int
			if err := f.db.db.QueryRow("SELECT count(*) FROM ai_jobs").Scan(&count); err != nil || count != 0 {
				t.Fatal("partial work", count, err)
			}
		})
	}
}
