package main

import (
	"context"
	"testing"
	"time"

	"immich-places-backend/internal/ai/jobs"
)

func TestAIJobHeartbeatRequiresExactUnexpiredLease(t *testing.T) {
	f := newAIJobFixture(t)
	ctx := context.Background()
	policy := jobs.DefaultPolicy()
	if _, err := f.store.Submit(ctx, f.input); err != nil {
		t.Fatal(err)
	}
	lease, ok, err := f.store.Claim(ctx, policy)
	if err != nil || !ok {
		t.Fatal(err)
	}
	if err = f.store.Authorize(ctx, lease); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"owner", "job", "item", "token", "asset", "installation"} {
		bad := lease
		switch field {
		case "owner":
			bad.Owner = "foreign"
		case "job":
			bad.JobID = "foreign"
		case "item":
			bad.ItemID = "foreign"
		case "token":
			bad.Token = "foreign"
		case "asset":
			bad.Asset = selectionB
		case "installation":
			bad.Installation = "foreign"
		}
		if f.store.Heartbeat(ctx, bad, policy) == nil || f.store.Authorize(ctx, bad) == nil || f.store.Reserve(ctx, bad) == nil {
			t.Fatal("foreign lease admitted", field)
		}
	}
	f.now = f.now.Add(30 * time.Second)
	if err = f.store.Heartbeat(ctx, lease, policy); err != nil {
		t.Fatal(err)
	}
	var expiry int64
	if err = f.db.db.QueryRow("SELECT leaseExpiresAt FROM ai_job_items WHERE id=?", lease.ItemID).Scan(&expiry); err != nil || expiry != f.now.Add(policy.LeaseDuration).UnixNano() {
		t.Fatal("lease not renewed", expiry, err)
	}
	selectionSQL(t, f.db, "UPDATE ai_job_items SET leaseToken='replacement' WHERE id=?", lease.ItemID)
	if f.store.Heartbeat(ctx, lease, policy) == nil || f.store.Authorize(ctx, lease) == nil {
		t.Fatal("old token admitted")
	}
	lease.Token = "replacement"
	f.now = time.Unix(0, expiry)
	if f.store.Heartbeat(ctx, lease, policy) == nil || f.store.Authorize(ctx, lease) == nil || f.store.Reserve(ctx, lease) == nil {
		t.Fatal("expired lease admitted")
	}
}
