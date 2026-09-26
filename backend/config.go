package main

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"immich-places-backend/internal/ai/jobs"
	"immich-places-backend/internal/ai/providers"
	"immich-places-backend/internal/ai/writeback"
)

type Config struct {
	ImmichURL                string `env:"IMMICH_URL,notEmpty"`
	ImmichExternalURL        string `env:"IMMICH_EXTERNAL_URL"`
	Port                     int    `env:"PORT" envDefault:"8082"`
	DataDir                  string `env:"DATA_DIR" envDefault:"/data"`
	SyncIntervalMS           int    `env:"SYNC_INTERVAL_MS" envDefault:"300000"`
	TrustProxyTLS            bool   `env:"TRUST_PROXY_TLS" envDefault:"true"`
	AllowInsecure            bool   `env:"ALLOW_INSECURE" envDefault:"false"`
	RegistrationEnabled      bool   `env:"REGISTRATION_ENABLED" envDefault:"true"`
	EncryptionKey            string `env:"ENCRYPTION_KEY,notEmpty"`
	DawarichURL              string `env:"DAWARICH_URL"`
	DawarichSyncIntervalMS   int    `env:"DAWARICH_SYNC_INTERVAL_MS" envDefault:"86400000"`
	DefaultTimezone          string `env:"DEFAULT_TIMEZONE"`
	GeocodeProvider          string `env:"GEOCODE_PROVIDER" envDefault:"nominatim"`
	GeocodeAPIKey            string `env:"GEOCODE_API_KEY"`
	HereAPIKey               string `env:"HERE_API_KEY"`
	GoogleAPIKey             string `env:"GOOGLE_API_KEY"`
	GeocodeTimeoutSecs       int    `env:"GEOCODE_TIMEOUT" envDefault:"10"`
	NeighborWindowHours      int    `env:"SUGGESTIONS_NEIGHBOR_WINDOW_HOURS" envDefault:"6"`
	Debug                    bool   `env:"DEBUG" envDefault:"false"`
	AIWriteEnabled           bool   `env:"AI_WRITE_ENABLED" envDefault:"false"`
	AIWriteProfile           string `env:"AI_WRITE_PROFILE"`
	AIWriteCapabilitiesJSON  string `env:"AI_WRITE_CAPABILITIES"`
	AIWriteCapabilities      writeback.CapabilityPolicy
	AIEnabled                bool   `env:"AI_ENABLED" envDefault:"false"`
	AIPublicOrigin           string `env:"AI_PUBLIC_ORIGIN"`
	AIEgressPolicyJSON       string `env:"AI_PROVIDER_EGRESS_POLICY"`
	AIProviderEgressPolicy   providers.EgressPolicy
	AIExecutionPoliciesJSON  string `env:"AI_EXECUTION_POLICIES"`
	AIExecutionPolicies      jobs.ExecutionPolicies
	AISelectionMaxAssets     int `env:"AI_SELECTION_MAX_ASSETS" envDefault:"500"`
	AISelectionTTLSeconds    int `env:"AI_SELECTION_TTL_SECONDS" envDefault:"900"`
	AIJobWorkers             int `env:"AI_JOB_WORKERS" envDefault:"2"`
	AIJobPerOwner            int `env:"AI_JOB_PER_OWNER" envDefault:"1"`
	AIJobLeaseSeconds        int `env:"AI_JOB_LEASE_SECONDS" envDefault:"180"`
	AIJobHeartbeatSeconds    int `env:"AI_JOB_HEARTBEAT_SECONDS" envDefault:"30"`
	AIJobIdleMS              int `env:"AI_JOB_IDLE_MS" envDefault:"1000"`
	AIJobSettings            jobs.ConsumerSettings
	AIResearchTimeoutSeconds int    `env:"AI_RESEARCH_TIMEOUT_SECONDS" envDefault:"600"`
	AIInstanceEpoch          string `env:"AI_INSTANCE_EPOCH" envDefault:"1"`

	defaultTimezoneLocation *time.Location
}

func loadConfig() (*Config, error) {
	godotenv.Load()

	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return nil, err
	}

	if cfg.ImmichExternalURL == "" {
		cfg.ImmichExternalURL = cfg.ImmichURL
	}
	cfg.AIWriteCapabilities, err = writeback.ParseCapabilityPolicy(cfg.AIWriteCapabilitiesJSON)
	if err != nil {
		return nil, fmt.Errorf("AI_WRITE_CAPABILITIES is invalid; provide an installation/profile capability attestation")
	}
	cfg.AIExecutionPolicies, err = jobs.ParseExecutionPolicies(cfg.AIExecutionPoliciesJSON)
	if err != nil {
		return nil, fmt.Errorf("AI_EXECUTION_POLICIES is invalid; provide complete revision-bound execution attestations")
	}

	if cfg.AIJobLeaseSeconds < 1 || cfg.AIJobLeaseSeconds > 600 || cfg.AIJobHeartbeatSeconds < 1 || cfg.AIJobHeartbeatSeconds > 60 || cfg.AIJobIdleMS < 100 || cfg.AIJobIdleMS > 30000 || cfg.AIResearchTimeoutSeconds < 1 || cfg.AIResearchTimeoutSeconds > 600 {
		return nil, fmt.Errorf("AI job worker timing settings are outside finite bounds")
	}
	cfg.AIJobSettings = jobs.ConsumerSettings{Policy: jobs.Policy{Global: cfg.AIJobWorkers, PerOwner: cfg.AIJobPerOwner, LeaseDuration: time.Duration(cfg.AIJobLeaseSeconds) * time.Second}, Heartbeat: time.Duration(cfg.AIJobHeartbeatSeconds) * time.Second, Idle: time.Duration(cfg.AIJobIdleMS) * time.Millisecond}
	cfg.AIJobSettings.Policy.ResearchDuration = time.Duration(cfg.AIResearchTimeoutSeconds) * time.Second
	if !cfg.AIJobSettings.Valid() {
		return nil, fmt.Errorf("AI job worker concurrency or lease settings are invalid")
	}
	if cfg.SyncIntervalMS <= 0 {
		return nil, fmt.Errorf("SYNC_INTERVAL_MS must be > 0, got %d", cfg.SyncIntervalMS)
	}

	if cfg.GeocodeTimeoutSecs <= 0 {
		return nil, fmt.Errorf("GEOCODE_TIMEOUT must be > 0, got %d", cfg.GeocodeTimeoutSecs)
	}

	if cfg.NeighborWindowHours <= 0 {
		return nil, fmt.Errorf("SUGGESTIONS_NEIGHBOR_WINDOW_HOURS must be > 0, got %d", cfg.NeighborWindowHours)
	}

	if !cfg.TrustProxyTLS && !cfg.AllowInsecure {
		return nil, fmt.Errorf("TRUST_PROXY_TLS is false and no TLS is configured; set ALLOW_INSECURE=true to run without TLS")
	}

	if cfg.DefaultTimezone != "" {
		loc, err := time.LoadLocation(cfg.DefaultTimezone)
		if err != nil {
			return nil, fmt.Errorf("invalid DEFAULT_TIMEZONE %q: %w", cfg.DefaultTimezone, err)
		}
		cfg.defaultTimezoneLocation = loc
	}
	if cfg.AISelectionMaxAssets < 1 || cfg.AISelectionMaxAssets > 5000 {
		return nil, fmt.Errorf("AI_SELECTION_MAX_ASSETS must be between 1 and 5000")
	}
	if cfg.AISelectionTTLSeconds < 60 || cfg.AISelectionTTLSeconds > 3600 {
		return nil, fmt.Errorf("AI_SELECTION_TTL_SECONDS must be between 60 and 3600")
	}
	if strings.TrimSpace(cfg.AIInstanceEpoch) == "" || len(cfg.AIInstanceEpoch) > 128 {
		return nil, fmt.Errorf("AI_INSTANCE_EPOCH must contain 1–128 bytes")
	}
	if cfg.AIEnabled {
		origin, err := url.Parse(cfg.AIPublicOrigin)
		if err != nil || origin.Hostname() == "" || (origin.Scheme != "https" && origin.Scheme != "http") || origin.User != nil || origin.Path != "" || origin.RawQuery != "" || origin.ForceQuery || origin.Fragment != "" {
			return nil, fmt.Errorf("AI_PUBLIC_ORIGIN must be an HTTP(S) origin without credentials, path, query or fragment when AI_ENABLED=true")
		}
		policy, err := providers.ParseEgressPolicy(cfg.AIEgressPolicyJSON)
		if err != nil {
			return nil, fmt.Errorf("AI_PROVIDER_EGRESS_POLICY is invalid; fix the installation allowlist and restart")
		}
		cfg.AIProviderEgressPolicy = policy
	}

	return &cfg, nil
}
