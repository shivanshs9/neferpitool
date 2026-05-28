package dns

import "testing"

func TestNormalizeRRString_ignoresTTL(t *testing.T) {
	a := "docs.avantisfi.com.\t26\tIN\tCNAME\tf27f7075c7-hosting.gitbook.io."
	b := "docs.avantisfi.com.\t30\tIN\tCNAME\tf27f7075c7-hosting.gitbook.io."
	if NormalizeRRString(a) != NormalizeRRString(b) {
		t.Fatalf("expected equal canonical, got %q vs %q", NormalizeRRString(a), NormalizeRRString(b))
	}
}

func TestNormalizeRRString_detectsTargetChange(t *testing.T) {
	a := "dev.avantisfi.com.\t30\tIN\tAAAA\t2606:4700:10::6814:2208"
	b := "dev.avantisfi.com.\t30\tIN\tAAAA\t2606:4700:10::ac42:a395"
	if NormalizeRRString(a) == NormalizeRRString(b) {
		t.Fatal("expected different AAAA targets")
	}
}
