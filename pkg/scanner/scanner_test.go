package scanner

import (
	"testing"

	"github.com/moorada/neferpitool/pkg/constants"
	"github.com/moorada/neferpitool/pkg/domains"
)

func TestShouldUpdateWhois_apexOnly(t *testing.T) {
	opts := ScanOptions{Apex: "example.com", WhoisOnlyApex: true}

	apex := domains.NewHost("example.com", "example.com", "", constants.SourceApex)
	apex.Status = constants.ACTIVE
	if !shouldUpdateWhois(apex, opts) {
		t.Error("expected WHOIS for apex when active")
	}

	sub := domains.NewHost("api.example.com", "example.com", "", constants.SourceSubdomainWordlist)
	sub.Status = constants.ACTIVE
	if shouldUpdateWhois(sub, opts) {
		t.Error("expected no WHOIS for subdomain when WhoisOnlyApex is set")
	}

	typo := domains.NewTypoDomain("examp1e.com", "example.com", "co")
	typo.Status = constants.INACTIVE
	if shouldUpdateWhois(typo, opts) {
		t.Error("expected no WHOIS for typo domain when WhoisOnlyApex is set")
	}
}

func TestShouldUpdateWhois_availableSkips(t *testing.T) {
	opts := ScanOptions{Apex: "example.com", WhoisOnlyApex: true}
	apex := domains.NewHost("example.com", "example.com", "", constants.SourceApex)
	apex.Status = constants.AVAILABLE
	if shouldUpdateWhois(apex, opts) {
		t.Error("available domains should not trigger WHOIS")
	}
}

func TestHostEqualsApex(t *testing.T) {
	if !hostEqualsApex("Example.COM", "example.com") {
		t.Error("expected case-insensitive apex match")
	}
	if hostEqualsApex("api.example.com", "example.com") {
		t.Error("subdomain should not match apex")
	}
}
