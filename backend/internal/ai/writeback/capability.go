package writeback

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"slices"
	"strings"

	"github.com/google/uuid"
)

type CapabilityPolicy struct {
	Version      int      `json:"version"`
	Installation string   `json:"installation"`
	Profile      string   `json:"profile"`
	Evidence     string   `json:"evidence"`
	Capabilities []string `json:"capabilities"`
}

func ParseCapabilityPolicy(raw string) (CapabilityPolicy, error) {
	var p CapabilityPolicy
	if raw == "" {
		return p, nil
	}
	if len(raw) > 4096 {
		return p, Failure("WRITE_POLICY_INVALID")
	}
	decoder := json.NewDecoder(strings.NewReader(raw))
	if token, err := decoder.Token(); err != nil || token != json.Delim('{') {
		return p, Failure("WRITE_POLICY_INVALID")
	}
	fields := map[string]any{"version": &p.Version, "installation": &p.Installation, "profile": &p.Profile, "evidence": &p.Evidence, "capabilities": &p.Capabilities}
	for decoder.More() {
		token, err := decoder.Token()
		key, ok := token.(string)
		target := fields[key]
		if err != nil || !ok || target == nil || decoder.Decode(target) != nil {
			return p, Failure("WRITE_POLICY_INVALID")
		}
		delete(fields, key)
	}
	if token, err := decoder.Token(); err != nil || token != json.Delim('}') || len(fields) != 0 {
		return p, Failure("WRITE_POLICY_INVALID")
	}
	if _, err := decoder.Token(); err != io.EOF || !p.valid() {
		return p, Failure("WRITE_POLICY_INVALID")
	}
	return p, nil
}

func (p CapabilityPolicy) Allows(installation, profile, capability string) bool {
	return p.valid() && p.Installation == installation && p.Profile == profile && slices.Contains(p.Capabilities, capability)
}

func (p CapabilityPolicy) Identity(installation, profile string) string {
	raw, _ := json.Marshal(struct {
		Version      string
		Installation string
		Profile      string
		Policy       CapabilityPolicy
	}{"standard-write-policy-v1", installation, profile, p})
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func (p CapabilityPolicy) valid() bool {
	_, err := uuid.Parse(p.Installation)
	if err != nil || p.Version != 1 || p.Profile != "immich-v3.2.2" || len(p.Evidence) > 256 || strings.TrimSpace(p.Evidence) == "" || len(p.Capabilities) < 1 || len(p.Capabilities) > 3 {
		return false
	}
	for i, capability := range p.Capabilities {
		if (capability != "description" && capability != "stack_gps" && capability != "metadata") || slices.Contains(p.Capabilities[:i], capability) {
			return false
		}
	}
	return true
}
