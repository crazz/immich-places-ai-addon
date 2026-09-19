package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"testing"
	"time"

	"immich-places-backend/internal/ai/providers"
)

func aiCapabilityTestHandler(t *testing.T, enabled bool, policy providers.EgressPolicy, transport providers.Transport) (*Database, http.Handler) {
	t.Helper()
	db := newTestDB(t)
	hash := sha256.Sum256([]byte("ai-session"))
	if err := db.createSession(context.Background(), hex.EncodeToString(hash[:]), testUserID, time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	dispatcher := newAIProviderDispatcher(db, enabled, policy, transport)
	return db, newAIProviderHandler(db, &Config{
		AIEnabled:              enabled,
		AIPublicOrigin:         aiTestOrigin,
		AIProviderEgressPolicy: policy,
	}, dispatcher)
}
