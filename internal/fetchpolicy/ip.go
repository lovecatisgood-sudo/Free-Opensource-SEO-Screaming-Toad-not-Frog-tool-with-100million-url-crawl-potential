package fetchpolicy

import "net/netip"

// IPDecision explains whether an address may be contacted by the crawler.
type IPDecision struct {
	Allowed bool
	Reason  string
}

// ClassifyIP rejects every non-global destination relevant to SSRF. Go's
// IsGlobalUnicast alone is insufficient because it includes private ranges.
func ClassifyIP(addr netip.Addr) IPDecision {
	if !addr.IsValid() {
		return IPDecision{Reason: "invalid_ip"}
	}
	addr = addr.Unmap()
	switch {
	case addr.IsLoopback():
		return IPDecision{Reason: "loopback"}
	case addr.IsPrivate():
		return IPDecision{Reason: "private"}
	case addr.IsLinkLocalUnicast(), addr.IsLinkLocalMulticast():
		return IPDecision{Reason: "link_local"}
	case addr.IsMulticast():
		return IPDecision{Reason: "multicast"}
	case addr.IsUnspecified():
		return IPDecision{Reason: "unspecified"}
	case !addr.IsGlobalUnicast():
		return IPDecision{Reason: "non_global"}
	}

	for _, blocked := range prohibitedPrefixes {
		if blocked.Contains(addr) {
			return IPDecision{Reason: "reserved"}
		}
	}
	return IPDecision{Allowed: true, Reason: "public"}
}

var prohibitedPrefixes = mustPrefixes(
	"0.0.0.0/8",
	"100.64.0.0/10",
	"192.0.0.0/24",
	"192.0.2.0/24",
	"198.18.0.0/15",
	"198.51.100.0/24",
	"203.0.113.0/24",
	"240.0.0.0/4",
	"2001:db8::/32",
	"2001:10::/28",
	"fc00::/7",
	"fe80::/10",
)

func mustPrefixes(values ...string) []netip.Prefix {
	result := make([]netip.Prefix, 0, len(values))
	for _, value := range values {
		result = append(result, netip.MustParsePrefix(value))
	}
	return result
}
