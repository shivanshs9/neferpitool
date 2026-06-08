package dns

import (
	"fmt"
	"net"
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
		if ip := net.ParseIP(s); ip != nil {
			if ip.To4() != nil {
				return "A:" + ip.String()
			}
			return "AAAA:" + ip.String()
		}
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
		// Serial is omitted: authoritative zones bump it on any edit (e.g. Cloudflare).
		return fmt.Sprintf("SOA:%s:%s:%d:%d:%d:%d",
			mdns.Fqdn(x.Ns), mdns.Fqdn(x.Mbox), x.Refresh, x.Retry, x.Expire, x.Minttl)
	default:
		return collapseWS(rr.String())
	}
}

func collapseWS(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// RecordsJoinedByType returns all RR strings of the given type (newline-separated),
// searching every slot so legacy misplaced answers still compare correctly.
func RecordsJoinedByType(d Dns, recordType uint16) string {
	var matches []string
	seen := map[string]struct{}{}
	add := func(line string) {
		if line == "" {
			return
		}
		if _, ok := seen[line]; ok {
			return
		}
		seen[line] = struct{}{}
		matches = append(matches, line)
	}
	for _, slot := range []struct {
		val      string
		allowRaw bool
	}{
		{d.SOA, false},
		{d.NS, false},
		{d.CNAME, false},
		{d.A, true},
		{d.AAAA, true},
		{d.MX, false},
	} {
		if slot.val == "" {
			continue
		}
		for _, line := range strings.Split(slot.val, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			rr, err := mdns.NewRR(line)
			if err == nil && rr.Header().Rrtype == recordType {
				add(line)
				continue
			}
			if err != nil && slot.allowRaw {
				ip := net.ParseIP(line)
				if ip == nil {
					continue
				}
				if recordType == mdns.TypeA && ip.To4() != nil {
					add(line)
				}
				if recordType == mdns.TypeAAAA && ip.To4() == nil {
					add(line)
				}
			}
		}
	}
	sort.Strings(matches)
	return strings.Join(matches, "\n")
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
