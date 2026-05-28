package whois

import (
	"sort"
	"strings"
)

// NormalizeNameServers returns a stable, order-independent form for WHOIS nameserver lists.
func NormalizeNameServers(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	var parts []string
	for _, p := range strings.Split(raw, ",") {
		p = strings.TrimSpace(strings.ToLower(p))
		p = strings.TrimSuffix(p, ".")
		if p != "" {
			parts = append(parts, p)
		}
	}
	sort.Strings(parts)
	return strings.Join(parts, ",")
}
