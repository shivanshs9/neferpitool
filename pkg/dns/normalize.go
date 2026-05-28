package dns

import (
	"fmt"
	"sort"
	"strings"

	mdns "github.com/miekg/dns"
)

// NormalizeRRString returns a TTL-independent canonical form for change detection.
func NormalizeRRString(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	rr, err := mdns.NewRR(s)
	if err != nil {
		return collapseWS(s)
	}
	return canonicalRdata(rr)
}

func canonicalRdata(rr mdns.RR) string {
	switch x := rr.(type) {
	case *mdns.A:
		return "A:" + x.A.String()
	case *mdns.AAAA:
		return "AAAA:" + x.AAAA.String()
	case *mdns.CNAME:
		return "CNAME:" + mdns.Fqdn(x.Target)
	case *mdns.MX:
		return fmt.Sprintf("MX:%d:%s", x.Preference, mdns.Fqdn(x.Mx))
	case *mdns.NS:
		return "NS:" + mdns.Fqdn(x.Ns)
	case *mdns.SOA:
		return fmt.Sprintf("SOA:%s:%s:%d:%d:%d:%d:%d",
			mdns.Fqdn(x.Ns), mdns.Fqdn(x.Mbox), x.Serial, x.Refresh, x.Retry, x.Expire, x.Minttl)
	default:
		return collapseWS(rr.String())
	}
}

func collapseWS(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// RecordStringByType returns the RR string for the requested type, searching all Dns
// slots (handles legacy rows where the wrong slot stored the answer).
func RecordStringByType(d Dns, recordType uint16) string {
	for _, s := range []string{d.SOA, d.NS, d.CNAME, d.A, d.AAAA, d.MX} {
		if s == "" {
			continue
		}
		for _, line := range strings.Split(s, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			rr, err := mdns.NewRR(line)
			if err == nil && rr.Header().Rrtype == recordType {
				return line
			}
		}
	}
	return ""
}

// NormalizeRecordSet compares one or more RR strings (newline-separated), ignoring TTL.
func NormalizeRecordSet(s string) string {
	if strings.TrimSpace(s) == "" {
		return ""
	}
	var parts []string
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts = append(parts, NormalizeRRString(line))
	}
	sort.Strings(parts)
	return strings.Join(parts, "\n")
}

// DnsSemanticallyEqual compares record sets ignoring TTL-only differences.
func DnsSemanticallyEqual(a, b Dns) bool {
	return NormalizeRRString(a.SOA) == NormalizeRRString(b.SOA) &&
		NormalizeRRString(a.NS) == NormalizeRRString(b.NS) &&
		NormalizeRRString(a.CNAME) == NormalizeRRString(b.CNAME) &&
		NormalizeRRString(a.A) == NormalizeRRString(b.A) &&
		NormalizeRRString(a.AAAA) == NormalizeRRString(b.AAAA) &&
		NormalizeRRString(a.MX) == NormalizeRRString(b.MX)
}
