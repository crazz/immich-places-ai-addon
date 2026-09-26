package writeback

import "testing"

func TestMetadataCapabilityIsIndependentAndInstallationBound(t *testing.T) {
	const installation = "caf12459-abcd-4321-8421-aaccff110033"
	const prefix = `{"version":1,"installation":"` + installation + `","profile":"immich-v3.2.2","evidence":"authorized metadata preservation/readback","capabilities":`
	for _, caps := range []string{`["metadata"]`, `["description","stack_gps","metadata"]`} {
		p, err := ParseCapabilityPolicy(prefix + caps + `}`)
		if err != nil || !p.Allows(installation, "immich-v3.2.2", "metadata") {
			t.Fatalf("explicit mirror capability unavailable: %v", err)
		}
		if p.Allows("other", p.Profile, "metadata") || p.Allows(installation, "other", "metadata") {
			t.Fatal("mirror capability escaped installation/profile")
		}
		if caps == `["metadata"]` && (p.Allows(installation, p.Profile, "description") || p.Allows(installation, p.Profile, "stack_gps")) {
			t.Fatal("mirror enabled standard capabilities")
		}
	}
	for _, raw := range []string{"", prefix + `["description","stack_gps"]}`} {
		p, err := ParseCapabilityPolicy(raw)
		if err != nil || p.Allows(installation, "immich-v3.2.2", "metadata") {
			t.Fatal("mirror authority inferred from other capabilities")
		}
	}
}
