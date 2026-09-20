package jobs

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"strings"

	"github.com/google/uuid"
)

type ExecutionBinding struct {
	Owner             string `json:"owner"`
	Installation      string `json:"installation"`
	Profile           string `json:"profile"`
	Revision          int    `json:"revision"`
	Model             string `json:"model"`
	EgressFingerprint string `json:"egressFingerprint"`
}

type ExecutionPolicy struct {
	Version                string           `json:"version"`
	Binding                ExecutionBinding `json:"binding"`
	OutputField            string           `json:"outputField"`
	MaxInputTokens         int64            `json:"maxInputTokens"`
	MaxOutputTokens        int64            `json:"maxOutputTokens"`
	MaxRequestBytes        int              `json:"maxRequestBytes"`
	MaxImageBytes          int              `json:"maxImageBytes"`
	EvidenceRef            string           `json:"evidenceRef"`
	Currency               string           `json:"currency,omitempty"`
	InputMicrosPerMillion  *int64           `json:"inputMicrosPerMillion,omitempty"`
	OutputMicrosPerMillion *int64           `json:"outputMicrosPerMillion,omitempty"`
}

type ExecutionPolicies struct{ entries []ExecutionPolicy }

func ParseExecutionPolicies(raw string) (ExecutionPolicies, error) {
	var policies ExecutionPolicies
	if len(raw) > 128<<10 {
		return policies, ErrInvalid
	}
	if strings.TrimSpace(raw) == "" {
		return policies, nil
	}
	if !unambiguousExecutionJSON(raw) {
		return ExecutionPolicies{}, ErrInvalid
	}
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&policies.entries) != nil || policies.entries == nil || len(policies.entries) > 128 {
		return ExecutionPolicies{}, ErrInvalid
	}
	if decoder.Decode(new(any)) != io.EOF {
		return ExecutionPolicies{}, ErrInvalid
	}
	seen := map[ExecutionBinding]bool{}
	for _, p := range policies.entries {
		if !p.Valid() || seen[p.Binding] {
			return ExecutionPolicies{}, ErrInvalid
		}
		seen[p.Binding] = true
	}
	return policies, nil
}

func (p ExecutionPolicy) Valid() bool {
	if p.Version != "execution-v1" || p.Binding.Revision < 1 || (p.OutputField != "max_tokens" && p.OutputField != "max_completion_tokens") {
		return false
	}
	for _, value := range []string{p.Binding.Owner, p.Binding.Profile, p.Binding.Model, p.EvidenceRef} {
		if !boundedIdentity(value) || strings.TrimSpace(value) == "" {
			return false
		}
	}
	id, err := uuid.Parse(p.Binding.Installation)
	if err != nil || id.String() != p.Binding.Installation {
		return false
	}
	digest, err := hex.DecodeString(p.Binding.EgressFingerprint)
	if err != nil || len(digest) != 32 {
		return false
	}
	if p.MaxInputTokens < 1 || p.MaxInputTokens > 1000000000 || p.MaxOutputTokens < 1 || p.MaxOutputTokens > 1000000000 || p.MaxRequestBytes < 1 || p.MaxRequestBytes > 15<<20 || p.MaxImageBytes < 1 || p.MaxImageBytes > 10<<20 {
		return false
	}
	if p.Currency == "" && p.InputMicrosPerMillion == nil && p.OutputMicrosPerMillion == nil {
		return true
	}
	if len(p.Currency) != 3 || p.InputMicrosPerMillion == nil || p.OutputMicrosPerMillion == nil {
		return false
	}
	for _, c := range p.Currency {
		if c < 'A' || c > 'Z' {
			return false
		}
	}
	return *p.InputMicrosPerMillion >= 0 && *p.InputMicrosPerMillion <= 1000000000 && *p.OutputMicrosPerMillion >= 0 && *p.OutputMicrosPerMillion <= 1000000000
}

func (p ExecutionPolicies) Find(binding ExecutionBinding) (ExecutionPolicy, string, bool) {
	for _, policy := range p.entries {
		if policy.Binding == binding {
			if policy.InputMicrosPerMillion != nil {
				input, output := *policy.InputMicrosPerMillion, *policy.OutputMicrosPerMillion
				policy.InputMicrosPerMillion, policy.OutputMicrosPerMillion = &input, &output
			}
			data, _ := json.Marshal(policy)
			digest := sha256.Sum256(data)
			return policy, hex.EncodeToString(digest[:]), true
		}
	}
	return ExecutionPolicy{}, "", false
}

func (ExecutionPolicy) String() string     { return "attested AI execution policy" }
func (ExecutionPolicy) GoString() string   { return "attested AI execution policy" }
func (ExecutionPolicies) String() string   { return "AI execution policies" }
func (ExecutionPolicies) GoString() string { return "AI execution policies" }
