package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"immich-places-backend/internal/ai/capabilities"
	"immich-places-backend/internal/ai/providers"
	"immich-places-backend/internal/aiadapters/providerhttp"
)

func TestPersistAndReloadPrivateObservation(t *testing.T) {
	hits := &atomic.Int32{}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	proxy := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capabilityFixtureResponse(w, r, hits)
	}))
	proxy.Listener = listener
	proxy.Start()
	t.Cleanup(proxy.Close)
	base, err := url.Parse(proxy.URL)
	if err != nil {
		t.Fatal(err)
	}
	approvedURL := "http://codex-proxy:" + base.Port() + "/v1"
	policy, err := providers.ParseEgressPolicy(`[{"baseURL":"` + approvedURL + `","addressClass":"local","allowedCIDRs":["127.0.0.0/8"]}]`)
	if err != nil {
		t.Fatal(err)
	}
	transport := providerhttp.New(providerhttp.Options{
		LookupIP: func(context.Context, string) ([]net.IP, error) {
			return []net.IP{net.ParseIP("127.0.0.1")}, nil
		},
	})

	dataDir := t.TempDir()
	db, err := newDatabase(dataDir, "persist-key")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.createUser(context.Background(), testUserID, "test@example.com", "hashed"); err != nil {
		t.Fatal(err)
	}
	if err := db.createUser(context.Background(), "other", "other@example.com", "hashed"); err != nil {
		t.Fatal(err)
	}
	hash := sha256.Sum256([]byte("ai-session"))
	if err := db.createSession(context.Background(), hex.EncodeToString(hash[:]), testUserID, time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	otherHash := sha256.Sum256([]byte("other-session"))
	if err := db.createSession(context.Background(), hex.EncodeToString(otherHash[:]), "other", time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	dispatcher := newAIProviderDispatcher(db, true, policy, transport)
	handler := newAIProviderHandler(db, &Config{AIEnabled: true, AIPublicOrigin: aiTestOrigin, AIProviderEgressPolicy: policy}, dispatcher)
	secret := "persist-secret"
	profile := mustCreateProvider(t, handler, providers.Input{
		Config:  providers.Config{Name: "Owned", BaseURL: approvedURL, Model: "gpt-5.6-sol"},
		Enabled: true,
		Secret:  &secret,
	})
	tested := aiRequest(handler, "POST", "/ai/providers/"+profile.ID+"/test", mustJSON(t, map[string]any{"expectedRevision": profile.Revision}), aiTestOrigin, true)
	var original struct {
		AttemptID    string `json:"attemptID"`
		Revision     int    `json:"revision"`
		Observations struct {
			Image struct {
				Status string `json:"status"`
			} `json:"image"`
		} `json:"observations"`
	}
	if err := json.Unmarshal(tested.Body.Bytes(), &original); err != nil || tested.Code != 200 || original.Observations.Image.Status != "supported" {
		t.Fatalf("initial test = %s err=%v", tested.Body.String(), err)
	}
	afterTestHits := hits.Load()
	db.close()

	reopened, err := newDatabase(dataDir, "persist-key")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { reopened.close() })
	reopenedHandler := newAIProviderHandler(reopened, &Config{AIEnabled: true, AIPublicOrigin: aiTestOrigin, AIProviderEgressPolicy: policy}, newAIProviderDispatcher(reopened, true, policy, transport))

	listed := aiRequest(reopenedHandler, "GET", "/ai/providers", "", "", true)
	var list struct {
		Items []struct {
			ID               string `json:"id"`
			Revision         int    `json:"revision"`
			CapabilityReport *struct {
				AttemptID    string `json:"attemptID"`
				Revision     int    `json:"revision"`
				Observations struct {
					Image struct {
						Status string `json:"status"`
					} `json:"image"`
				} `json:"observations"`
			} `json:"capabilityReport"`
		} `json:"items"`
	}
	if err := json.Unmarshal(listed.Body.Bytes(), &list); err != nil || listed.Code != 200 || len(list.Items) != 1 {
		t.Fatalf("reload list = %s err=%v", listed.Body.String(), err)
	}
	report := list.Items[0].CapabilityReport
	if report == nil || report.AttemptID != original.AttemptID || report.Revision != original.Revision || report.Observations.Image.Status != "supported" {
		t.Fatalf("reloaded report missing or changed: %+v", report)
	}
	if hits.Load() != afterTestHits {
		t.Fatal("reload triggered another provider call")
	}
	if strings.Contains(listed.Body.String(), "persist-secret") || strings.Contains(listed.Body.String(), "data:image/jpeg") {
		t.Fatal("credentials or image bytes exposed on reload")
	}

	foreignList := httptest.NewRequest("GET", "/ai/providers", nil)
	foreignList.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "other-session"})
	foreignRec := httptest.NewRecorder()
	reopenedHandler.ServeHTTP(foreignRec, foreignList)
	var foreign struct {
		Items []json.RawMessage `json:"items"`
	}
	if err := json.Unmarshal(foreignRec.Body.Bytes(), &foreign); err != nil || len(foreign.Items) != 0 {
		t.Fatalf("other user saw owner profiles: %s", foreignRec.Body.String())
	}
	foreignTest := httptest.NewRequest("POST", "/ai/providers/"+profile.ID+"/test", strings.NewReader(`{"expectedRevision":1}`))
	foreignTest.Header.Set("Content-Type", "application/json")
	foreignTest.Header.Set("Origin", aiTestOrigin)
	foreignTest.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "other-session"})
	foreignTestRec := httptest.NewRecorder()
	reopenedHandler.ServeHTTP(foreignTestRec, foreignTest)
	if foreignTestRec.Code != 404 {
		t.Fatalf("other user changed owner test: %d %s", foreignTestRec.Code, foreignTestRec.Body.String())
	}
	if hits.Load() != afterTestHits {
		t.Fatal("foreign access triggered provider call")
	}
}

func TestReloadingSettingsDoesNotRepeatCapabilityTest(t *testing.T) {
	hits := &atomic.Int32{}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	proxy := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capabilityFixtureResponse(w, r, hits)
	}))
	proxy.Listener = listener
	proxy.Start()
	t.Cleanup(proxy.Close)
	base, err := url.Parse(proxy.URL)
	if err != nil {
		t.Fatal(err)
	}
	approvedURL := "http://codex-proxy:" + base.Port() + "/v1"
	policy, err := providers.ParseEgressPolicy(`[{"baseURL":"` + approvedURL + `","addressClass":"local","allowedCIDRs":["127.0.0.0/8"]}]`)
	if err != nil {
		t.Fatal(err)
	}
	transport := providerhttp.New(providerhttp.Options{
		LookupIP: func(context.Context, string) ([]net.IP, error) {
			return []net.IP{net.ParseIP("127.0.0.1")}, nil
		},
	})
	_, handler := aiCapabilityTestHandler(t, true, policy, transport)
	secret := "secret"
	profile := mustCreateProvider(t, handler, providers.Input{
		Config:  providers.Config{Name: "Owned", BaseURL: approvedURL, Model: "gpt-5.6-sol"},
		Enabled: true,
		Secret:  &secret,
	})
	success := aiRequest(handler, "POST", "/ai/providers/"+profile.ID+"/test", mustJSON(t, map[string]any{"expectedRevision": 1}), aiTestOrigin, true)
	if success.Code != 200 {
		t.Fatalf("success test = %s", success.Body.String())
	}
	afterSuccess := hits.Load()

	for i := 0; i < 2; i++ {
		listed := aiRequest(handler, "GET", "/ai/providers", "", "", true)
		var list struct {
			Items []struct {
				CapabilityReport *struct {
					Observations struct {
						Image struct {
							Status string `json:"status"`
						} `json:"image"`
					} `json:"observations"`
				} `json:"capabilityReport"`
			} `json:"items"`
		}
		if err := json.Unmarshal(listed.Body.Bytes(), &list); err != nil || listed.Code != 200 || list.Items[0].CapabilityReport == nil || list.Items[0].CapabilityReport.Observations.Image.Status != "supported" {
			t.Fatalf("reload %d = %s err=%v", i, listed.Body.String(), err)
		}
	}
	if hits.Load() != afterSuccess {
		t.Fatal("settings reload repeated the capability test")
	}

	saved := aiRequest(handler, "PUT", "/ai/providers/"+profile.ID, mustJSON(t, aiProviderRequest{
		Input: providers.Input{
			Config:  providers.Config{Name: "Owned", BaseURL: approvedURL, Model: "gpt-5.6-sol"},
			Enabled: true,
		},
		ExpectedRevision: 1,
	}), aiTestOrigin, true)
	if saved.Code != 200 {
		t.Fatalf("save = %s", saved.Body.String())
	}
	if hits.Load() != afterSuccess {
		t.Fatal("settings save triggered a capability test")
	}
	listed := aiRequest(handler, "GET", "/ai/providers", "", "", true)
	var afterSave struct {
		Items []struct {
			Revision         int              `json:"revision"`
			CapabilityReport *json.RawMessage `json:"capabilityReport"`
		} `json:"items"`
	}
	if err := json.Unmarshal(listed.Body.Bytes(), &afterSave); err != nil || afterSave.Items[0].Revision != 2 || afterSave.Items[0].CapabilityReport != nil {
		t.Fatalf("new revision must not inherit old proof: %s", listed.Body.String())
	}
	if hits.Load() != afterSuccess {
		t.Fatal("listing new revision triggered a capability test")
	}
}

func TestRejectLatePublicationIntoNewerAttempt(t *testing.T) {
	db := newTestDB(t)
	profile, err := db.createAIProvider(context.Background(), testUserID, "profile", providers.Input{
		Config:  providers.Config{Name: "Owned", BaseURL: "https://provider.example/v1", Model: "vision"},
		Enabled: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	expiredStart := now.Add(-time.Hour)
	expiredDeadline := now.Add(-time.Minute)
	first, err := db.admitAIProviderCapability(context.Background(), testUserID, profile.ID, profile.Revision, "policy", expiredStart, expiredDeadline)
	if err != nil {
		t.Fatal(err)
	}
	second, err := db.admitAIProviderCapability(context.Background(), testUserID, profile.ID, profile.Revision, "policy", now, now.Add(2*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if first.AttemptID == second.AttemptID {
		t.Fatal("later attempt must receive a distinct attempt identity")
	}
	completed := now.Add(time.Second)
	late := capabilities.Report{
		AttemptID:         first.AttemptID,
		ProfileID:         profile.ID,
		Revision:          profile.Revision,
		ProtocolVersion:   capabilities.ProtocolVersion,
		PolicyFingerprint: "policy",
		Lifecycle:         "completed",
		StartedAt:         expiredStart,
		DeadlineAt:        expiredDeadline,
		CompletedAt:       &completed,
		RequestedModel:    "vision",
		Observations: capabilities.Observations{
			Image: capabilities.Observation{Status: capabilities.StatusSupported},
		},
		Compatibility: "incomplete",
		Applicable:    true,
	}
	if err := db.completeAIProviderCapability(context.Background(), testUserID, late); err == nil {
		t.Fatal("late completion for superseded attempt must be rejected")
	}
	loaded, err := db.loadAIProviderCapability(context.Background(), testUserID, profile.ID, profile.Revision, capabilities.ApplicabilityContext{
		AIEnabled: true, ProfileEnabled: true, ActiveRevision: profile.Revision,
		CurrentProtocolVersion: capabilities.ProtocolVersion, CurrentPolicyFingerprint: "policy",
	})
	if err != nil || loaded == nil || loaded.AttemptID != second.AttemptID || loaded.Lifecycle != "running" {
		t.Fatalf("newer attempt must remain current: %+v err=%v", loaded, err)
	}
	if loaded.Observations.Image.Status == capabilities.StatusSupported {
		t.Fatal("late observation must not become evidence for the newer attempt")
	}
}

func TestAccountDeletionRemovesCapabilityChecksAndRejectsLateCompletion(t *testing.T) {
	db := newTestDB(t)
	profile, err := db.createAIProvider(context.Background(), testUserID, "profile", providers.Input{
		Config:  providers.Config{Name: "Owned", BaseURL: "https://provider.example/v1", Model: "vision"},
		Enabled: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	started := time.Now().UTC()
	deadline := started.Add(2 * time.Minute)
	admission, err := db.admitAIProviderCapability(context.Background(), testUserID, profile.ID, profile.Revision, "policy", started, deadline)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.db.ExecContext(context.Background(), `DELETE FROM users WHERE ID=?`, testUserID); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.db.QueryRow(`SELECT count(*) FROM ai_provider_capability_checks`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("capability checks remain after account deletion: count=%d err=%v", count, err)
	}
	completed := started.Add(time.Second)
	late := capabilities.Report{
		AttemptID:         admission.AttemptID,
		ProfileID:         profile.ID,
		Revision:          profile.Revision,
		ProtocolVersion:   capabilities.ProtocolVersion,
		PolicyFingerprint: "policy",
		Lifecycle:         "completed",
		StartedAt:         started,
		DeadlineAt:        deadline,
		CompletedAt:       &completed,
		RequestedModel:    "vision",
		Observations: capabilities.Observations{
			Image: capabilities.Observation{Status: capabilities.StatusSupported},
		},
		Compatibility: "incomplete",
	}
	if err := db.completeAIProviderCapability(context.Background(), testUserID, late); err == nil {
		t.Fatal("late completion must not recreate deleted owner capability rows")
	}
	if err := db.db.QueryRow(`SELECT count(*) FROM ai_provider_capability_checks`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("late completion recreated capability rows: count=%d err=%v", count, err)
	}
}

func TestPolicyOrProtocolChangeMakesStoredReportInapplicableWithoutRetest(t *testing.T) {
	db := newTestDB(t)
	profile, err := db.createAIProvider(context.Background(), testUserID, "profile", providers.Input{
		Config:  providers.Config{Name: "Owned", BaseURL: "https://provider.example/v1", Model: "vision"},
		Enabled: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	started := time.Now().UTC()
	deadline := started.Add(2 * time.Minute)
	admission, err := db.admitAIProviderCapability(context.Background(), testUserID, profile.ID, profile.Revision, "policy-a", started, deadline)
	if err != nil {
		t.Fatal(err)
	}
	completed := started.Add(time.Second)
	report := capabilities.Report{
		AttemptID:         admission.AttemptID,
		ProfileID:         profile.ID,
		Revision:          profile.Revision,
		ProtocolVersion:   capabilities.ProtocolVersion,
		PolicyFingerprint: "policy-a",
		Lifecycle:         "completed",
		StartedAt:         started,
		DeadlineAt:        deadline,
		CompletedAt:       &completed,
		RequestedModel:    "vision",
		Observations: capabilities.Observations{
			Image: capabilities.Observation{Status: capabilities.StatusSupported},
		},
		Compatibility: "incomplete",
	}
	if err := db.completeAIProviderCapability(context.Background(), testUserID, report); err != nil {
		t.Fatal(err)
	}
	stalePolicy, err := db.loadAIProviderCapability(context.Background(), testUserID, profile.ID, profile.Revision, capabilities.ApplicabilityContext{
		AIEnabled: true, ProfileEnabled: true, ActiveRevision: profile.Revision,
		CurrentProtocolVersion: capabilities.ProtocolVersion, CurrentPolicyFingerprint: "policy-b",
	})
	if err != nil || stalePolicy == nil || stalePolicy.Applicable {
		t.Fatalf("policy change must invalidate current proof: %+v err=%v", stalePolicy, err)
	}
	staleProtocol, err := db.loadAIProviderCapability(context.Background(), testUserID, profile.ID, profile.Revision, capabilities.ApplicabilityContext{
		AIEnabled: true, ProfileEnabled: true, ActiveRevision: profile.Revision,
		CurrentProtocolVersion: "capability-v2", CurrentPolicyFingerprint: "policy-a",
	})
	if err != nil || staleProtocol == nil || staleProtocol.Applicable {
		t.Fatalf("protocol change must invalidate current proof: %+v err=%v", staleProtocol, err)
	}
}

func TestValidateBoundedConfigurationWithoutExternalCalls(t *testing.T) {
	hits := &atomic.Int32{}
	provider := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		hits.Add(1)
	}))
	t.Cleanup(provider.Close)
	policy, err := providers.ParseEgressPolicy(`[{"baseURL":"https://approved.example/v1","addressClass":"public"}]`)
	if err != nil {
		t.Fatal(err)
	}
	_, handler := aiCapabilityTestHandler(t, true, policy, nil)
	secret := "offline-secret"
	created := aiRequest(handler, "POST", "/ai/providers", mustJSON(t, providers.Input{
		Config:  providers.Config{Name: "Offline", BaseURL: provider.URL + "/v1", Model: "manual-model"},
		Enabled: true,
		Secret:  &secret,
	}), aiTestOrigin, true)
	var profile aiProviderProfile
	if err := json.Unmarshal(created.Body.Bytes(), &profile); err != nil || created.Code != 201 {
		t.Fatalf("create = %s err=%v", created.Body.String(), err)
	}
	edited := aiRequest(handler, "PUT", "/ai/providers/"+profile.ID, mustJSON(t, aiProviderRequest{
		Input: providers.Input{
			Config:  providers.Config{Name: "Offline", BaseURL: provider.URL + "/v1", Model: "manual-model"},
			Enabled: false,
		},
		ExpectedRevision: 1,
	}), aiTestOrigin, true)
	if edited.Code != 200 {
		t.Fatalf("edit/disable = %s", edited.Body.String())
	}
	listed := aiRequest(handler, "GET", "/ai/providers", "", "", true)
	if listed.Code != 200 {
		t.Fatalf("list = %s", listed.Body.String())
	}
	if hits.Load() != 0 {
		t.Fatalf("configuration operations made %d provider requests", hits.Load())
	}
}
