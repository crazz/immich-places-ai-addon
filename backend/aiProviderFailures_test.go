package main

import (
	"context"
	"errors"
	"sync"
	"testing"

	"immich-places-backend/internal/ai/capabilities"
)

func TestAIProviderFailedEditsPreserveRevision(t *testing.T) {
	for _, tc := range []string{"destination", "plaintext", "corrupt", "invalid", "revision"} {
		t.Run(tc, func(t *testing.T) {
			db := newTestDB(t)
			ctx := context.Background()
			if _, err := db.createAIProvider(ctx, testUserID, "profile", providerInput()); err != nil {
				t.Fatal(err)
			}
			input := providerInput()
			input.Secret = nil
			revision := 1
			switch tc {
			case "destination":
				input.BaseURL = "https://other.example/v1"
			case "plaintext", "corrupt":
				stored := "plaintext-secret"
				if tc == "corrupt" {
					stored = "enc:broken"
				}
				if _, err := db.db.ExecContext(ctx, "UPDATE ai_provider_versions SET secretCiphertext=?", stored); err != nil {
					t.Fatal(err)
				}
			case "invalid":
				input.Name = ""
			case "revision":
				revision = 0
			}
			if _, err := db.updateAIProvider(ctx, testUserID, "profile", revision, input); err == nil {
				t.Fatal("invalid edit succeeded")
			}
			profiles, err := db.listAIProviders(ctx, testUserID, capabilities.ApplicabilityContext{})
			if err != nil || len(profiles) != 1 || profiles[0].Revision != 1 {
				t.Fatalf("failed edit changed current revision: %+v, error = %v", profiles, err)
			}
			var versions int
			if err := db.db.QueryRowContext(ctx, "SELECT count(*) FROM ai_provider_versions").Scan(&versions); err != nil || versions != 1 {
				t.Fatalf("failed edit left %d versions, error = %v", versions, err)
			}
		})
	}
}

func TestAIProviderConcurrentEditsHaveOneWinner(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	if _, err := db.createAIProvider(ctx, testUserID, "profile", providerInput()); err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	results := make(chan error, 2)
	var wg sync.WaitGroup
	for _, name := range []string{"first", "second"} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			input := providerInput()
			input.Name = name
			<-start
			_, err := db.updateAIProvider(ctx, testUserID, "profile", 1, input)
			results <- err
		}()
	}
	close(start)
	wg.Wait()
	close(results)
	succeeded, conflicted := 0, 0
	for err := range results {
		if err == nil {
			succeeded++
		} else if errors.Is(err, errAIProviderConflict) {
			conflicted++
		} else {
			t.Fatalf("unexpected concurrent edit error: %v", err)
		}
	}
	if succeeded != 1 || conflicted != 1 {
		t.Fatalf("successes = %d, conflicts = %d", succeeded, conflicted)
	}
}
