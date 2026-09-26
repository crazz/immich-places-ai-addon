package writeback

import "testing"

func TestStackCapabilityIsIndependentAndInstallationBound(t *testing.T) {
	const installation = "caf12459-abcd-4321-8421-aaccff110033"
	const prefix = `{"version":1,"installation":"` + installation + `","profile":"immich-v3.2.2","evidence":"synthetic exact member test","capabilities":`
	for _, tc := range []struct {
		caps               string
		description, stack bool
	}{
		{`["stack_gps"]`, false, true},
		{`["description"]`, true, false},
		{`["description","stack_gps"]`, true, true},
	} {
		p, err := ParseCapabilityPolicy(prefix + tc.caps + `}`)
		if err != nil || p.Allows(installation, "immich-v3.2.2", "description") != tc.description || p.Allows(installation, "immich-v3.2.2", "stack_gps") != tc.stack {
			t.Fatalf("capabilities not independent: %s %v", tc.caps, err)
		}
		if p.Allows("other", p.Profile, "stack_gps") || p.Allows(installation, "other", "stack_gps") || p.Allows(installation, p.Profile, "unknown") {
			t.Fatal("capability escaped its binding")
		}
	}
	p, err := ParseCapabilityPolicy("")
	if err != nil || p.Allows(installation, "immich-v3.2.2", "stack_gps") {
		t.Fatal("stack scope enabled by default")
	}
}
