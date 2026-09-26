package writeback

import (
	"strings"
	"testing"
)

func TestCapabilityPolicyRequiresExactInstallationProfileEvidence(t *testing.T) {
	const installation = "caf12459-abcd-4321-8421-aaccff110033"
	const raw = `{"version":1,"installation":"` + installation + `","profile":"immich-v3.2.2","evidence":"docs/evidence/authorized-description-run.md","capabilities":["description"]}`
	policy, err := ParseCapabilityPolicy(raw)
	if err != nil || !policy.Allows(installation, "immich-v3.2.2", "description") {
		t.Fatal("exact attestation unavailable", err)
	}
	if policy.Allows("other", policy.Profile, "description") || policy.Allows(installation, "other", "description") || policy.Allows(installation, policy.Profile, "gps") || policy.Allows(installation, policy.Profile, "metadata") {
		t.Fatal("attestation expanded scope")
	}
	disabled, err := ParseCapabilityPolicy("")
	if err != nil || disabled.Allows(installation, policy.Profile, "description") {
		t.Fatal("capabilities did not default off")
	}
	identity := policy.Identity(installation, policy.Profile)
	if len(identity) != 64 || identity == disabled.Identity(installation, policy.Profile) || identity == policy.Identity("other", policy.Profile) {
		t.Fatal("policy identity omitted authority")
	}
	for _, invalid := range []string{
		`{}`, `null`, `[]`, raw + `{}`, strings.Replace(raw, `"version":1`, `"version":1,"version":1`, 1),
		strings.Replace(raw, `"description"`, `"description","description"`, 1), strings.Replace(raw, `"description"`, `"unknown"`, 1),
		strings.Replace(raw, `authorized-description-run.md`, strings.Repeat("a", 257), 1), strings.Replace(raw, `"evidence":`, `"unknown":`, 1),
		strings.Replace(raw, installation, "invalid", 1), strings.Replace(raw, `immich-v3.2.2`, `immich-v0.0.0`, 1),
	} {
		if _, err := ParseCapabilityPolicy(invalid); err == nil {
			t.Fatal("invalid capability policy accepted")
		}
	}
}
