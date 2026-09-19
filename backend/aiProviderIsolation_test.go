package main

import (
	"context"
	"testing"

	"immich-places-backend/internal/ai/capabilities"
)

func TestAIProviderIsolationUsesOwnerForSameProfileID(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	if err := db.createUser(ctx, "other", "other@example.com", "hashed"); err != nil {
		t.Fatal(err)
	}
	for _, owner := range []string{testUserID, "other"} {
		if _, err := db.createAIProvider(ctx, owner, "same-id", providerInput()); err != nil {
			t.Fatal(err)
		}
	}
	input := providerInput()
	input.Enabled, input.Secret = false, ptr("")
	if _, err := db.updateAIProvider(ctx, testUserID, "same-id", 1, input); err != nil {
		t.Fatal(err)
	}
	if _, err := db.db.ExecContext(ctx, "DELETE FROM users WHERE ID=?", testUserID); err != nil {
		t.Fatal(err)
	}
	other, err := db.listAIProviders(ctx, "other", capabilities.ApplicabilityContext{})
	if err != nil || len(other) != 1 || !other[0].Enabled || !other[0].HasSecret || other[0].Revision != 1 {
		t.Fatalf("another owner's identical ID was changed: %+v, %v", other, err)
	}
	var ciphertext string
	if err := db.db.QueryRowContext(ctx, `SELECT secretCiphertext FROM ai_provider_versions WHERE userID='other' AND profileID='same-id'`).Scan(&ciphertext); err != nil {
		t.Fatal(err)
	}
	plaintext, err := decryptValue(db.encryptionKey, ciphertext)
	if err != nil || plaintext != *providerInput().Secret {
		t.Fatalf("other owner's credential was removed or changed: %v", err)
	}
}
