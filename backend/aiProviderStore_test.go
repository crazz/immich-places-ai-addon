package main

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"immich-places-backend/internal/ai/providers"
)

func providerInput() providers.Input {
	return providers.Input{Config: providers.Config{Name: "Private", BaseURL: "https://provider.example/v1", Model: "manual-model"}, Enabled: true, Secret: ptr("private-provider-secret")}
}

func TestAIProviderRevisionAndCredentialLifecycle(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	input := providerInput()
	if _, err := db.createAIProvider(ctx, testUserID, "profile", input); err != nil {
		t.Fatal(err)
	}
	input.Name, input.Secret = "Renamed", nil
	updated, err := db.updateAIProvider(ctx, testUserID, "profile", 1, input)
	if err != nil || updated.Revision != 2 || updated.Name != "Renamed" || !updated.HasSecret {
		t.Fatalf("edit = %+v, error = %v", updated, err)
	}
	if _, err := db.updateAIProvider(ctx, testUserID, "profile", 1, input); !errors.Is(err, errAIProviderConflict) {
		t.Fatalf("stale edit error = %v", err)
	}
	if _, err := db.updateAIProvider(ctx, "other", "profile", 2, input); !errors.Is(err, errAIProviderNotFound) {
		t.Fatalf("cross-user edit error = %v", err)
	}
	input.Secret, input.Enabled = ptr("replacement-secret"), false
	updated, err = db.updateAIProvider(ctx, testUserID, "profile", 2, input)
	if err != nil || updated.Revision != 3 || updated.Enabled {
		t.Fatalf("disable = %+v, error = %v", updated, err)
	}
	for revision, secret := range map[int]string{1: *providerInput().Secret, 2: *providerInput().Secret, 3: "replacement-secret"} {
		var stored, name string
		var enabled bool
		if err := db.db.QueryRowContext(ctx, `SELECT v.secretCiphertext, v.name, p.enabled FROM ai_provider_versions v JOIN ai_provider_profiles p ON p.userID=v.userID AND p.id=v.profileID WHERE v.userID=? AND v.profileID=? AND v.revision=?`, testUserID, "profile", revision).Scan(&stored, &name, &enabled); err != nil {
			t.Fatal(err)
		}
		plain, err := decryptValue(db.encryptionKey, stored)
		if err != nil || plain != secret || enabled || (revision == 1 && name != "Private") {
			t.Fatalf("revision %d mutated, enabled or credential lost", revision)
		}
	}
	if _, err := db.createAIProvider(ctx, testUserID, "unrelated", providerInput()); err != nil {
		t.Fatal(err)
	}
	input.Secret = ptr("")
	cleared, err := db.updateAIProvider(ctx, testUserID, "profile", 3, input)
	if err != nil || cleared.HasSecret || cleared.Revision != 4 {
		t.Fatalf("erase = %+v, error = %v", cleared, err)
	}
	var count int
	if err := db.db.QueryRowContext(ctx, `SELECT count(*) FROM ai_provider_versions WHERE userID=? AND profileID=? AND secretCiphertext IS NOT NULL`, testUserID, "profile").Scan(&count); err != nil || count != 0 {
		t.Fatalf("retained secret count = %d, error = %v", count, err)
	}
	if err := db.db.QueryRowContext(ctx, `SELECT count(*) FROM ai_provider_versions WHERE profileID='unrelated' AND secretCiphertext IS NOT NULL`).Scan(&count); err != nil || count != 1 {
		t.Fatalf("unrelated secret count = %d, error = %v", count, err)
	}
	if _, err := db.db.ExecContext(ctx, "DELETE FROM users WHERE ID=?", testUserID); err != nil {
		t.Fatal(err)
	}
	for _, table := range []string{"ai_provider_profiles", "ai_provider_versions"} {
		if err := db.db.QueryRowContext(ctx, "SELECT count(*) FROM "+table).Scan(&count); err != nil || count != 0 {
			t.Errorf("account deletion left %s: %d, error = %v", table, count, err)
		}
	}
}

func TestAIProviderCreateListAndReopen(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	if err := db.createUser(ctx, "other", "other@example.com", "hashed"); err != nil {
		t.Fatal(err)
	}
	created, err := db.createAIProvider(ctx, testUserID, "provider", providerInput())
	if err != nil || created.Revision != 1 || !created.Enabled || !created.HasSecret {
		t.Fatalf("create = %+v, error = %v", created, err)
	}
	var cipher string
	if err := db.db.QueryRowContext(ctx, "SELECT secretCiphertext FROM ai_provider_versions WHERE userID = ? AND profileID = ?", testUserID, "provider").Scan(&cipher); err != nil {
		t.Fatal(err)
	}
	plaintext, err := decryptValue(db.encryptionKey, cipher)
	if err != nil || plaintext != *providerInput().Secret || !strings.HasPrefix(cipher, encryptedPrefix) || strings.Contains(cipher, plaintext) {
		t.Fatal("stored credential is not valid ciphertext")
	}
	var seq int
	var name, dbPath string
	if err := db.db.QueryRowContext(ctx, "PRAGMA database_list").Scan(&seq, &name, &dbPath); err != nil {
		t.Fatal(err)
	}
	db.close()
	reopened, err := newDatabase(filepath.Dir(dbPath), "test-encryption-key")
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.close()
	for _, user := range []string{testUserID, "other"} {
		profiles, err := reopened.listAIProviders(ctx, user)
		if err != nil {
			t.Fatal(err)
		}
		if user == testUserID && (len(profiles) != 1 || profiles[0] != created) {
			t.Errorf("reopened profiles = %+v, want %+v", profiles, created)
		}
		if user == "other" && len(profiles) != 0 {
			t.Error("another user's profile leaked")
		}
		body, err := json.Marshal(profiles)
		if err != nil || strings.Contains(string(body), plaintext) || strings.Contains(string(body), cipher) {
			t.Fatal("public profile exposes credential material")
		}
	}
}
