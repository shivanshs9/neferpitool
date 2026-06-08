package dns

import "testing"

func TestNormalizeRRString_ignoresTTL(t *testing.T) {
	a := "docs.avantisfi.com.\t26\tIN\tCNAME\tf27f7075c7-hosting.gitbook.io."
	b := "docs.avantisfi.com.\t30\tIN\tCNAME\tf27f7075c7-hosting.gitbook.io."
	if NormalizeRRString(a) != NormalizeRRString(b) {
		t.Fatalf("expected equal canonical, got %q vs %q", NormalizeRRString(a), NormalizeRRString(b))
	}
}

func TestNormalizeRRString_ignoresSOASerial(t *testing.T) {
	a := "avantisfi.com.\t30\tIN\tSOA\tkyrie.ns.cloudflare.com. dns.cloudflare.com. 2406305687 10000 2400 604800 1800"
	b := "avantisfi.com.\t30\tIN\tSOA\tkyrie.ns.cloudflare.com. dns.cloudflare.com. 2406382306 10000 2400 604800 1800"
	if NormalizeRRString(a) != NormalizeRRString(b) {
		t.Fatalf("expected equal SOA ignoring serial, got %q vs %q", NormalizeRRString(a), NormalizeRRString(b))
	}
}

func TestNormalizeRRString_detectsSOAMnameChange(t *testing.T) {
	a := "avantisfi.com.\t30\tIN\tSOA\tkyrie.ns.cloudflare.com. dns.cloudflare.com. 1 10000 2400 604800 1800"
	b := "avantisfi.com.\t30\tIN\tSOA\trita.ns.cloudflare.com. dns.cloudflare.com. 1 10000 2400 604800 1800"
	if NormalizeRRString(a) == NormalizeRRString(b) {
		t.Fatal("expected different SOA when primary NS changes")
	}
}

func TestNormalizeRRString_detectsTargetChange(t *testing.T) {
	a := "dev.avantisfi.com.\t30\tIN\tAAAA\t2606:4700:10::6814:2208"
	b := "dev.avantisfi.com.\t30\tIN\tAAAA\t2606:4700:10::ac42:a395"
	if NormalizeRRString(a) == NormalizeRRString(b) {
		t.Fatal("expected different AAAA targets")
	}
}
