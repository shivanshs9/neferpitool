package changes

import (
	"testing"

	"github.com/moorada/neferpitool/pkg/dns"
)

func TestMakeDNSChanges_ignoresTTLOnly(t *testing.T) {
	oldR := dns.Dns{CNAME: "docs.example.com.\t26\tIN\tCNAME\ttarget.example."}
	newR := dns.Dns{CNAME: "docs.example.com.\t30\tIN\tCNAME\ttarget.example."}
	cl := makeDNSChanges("docs.example.com", oldR, newR)
	if len(cl) != 0 {
		t.Fatalf("expected no changes, got %+v", cl)
	}
}

func TestMakeDNSChanges_misplacedLegacySlot(t *testing.T) {
	// CNAME stored in SOA slot (legacy); should report DNS CNAME, not SOA.
	oldR := dns.Dns{SOA: "docs.example.com.\t26\tIN\tCNAME\ttarget.example."}
	newR := dns.Dns{CNAME: "docs.example.com.\t30\tIN\tCNAME\ttarget.example."}
	cl := makeDNSChanges("docs.example.com", oldR, newR)
	if len(cl) != 0 {
		t.Fatalf("expected no changes after canonical compare, got %+v", cl)
	}
}

func TestMakeDNSChanges_ignoresSOASerialOnly(t *testing.T) {
	oldR := dns.Dns{SOA: "avantisfi.com.\t30\tIN\tSOA\tkyrie.ns.cloudflare.com. dns.cloudflare.com. 2406305687 10000 2400 604800 1800"}
	newR := dns.Dns{SOA: "avantisfi.com.\t30\tIN\tSOA\tkyrie.ns.cloudflare.com. dns.cloudflare.com. 2406382306 10000 2400 604800 1800"}
	cl := makeDNSChanges("avantisfi.com", oldR, newR)
	if len(cl) != 0 {
		t.Fatalf("expected no SOA change when only serial differs, got %+v", cl)
	}
}

func TestMakeDNSChanges_NSOrderStable(t *testing.T) {
	ns1 := "example.com.\t30\tIN\tNS\tns1.cloudflare.com."
	ns2 := "example.com.\t30\tIN\tNS\tns2.cloudflare.com."
	oldR := dns.Dns{NS: ns1 + "\n" + ns2}
	newR := dns.Dns{NS: ns2 + "\n" + ns1}
	cl := makeDNSChanges("example.com", oldR, newR)
	if len(cl) != 0 {
		t.Fatalf("expected no change when NS set is same, got %+v", cl)
	}
}
