package providers

import (
	"net/netip"
	"testing"
)

func TestLocalExceptionCannotAuthorizeMetadata(t *testing.T) {
	rule := EgressRule{
		BaseURL:      "http://local-provider:80/v1",
		AddressClass: AddressClassLocal,
		AllowedCIDRs: []string{"10.0.0.0/8", "127.0.0.0/8", "169.254.0.0/16", "224.0.0.0/4", "0.0.0.0/32", "100.100.100.0/24"},
	}
	for _, tc := range []struct {
		name, ip string
	}{
		{"unspecified", "0.0.0.0"},
		{"multicast", "224.0.0.1"},
		{"link-local", "169.254.1.1"},
		{"metadata", "169.254.169.254"},
		{"aliyun metadata", "100.100.100.200"},
		{"ipv4-mapped metadata", "::ffff:169.254.169.254"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ValidateResolvedAddresses(rule, []netip.Addr{netip.MustParseAddr(tc.ip)})
			if err == nil {
				t.Fatal("special or metadata destination must be denied")
			}
		})
	}
}
