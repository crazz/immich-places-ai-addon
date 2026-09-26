package main

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"
	"time"

	"immich-places-backend/internal/ai/writeback"
)

func TestAIStackReconcileHTTPFencesExactGenerationWithoutSending(t *testing.T) {
	f := stackWriteFixture(t)
	op := approveStackFixture(t, f, f.members)
	ctx := context.Background()
	asset := op.Plan.Manifest.Targets[1].AssetID
	step, err := writeback.TargetStep(op, asset)
	if err != nil {
		t.Fatal(err)
	}
	a := &aiStackAttempt{store: f.writer, parent: op, assetID: asset}
	if _, err := a.Read(ctx, step); err != nil {
		t.Fatal(err)
	}
	if err := a.Reserve(ctx, step, false); err != nil {
		t.Fatal(err)
	}
	if err := a.Sent(ctx, step, writeback.Completion{Known: false}); err != nil {
		t.Fatal(err)
	}
	for range 3 {
		f.f.now = f.f.now.Add(time.Minute)
		claimed, err := a.recover(ctx)
		if err != nil || !claimed {
			t.Fatal("could not exhaust bounded readbacks", err)
		}
		if err := a.Verify(ctx, step); err != nil {
			t.Fatal(err)
		}
	}
	saved, err := f.writer.get(ctx, testUserID, op.ID, false)
	if err != nil {
		t.Fatal(err)
	}
	generation := saved.Targets[1].Generation
	f.writer.drafts.results.origin = aiTestOrigin
	h := newAIResultHandler(f.writer.drafts.results, f.image.service)
	path := "/ai/write-operations/" + op.ID + "/targets/" + asset + "/reconcile"
	body := `{"generation":` + strconv.Itoa(generation) + `}`
	before, writes := f.image.counts()
	rec := aiRequest(h, "POST", path, body, aiTestOrigin, true)
	var returned writeback.Operation
	if rec.Code != 200 || json.Unmarshal(rec.Body.Bytes(), &returned) != nil || returned.Targets[1].Generation != generation+1 || returned.Targets[1].Attempts != 1 {
		t.Fatal("exact reconciliation request failed", rec.Code, rec.Body.String())
	}
	var reads int
	if err := f.f.db.db.QueryRow(`SELECT reads FROM ai_stack_write_targets WHERE operationID=? AND assetID=?`, op.ID, asset).Scan(&reads); err != nil || reads != 0 {
		t.Fatal("target read allowance was not renewed", reads, err)
	}
	if stale := aiRequest(h, "POST", path, body, aiTestOrigin, true); stale.Code != 409 {
		t.Fatal("stale generation changed target", stale.Code)
	}
	if foreign := aiRequest(h, "POST", path, body, "https://foreign.invalid", true); foreign.Code != 403 {
		t.Fatal("origin guard bypassed")
	}
	if noAuth := aiRequest(h, "POST", path, body, aiTestOrigin, false); noAuth.Code != 401 {
		t.Fatal("authentication bypassed")
	}
	after, laterWrites := f.image.counts()
	if before != after || writes != 0 || laterWrites != 0 {
		t.Fatal("reconciliation admission used remote I/O")
	}
}
