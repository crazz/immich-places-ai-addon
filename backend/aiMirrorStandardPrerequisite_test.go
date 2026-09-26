package main

import (
	"context"
	"net/http"
	"testing"
	"time"

	"immich-places-backend/internal/ai/writeback"
)

func TestAIMirrorActualPartialConflictingAndUnknownStandardReadbackBlocksMetadata(t *testing.T) {
	for _, scenario := range []string{"partial", "conflict", "unknown"} {
		t.Run(scenario, func(t *testing.T) {
			f := mirrorWriteFixture(t)
			mirrorMutationServer(t, f)
			ctx := context.Background()
			f.handle = func(w http.ResponseWriter, r *http.Request) bool {
				if r.Method != "PATCH" {
					return false
				}
				f.mu.Lock()
				defer f.mu.Unlock()
				f.standardSends++
				fields := f.meta["exifInfo"].(map[string]any)
				switch scenario {
				case "partial":
					fields["latitude"] = float64(0)
					fields["longitude"] = float64(12)
				case "conflict":
					fields["latitude"] = float64(10)
					fields["longitude"] = float64(11)
					fields["description"] = "external text"
				case "unknown":
					w.WriteHeader(500)
					return true
				}
				w.WriteHeader(200)
				return true
			}
			op, err := f.writer.confirm(ctx, testUserID, writeback.Confirmation{PreviewID: f.preview.Plan.ID, Digest: f.preview.Digest, Key: "prerequisite"})
			if err != nil {
				t.Fatal(err)
			}
			for range 3 {
				if err = f.writer.runOne(ctx, testUserID, op.ID); err != nil {
					t.Fatal(err)
				}
				f.f.now = f.f.now.Add(46 * time.Second)
			}
			saved, err := f.writer.get(ctx, testUserID, op.ID, false)
			if err != nil || saved.Targets[0].Verified || saved.Mirror.Status != "blocked" || saved.Mirror.Attempts != 0 || f.mirrorSends != 0 || f.standardSends != 1 {
				t.Fatal("incomplete standard readback authorized metadata", saved.Targets, saved.Mirror, err)
			}
			target := saved.Targets[0]
			if scenario == "partial" && (target.Status != "partial" || !target.Refreshed || len(target.Fields) != 2 || !target.Fields[0].WasVerified || target.Fields[1].WasVerified) {
				t.Fatal("fixture did not produce independent partial fields", target)
			}
			if scenario == "conflict" && target.Status != "conflict" {
				t.Fatal("fixture did not produce standard conflict", target.Status)
			}
			if scenario == "unknown" && (target.Status != "verifying" || target.Settled) {
				t.Fatal("fixture did not retain unknown standard sender", target)
			}
		})
	}
}
