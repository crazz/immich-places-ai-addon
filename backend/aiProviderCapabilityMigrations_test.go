package main

import (
	"context"
	"testing"
)

func TestAIProviderCapabilityStorageMigration(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	var count int
	if err := db.db.QueryRowContext(ctx, "SELECT count(*) FROM ai_provider_capability_checks").Scan(&count); err != nil {
		t.Fatalf("capability checks table unavailable: %v", err)
	}
	if count != 0 {
		t.Errorf("new capability checks table contains %d rows", count)
	}
}
