package review

import (
	"net/netip"
	"net/url"
	"slices"
	"strings"
	"unicode"

	"immich-places-backend/internal/ai/results"
)

func PresentSources(sources []results.AnswerSource) []results.AnswerSource {
	presented := slices.Clone(sources)
	for i := range presented {
		if !publicReference(presented[i].URL) {
			presented[i].URL = ""
		}
	}
	return presented
}

func publicReference(raw string) bool {
	decoded, err := url.PathUnescape(raw)
	if err != nil || strings.IndexFunc(decoded, unicode.IsControl) >= 0 {
		return false
	}
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil || u.Opaque != "" {
		return false
	}
	host := strings.TrimSuffix(strings.ToLower(u.Hostname()), ".")
	if address, err := netip.ParseAddr(host); err == nil {
		address = address.Unmap()
		return address.IsGlobalUnicast() && !address.IsPrivate() && !address.IsLoopback() && !address.IsLinkLocalUnicast() && !netip.MustParsePrefix("100.64.0.0/10").Contains(address)
	}
	if !strings.Contains(host, ".") || strings.Trim(host, "0123456789.") == "" {
		return false
	}
	for _, suffix := range []string{".localhost", ".local", ".localdomain", ".internal", ".lan", ".home", ".ts.net"} {
		if strings.HasSuffix(host, suffix) {
			return false
		}
	}
	return true
}
