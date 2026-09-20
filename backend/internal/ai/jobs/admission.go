package jobs

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/google/uuid"
	"immich-places-backend/internal/ai/results"
)

type Limits struct {
	MaxCalls           int    `json:"maxCalls"`
	MaxTokens          int64  `json:"maxTokens"`
	OutputTokens       int64  `json:"outputTokens"`
	MaxEstimatedMicros *int64 `json:"maxEstimatedMicros,omitempty"`
}
type Configuration struct {
	SelectionToken  string          `json:"selectionToken"`
	ProfileID       string          `json:"profileId"`
	Revision        int             `json:"revision"`
	Mode            string          `json:"mode"`
	Format          string          `json:"format"`
	AllowJSON       bool            `json:"allowJson"`
	Languages       []string        `json:"languages"`
	PrimaryLanguage string          `json:"primaryLanguage"`
	PolicyID        string          `json:"policyId"`
	Limits          Limits          `json:"limits"`
	Context         *ContextChoices `json:"context,omitempty"`
	Rerun           *RerunChoice    `json:"rerun,omitempty"`
}
type ImageConsent struct {
	Version       string        `json:"version"`
	Image         bool          `json:"image"`
	Configuration Configuration `json:"configuration"`
}
type Admission struct {
	Configuration  Configuration `json:"configuration"`
	Consent        ImageConsent  `json:"consent"`
	IdempotencyKey string        `json:"idempotencyKey"`
}

func NormalizeAdmission(req Admission) (Admission, string, error) {
	cfg, err := normalizeConfiguration(req.Configuration)
	if err != nil {
		return Admission{}, "", err
	}
	consent, err := normalizeConfiguration(req.Consent.Configuration)
	if err != nil || !req.Consent.Image || req.Consent.Version != "image-consent-v1" || !boundedIdentity(req.IdempotencyKey) {
		return Admission{}, "", ErrInvalid
	}
	a, _ := json.Marshal(cfg)
	b, _ := json.Marshal(consent)
	if string(a) != string(b) {
		return Admission{}, "", ErrInvalid
	}
	req.Configuration, req.Consent.Configuration = cfg, consent
	data, _ := json.Marshal(req)
	digest := sha256.Sum256(data)
	return req, hex.EncodeToString(digest[:]), nil
}
func normalizeConfiguration(cfg Configuration) (Configuration, error) {
	id, err := uuid.Parse(cfg.SelectionToken)
	if err != nil || !boundedIdentity(cfg.ProfileID) || cfg.Revision < 1 || (cfg.Mode != "visual" && cfg.Mode != "context-assisted") || (cfg.Format != "strict" && cfg.Format != "json") || (cfg.Format == "json" && !cfg.AllowJSON) {
		return Configuration{}, ErrInvalid
	}
	cfg.Context, err = normalizeContext(cfg.Mode, cfg.Context)
	if err != nil {
		return Configuration{}, err
	}
	if cfg.Rerun != nil {
		rerun := *cfg.Rerun
		parent, err := uuid.Parse(rerun.ParentJobID)
		if err != nil || (rerun.Kind != "retry-failed" && rerun.Kind != "reanalysis") {
			return Configuration{}, ErrInvalid
		}
		rerun.ParentJobID = parent.String()
		cfg.Rerun = &rerun
	}
	cfg.SelectionToken = id.String()
	if digest, err := hex.DecodeString(cfg.PolicyID); err != nil || len(digest) != 32 {
		return Configuration{}, ErrInvalid
	}
	cfg.PolicyID = stringLowerHex(cfg.PolicyID)
	cfg.Languages, cfg.PrimaryLanguage, err = results.NormalizeLanguages(cfg.Languages, cfg.PrimaryLanguage)
	if err != nil {
		return Configuration{}, ErrInvalid
	}
	l := cfg.Limits
	if l.MaxCalls < 0 || l.MaxCalls > 1500 || l.MaxTokens < 1 || l.MaxTokens > 1_000_000_000_000 || l.OutputTokens < 1 || l.OutputTokens > 1_000_000_000 || (l.MaxEstimatedMicros != nil && (*l.MaxEstimatedMicros < 0 || *l.MaxEstimatedMicros > 1_000_000_000_000_000)) {
		return Configuration{}, ErrInvalid
	}
	return cfg, nil
}
func stringLowerHex(value string) string {
	data, _ := hex.DecodeString(value)
	return hex.EncodeToString(data)
}
func (p ExecutionPolicy) Allows(l Limits) bool {
	return p.Valid() && l.OutputTokens <= p.MaxOutputTokens && l.MaxTokens >= p.MaxInputTokens+l.OutputTokens && (l.MaxEstimatedMicros == nil || p.InputMicrosPerMillion != nil)
}
