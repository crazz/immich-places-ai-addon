package main

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"

	"immich-places-backend/internal/ai/writeback"
)

func TestAIStackHTTPRequiresExactTargetForRetryAndPreservesAcceptedReplay(t *testing.T) {
	f := stackWriteFixture(t)
	op := approveStackFixture(t, f, f.members)
	asset := op.Plan.Manifest.Targets[1].AssetID
	step, err := writeback.TargetStep(op, asset)
	if err != nil {
		t.Fatal(err)
	}
	a := &aiStackAttempt{store: f.writer, parent: op, assetID: asset}
	ctx := context.Background()
	if _, err := a.Read(ctx, step); err != nil {
		t.Fatal(err)
	}
	if err := a.Reserve(ctx, step, false); err != nil {
		t.Fatal(err)
	}
	if err := a.Sent(ctx, step, writeback.Completion{Known: true}); err != nil {
		t.Fatal(err)
	}
	if err := a.Verify(ctx, step); err != nil {
		t.Fatal(err)
	}
	saved, err := f.writer.get(ctx, testUserID, op.ID, false)
	if err != nil {
		t.Fatal(err)
	}
	f.writer.drafts.results.origin = aiTestOrigin
	h := newAIResultHandler(f.writer.drafts.results, f.image.service)
	body := `{"generation":` + strconv.Itoa(saved.Targets[1].Generation) + `}`
	for _, action := range []string{"retry", "reconcile"} {
		requestBody := body
		if action == "reconcile" {
			requestBody = `{}`
		}
		rec := aiRequest(h, "POST", "/ai/write-operations/"+op.ID+"/"+action, requestBody, aiTestOrigin, true)
		if rec.Code != 409 {
			t.Fatal("aggregate action admitted stack or returned storage error", action, rec.Code)
		}
	}
	path := "/ai/write-operations/" + op.ID + "/targets/" + asset + "/retry"
	rec := aiRequest(h, "POST", path, body, aiTestOrigin, true)
	var accepted writeback.Operation
	if rec.Code != 200 || json.Unmarshal(rec.Body.Bytes(), &accepted) != nil || accepted.Targets[1].Status != "queued" {
		t.Fatal("exact HTTP retry rejected", rec.Code, rec.Body.String())
	}
	f.writer.enabled = func() bool { return false }
	before, writes := f.image.counts()
	if replay := aiRequest(h, "POST", path, body, aiTestOrigin, true); replay.Code != 200 {
		t.Fatal("accepted HTTP retry replay rejected", replay.Code)
	}
	for _, malformed := range []string{`{}`, `{"generation":-1}`, `{"generation":1,"targetIds":[]}`} {
		if invalid := aiRequest(h, "POST", path, malformed, aiTestOrigin, true); invalid.Code != 400 {
			t.Fatal("invalid retry body accepted", invalid.Code)
		}
	}
	after, laterWrites := f.image.counts()
	if before != after || writes != 0 || laterWrites != 0 {
		t.Fatal("replay/invalid request contacted upstream")
	}
}
