package main

import (
	"context"
	"testing"
)

func TestAIProviderStorageFailureRollsBackRevisionAndErasure(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	input := providerInput()
	if _, err := db.createAIProvider(ctx, testUserID, "kept", input); err != nil {
		t.Fatal(err)
	}
	if _, err := db.db.Exec(`CREATE TRIGGER reject_provider_version BEFORE INSERT ON ai_provider_versions BEGIN SELECT RAISE(ABORT, 'synthetic storage failure'); END`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.createAIProvider(ctx, testUserID, "failed", input); err == nil {
		t.Fatal("failed version insert must not create a logical profile")
	}
	empty := ""
	input.Secret, input.Enabled = &empty, false
	if _, err := db.updateAIProvider(ctx, testUserID, "kept", 1, input); err == nil {
		t.Fatal("failed version insert must roll back the entire edit")
	}
	profiles, err := db.listAIProviders(ctx, testUserID)
	if err != nil || len(profiles) != 1 || profiles[0].ID != "kept" || profiles[0].Revision != 1 || !profiles[0].Enabled || !profiles[0].HasSecret {
		t.Fatalf("partial write survived rollback: %+v, %v", profiles, err)
	}
	var ciphertext string
	if err := db.db.QueryRow(`SELECT secretCiphertext FROM ai_provider_versions WHERE userID=? AND profileID='kept'`, testUserID).Scan(&ciphertext); err != nil {
		t.Fatal(err)
	}
	plaintext, err := decryptValue(db.encryptionKey, ciphertext)
	if err != nil || plaintext != *providerInput().Secret {
		t.Fatalf("failed erasure lost retained secret: %v", err)
	}
}
