package providers

import (
	"fmt"
	"net/netip"
)

var deniedExact = []netip.Addr{
	netip.MustParseAddr("169.254.169.254"),
	netip.MustParseAddr("100.100.100.200"),
}

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
	if addr.Is6() && netip.MustParsePrefix("fc00::/7").Contains(addr) {
		return false
	}
	return true
}
