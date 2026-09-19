package providers

import (
	"net/netip"
	"testing"
)

func TestPublicDestinationRejectsSpecialPurposeRanges(t *testing.T) {
	rule := EgressRule{AddressClass: AddressClassPublic}
	for _, address := range []string{
		"0.1.2.3", "100.64.0.1", "192.0.0.8", "192.0.2.1", "192.88.99.1",
		"198.18.0.1", "198.51.100.1", "203.0.113.1", "240.0.0.1",
		"::ffff:100.64.0.1", "::7f00:1", "64:ff9b::7f00:1", "64:ff9b:1::1",
		"100::1", "100:0:0:1::1", "2001::1", "2001:2::1", "2001:db8::1",
		"2002:7f00:1::1", "3fff::1", "5f00::1",
	} {
		t.Run(address, func(t *testing.T) {
			if _, err := ValidateResolvedAddresses(rule, []netip.Addr{netip.MustParseAddr(address)}); err == nil {
				t.Fatal("special-purpose address must not be admitted as a public provider")
			}
		})
	}
	for _, address := range []string{"8.8.8.8", "1.1.1.1", "2606:4700:4700::1111", "2001:4860:4860::8888"} {
		if _, err := ValidateResolvedAddresses(rule, []netip.Addr{netip.MustParseAddr(address)}); err != nil {
			t.Fatalf("public unicast address %s rejected: %v", address, err)
		}
	}
}
