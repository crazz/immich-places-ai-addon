package providers

import (
	"fmt"
	"net/netip"
)

var deniedExact = []netip.Addr{
	netip.MustParseAddr("169.254.169.254"),
	netip.MustParseAddr("100.100.100.200"),
}

// Public provider rules exclude special-purpose networks, including transition
// mechanisms that can embed a non-public IPv4 destination. Registry references:
// https://www.iana.org/assignments/iana-ipv4-special-registry
// https://www.iana.org/assignments/iana-ipv6-special-registry
var publicExcludedPrefixes = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),
	netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("192.0.0.0/24"),
	netip.MustParsePrefix("192.0.2.0/24"),
	netip.MustParsePrefix("192.31.196.0/24"),
	netip.MustParsePrefix("192.52.193.0/24"),
	netip.MustParsePrefix("192.88.99.0/24"),
	netip.MustParsePrefix("192.175.48.0/24"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("198.51.100.0/24"),
	netip.MustParsePrefix("203.0.113.0/24"),
	netip.MustParsePrefix("240.0.0.0/4"),
	netip.MustParsePrefix("2001::/23"),
	netip.MustParsePrefix("2001:db8::/32"),
	netip.MustParsePrefix("2002::/16"),
	netip.MustParsePrefix("2620:4f:8000::/48"),
	netip.MustParsePrefix("3fff::/20"),
}

var publicIPv6Prefix = netip.MustParsePrefix("2000::/3")

func ValidateResolvedAddresses(rule EgressRule, addresses []netip.Addr) ([]netip.Addr, error) {
	if len(addresses) == 0 {
		return nil, fmt.Errorf("destination resolution failed")
	}
	approved := make([]netip.Addr, 0, len(addresses))
	for _, addr := range addresses {
		canonical := canonicalizeAddr(addr)
		if addressDenied(canonical) {
			return nil, fmt.Errorf("destination address is not permitted")
		}
		ok, err := addressAllowedByRule(rule, canonical)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("destination address is not permitted")
		}
		approved = append(approved, canonical)
	}
	return approved, nil
}

func canonicalizeAddr(addr netip.Addr) netip.Addr {
	if addr.Is4In6() {
		return addr.Unmap()
	}
	return addr
}

func addressDenied(addr netip.Addr) bool {
	if !addr.IsValid() || addr.IsUnspecified() || addr.IsMulticast() || addr.IsLinkLocalUnicast() || addr.IsLinkLocalMulticast() {
		return true
	}
	for _, denied := range deniedExact {
		if addr == denied {
			return true
		}
	}
	return false
}

func addressAllowedByRule(rule EgressRule, addr netip.Addr) (bool, error) {
	switch rule.AddressClass {
	case AddressClassLocal:
		for _, cidr := range rule.AllowedCIDRs {
			prefix, err := netip.ParsePrefix(cidr)
			if err != nil {
				return false, fmt.Errorf("invalid allowedCIDRs")
			}
			if prefix.Contains(addr) {
				return true, nil
			}
		}
		return false, nil
	case AddressClassPublic:
		return isPermittedPublicAddress(addr), nil
	default:
		return false, fmt.Errorf("invalid address class")
	}
}

func isPermittedPublicAddress(addr netip.Addr) bool {
	if addressDenied(addr) || addr.IsLoopback() || addr.IsPrivate() || !addr.IsGlobalUnicast() {
		return false
	}
	if addr.Is6() && !publicIPv6Prefix.Contains(addr) {
		return false
	}
	for _, prefix := range publicExcludedPrefixes {
		if prefix.Contains(addr) {
			return false
		}
	}
	return true
}
