package whois

import "testing"

func TestNormalizeNameServers_orderIndependent(t *testing.T) {
	a := "NS1.CLOUDFLARE.COM, ns2.cloudflare.com."
	b := "ns2.cloudflare.com,ns1.cloudflare.com"
	if NormalizeNameServers(a) != NormalizeNameServers(b) {
		t.Fatalf("expected equal, got %q vs %q", NormalizeNameServers(a), NormalizeNameServers(b))
	}
}

func TestNormalizeNameServers_detectsChange(t *testing.T) {
	a := "kyrie.ns.cloudflare.com"
	b := "rita.ns.cloudflare.com"
	if NormalizeNameServers(a) == NormalizeNameServers(b) {
		t.Fatal("expected different normalized NS sets")
	}
}
