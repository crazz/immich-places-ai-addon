package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"immich-places-backend/internal/ai/writeback"
)

func TestAIMirrorRecoveryHTTPBindsOwnerTargetStepAndGeneration(t *testing.T) {
	for _, action := range []string{"retry", "reconcile"} {
		t.Run(action, func(t *testing.T) {
			f := mirrorWriteFixture(t)
			mirrorMutationServer(t, f)
			ctx := context.Background()
			f.handle = func(w http.ResponseWriter, r *http.Request) bool {
				if r.Method == "PUT" {
					f.mu.Lock()
					f.mirrorSends++
					f.mu.Unlock()
					status := 400
					if action == "reconcile" {
						status = 500
					}
					w.WriteHeader(status)
					return true
				}
				return false
			}
			op, err := f.writer.confirm(ctx, testUserID, writeback.Confirmation{PreviewID: f.preview.Plan.ID, Digest: f.preview.Digest, Key: action})
			if err != nil {
				t.Fatal(err)
			}
			for range 2 {
				if err = f.writer.runOne(ctx, testUserID, op.ID); err != nil {
					t.Fatal(err)
				}
			}
			if action == "reconcile" {
				for range 3 {
					f.f.now = f.f.now.Add(46 * time.Second)
					if err = f.writer.runOne(ctx, testUserID, op.ID); err != nil {
						t.Fatal(err)
					}
				}
			}
			op, err = f.writer.get(ctx, testUserID, op.ID, false)
			if err != nil {
				t.Fatal(err)
			}
			f.writer.drafts.results.origin = aiTestOrigin
			handler := newAIResultHandler(f.writer.drafts.results, f.image.service)
			path := "/ai/write-operations/" + op.ID + "/targets/" + f.draft.AssetID + "/metadata/" + action
			body := fmt.Sprintf(`{"generation":%d}`, op.Mirror.Generation)
			for _, tc := range []struct {
				path, body, origin string
				auth               bool
				want               int
			}{
				{path, body, aiTestOrigin, false, 401}, {path, body, "https://foreign.invalid", true, 403},
				{path, `{"generation":0}`, aiTestOrigin, true, 400}, {path, `{"generation":1,"step":"gps"}`, aiTestOrigin, true, 400},
				{"/ai/write-operations/" + op.ID + "/targets/" + selectionB + "/metadata/" + action, body, aiTestOrigin, true, 409},
			} {
				rec := aiRequest(handler, "POST", tc.path, tc.body, tc.origin, tc.auth)
				if rec.Code != tc.want {
					t.Fatalf("scope guard: got %d want %d", rec.Code, tc.want)
				}
			}
			rec := aiRequest(handler, "POST", path, body, aiTestOrigin, true)
			var next writeback.Operation
			if rec.Code != 200 || json.Unmarshal(rec.Body.Bytes(), &next) != nil || next.Mirror.Generation != op.Mirror.Generation+1 {
				t.Fatalf("metadata action failed: %d %s", rec.Code, rec.Body.String())
			}
			if f.standardSends != 1 || f.mirrorSends != 1 {
				t.Fatal("recovery action directly sent mutation")
			}
		})
	}
}
