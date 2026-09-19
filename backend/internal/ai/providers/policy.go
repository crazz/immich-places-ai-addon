package providers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/netip"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

const (
	AddressClassPublic = "public"
	AddressClassLocal  = "local"
)

// EgressPolicy is the installation allowlist for provider destinations.
type EgressPolicy struct {
	Rules []EgressRule
}

// EgressRule approves one exact canonical base URL and its address class.
type EgressRule struct {
	BaseURL      string
	AddressClass string
	AllowedCIDRs []string
}

func ParseEgressPolicy(raw string) (EgressPolicy, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return EgressPolicy{}, nil
	}
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.DisallowUnknownFields()
	var items []egressRuleJSON
	if err := decoder.Decode(&items); err != nil {
		return EgressPolicy{}, fmt.Errorf("invalid policy")
	}
	if err := decoder.Decode(new(any)); !errors.Is(err, io.EOF) {
		return EgressPolicy{}, fmt.Errorf("invalid policy")
	}
	rules := make([]EgressRule, 0, len(items))
	seen := map[string]EgressRule{}
	for _, item := range items {
		rule, err := normalizeEgressRule(item)
		if err != nil {
			return EgressPolicy{}, err
		}
		if previous, ok := seen[rule.BaseURL]; ok {
			if !sameEgressRule(previous, rule) {
				return EgressPolicy{}, fmt.Errorf("conflicting policy rules")
			}
			continue
		}
		seen[rule.BaseURL] = rule
		rules = append(rules, rule)
	}
	return EgressPolicy{Rules: rules}, nil
}

type egressRuleJSON struct {
	BaseURL      string   `json:"baseURL"`
	AddressClass string   `json:"addressClass"`
	AllowedCIDRs []string `json:"allowedCIDRs"`
}

func normalizeEgressRule(item egressRuleJSON) (EgressRule, error) {
	canonical, err := CanonicalBaseURL(item.BaseURL)
	if err != nil {
		return EgressRule{}, err
	}
	class := strings.TrimSpace(item.AddressClass)
	switch class {
	case AddressClassPublic:
		if canonical.Scheme != "https" {
			return EgressRule{}, fmt.Errorf("public destinations require HTTPS")
		}
		if len(item.AllowedCIDRs) != 0 {
			return EgressRule{}, fmt.Errorf("public destinations must not set allowedCIDRs")
		}
		return EgressRule{BaseURL: canonical.String(), AddressClass: class}, nil
	case AddressClassLocal:
		if len(item.AllowedCIDRs) == 0 {
			return EgressRule{}, fmt.Errorf("local destinations require allowedCIDRs")
		}
		cidrs, err := parseLocalCIDRs(item.AllowedCIDRs)
		if err != nil {
			return EgressRule{}, err
		}
		return EgressRule{BaseURL: canonical.String(), AddressClass: class, AllowedCIDRs: cidrs}, nil
	default:
		return EgressRule{}, fmt.Errorf("addressClass must be public or local")
	}
}

type CanonicalURL struct {
	Scheme string
	Host   string
	Port   string
	Path   string
}

func (c CanonicalURL) String() string {
	return c.Scheme + "://" + net.JoinHostPort(c.Host, c.Port) + c.Path
}

func CanonicalBaseURL(raw string) (CanonicalURL, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || len(raw) > 2048 {
		return CanonicalURL{}, fmt.Errorf("invalid base URL")
	}
	if strings.ContainsAny(raw, `*\`) || strings.Contains(raw, "%") {
		return CanonicalURL{}, fmt.Errorf("invalid base URL")
	}
	u, err := url.Parse(raw)
	if err != nil || u.Hostname() == "" || (u.Scheme != "https" && u.Scheme != "http") || u.User != nil || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" {
		return CanonicalURL{}, fmt.Errorf("invalid base URL")
	}
	host := strings.ToLower(u.Hostname())
	if strings.ContainsAny(host, "*[]") || strings.Contains(host, "%") {
		return CanonicalURL{}, fmt.Errorf("invalid base URL")
	}
	if ip, err := netip.ParseAddr(host); err == nil {
		if ip.Zone() != "" || ip.Is4In6() || (ip.Is6() && !strings.HasPrefix(u.Host, "[")) {
			return CanonicalURL{}, fmt.Errorf("invalid base URL")
		}
		host = ip.String()
	} else if !validASCIIHostname(host) {
		return CanonicalURL{}, fmt.Errorf("invalid base URL")
	}
	port := u.Port()
	if strings.HasSuffix(u.Host, ":") {
		return CanonicalURL{}, fmt.Errorf("invalid base URL")
	}
	if port == "" {
		if u.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}
	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		return CanonicalURL{}, fmt.Errorf("invalid base URL")
	}
	port = strconv.Itoa(portNumber)
	path := u.EscapedPath()
	if path == "" {
		path = "/"
	}
	if !strings.HasPrefix(path, "/") || strings.Contains(path, "//") {
		return CanonicalURL{}, fmt.Errorf("invalid base URL")
	}
	for _, segment := range strings.Split(path, "/") {
		if segment == "." || segment == ".." {
			return CanonicalURL{}, fmt.Errorf("invalid base URL")
		}
	}
	path = strings.TrimRight(path, "/")
	if path == "" {
		path = "/"
	}
	return CanonicalURL{Scheme: u.Scheme, Host: host, Port: port, Path: path}, nil
}

func validASCIIHostname(host string) bool {
	if len(host) > 253 {
		return false
	}
	labels := strings.Split(host, ".")
	for _, label := range labels {
		if label == "" || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, ch := range label {
			if !(ch >= 'a' && ch <= 'z' || ch >= '0' && ch <= '9' || ch == '-') {
				return false
			}
		}
	}
	// Numeric hosts must have passed netip.ParseAddr above; reject legacy IPv4 forms.
	last := labels[len(labels)-1]
	_, numeric := strconv.ParseUint(last, 10, 64)
	return numeric != nil && !strings.HasPrefix(last, "0x")
}

func parseLocalCIDRs(values []string) ([]string, error) {
	out := make([]string, 0, len(values))
	for _, value := range values {
		prefix, err := netip.ParsePrefix(strings.TrimSpace(value))
		if err != nil {
			return nil, fmt.Errorf("invalid allowedCIDRs")
		}
		prefix = prefix.Masked()
		if !localPrefixAllowed(prefix) {
			return nil, fmt.Errorf("allowedCIDRs must stay within private or loopback ranges")
		}
		out = append(out, prefix.String())
	}
	return out, nil
}

func localPrefixAllowed(prefix netip.Prefix) bool {
	addr := prefix.Addr()
	if !addr.Is4() && !addr.Is6() {
		return false
	}
	for _, allowed := range localContainmentPrefixes() {
		if allowed.Addr().Is4() != addr.Is4() {
			continue
		}
		if allowed.Contains(prefix.Addr()) && allowed.Bits() <= prefix.Bits() && containsPrefix(allowed, prefix) {
			return true
		}
	}
	return false
}

func containsPrefix(outer, inner netip.Prefix) bool {
	if outer.Addr().BitLen() != inner.Addr().BitLen() {
		return false
	}
	if outer.Bits() > inner.Bits() {
		return false
	}
	return outer.Contains(inner.Addr()) && outer.Contains(lastAddr(inner))
}

func lastAddr(prefix netip.Prefix) netip.Addr {
	addr := prefix.Addr()
	bits := make([]byte, addr.BitLen()/8)
	copy(bits, addr.AsSlice())
	hostBits := addr.BitLen() - prefix.Bits()
	for i := len(bits) - 1; hostBits > 0; i-- {
		n := hostBits
		if n > 8 {
			n = 8
		}
		bits[i] |= byte((1 << n) - 1)
		hostBits -= n
	}
	out, ok := netip.AddrFromSlice(bits)
	if !ok {
		return addr
	}
	return out
}

func localContainmentPrefixes() []netip.Prefix {
	return []netip.Prefix{
		netip.MustParsePrefix("10.0.0.0/8"),
		netip.MustParsePrefix("172.16.0.0/12"),
		netip.MustParsePrefix("192.168.0.0/16"),
		netip.MustParsePrefix("127.0.0.0/8"),
		netip.MustParsePrefix("fc00::/7"),
		netip.MustParsePrefix("::1/128"),
	}
}

func sameEgressRule(a, b EgressRule) bool {
	if a.BaseURL != b.BaseURL || a.AddressClass != b.AddressClass || len(a.AllowedCIDRs) != len(b.AllowedCIDRs) {
		return false
	}
	for i := range a.AllowedCIDRs {
		if a.AllowedCIDRs[i] != b.AllowedCIDRs[i] {
			return false
		}
	}
	return true
}

func MatchEgressDestination(policy EgressPolicy, baseURL string) (EgressRule, error) {
	canonical, err := CanonicalBaseURL(baseURL)
	if err != nil {
		return EgressRule{}, fmt.Errorf("destination not approved")
	}
	target := canonical.String()
	for _, rule := range policy.Rules {
		if rule.BaseURL == target {
			return rule, nil
		}
	}
	return EgressRule{}, fmt.Errorf("destination not approved")
}

func FingerprintPolicy(policy EgressPolicy) string {
	type stableRule struct {
		BaseURL      string   `json:"baseURL"`
		AddressClass string   `json:"addressClass"`
		AllowedCIDRs []string `json:"allowedCIDRs,omitempty"`
	}
	rules := make([]stableRule, 0, len(policy.Rules))
	for _, rule := range policy.Rules {
		cidrs := append([]string(nil), rule.AllowedCIDRs...)
		sort.Strings(cidrs)
		rules = append(rules, stableRule{
			BaseURL:      rule.BaseURL,
			AddressClass: rule.AddressClass,
			AllowedCIDRs: cidrs,
		})
	}
	sort.Slice(rules, func(i, j int) bool { return rules[i].BaseURL < rules[j].BaseURL })
	raw, err := json.Marshal(rules)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}
