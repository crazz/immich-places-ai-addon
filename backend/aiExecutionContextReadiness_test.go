package main

import (
	"context"
	"encoding/json"
	"immich-places-backend/internal/ai/providers"
	"testing"
)

func TestAIExecutionReadinessDistinguishesContextAttestation(t *testing.T) {
	f, p, req := contextProductionFixture(t)
	h := &aiProviderHandlers{db: f.image.db, enabled: true, policy: f.analyzer.dispatcher.Policy, executionPolicies: p.policies}
	ready, err := h.executionReadiness(context.Background(), testUserID, aiProviderProfile{ID: req.Configuration.ProfileID, Revision: 1, Config: providers.Config{Model: "bound-model"}, Enabled: true})
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(ready)
	var value map[string]any
	_ = json.Unmarshal(raw, &value)
	if value["contextAllowed"] != true {
		t.Fatal("context readiness missing")
	}
	h.executionPolicies = attestedExecutionPolicy(t, f)
	ready, err = h.executionReadiness(context.Background(), testUserID, aiProviderProfile{ID: req.Configuration.ProfileID, Revision: 1, Config: providers.Config{Model: "bound-model"}, Enabled: true})
	if err != nil {
		t.Fatal(err)
	}
	raw, _ = json.Marshal(ready)
	_ = json.Unmarshal(raw, &value)
	if value["contextAllowed"] != false {
		t.Fatal("Visual policy advertised Context")
	}
	if f.hits.Load() != 0 {
		t.Fatal("readiness sent a probe")
	}
}
