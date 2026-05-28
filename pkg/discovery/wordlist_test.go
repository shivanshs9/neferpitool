package discovery

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverWordlist_missingFile(t *testing.T) {
	_, err := discoverWordlist("example.com", "/nonexistent/wordlist.txt")
	if err == nil {
		t.Error("expected error for missing wordlist file")
	}
}

func TestDiscoverWordlist_findsResolvableHost(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "wl.txt")
	if err := os.WriteFile(path, []byte("www\n# comment\n\n"), 0644); err != nil {
		t.Fatal(err)
	}

	names, err := discoverWordlist("google.com", path)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) == 0 {
		t.Fatal("expected www.google.com to resolve")
	}
	found := false
	for _, n := range names {
		if n == "www.google.com" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected www.google.com in %v", names)
	}
}

func TestDiscoverSubdomains_includesApex(t *testing.T) {
	// Uses package discovery with default config from environment; at minimum apex is always present.
	hosts := DiscoverSubdomains("example.com")
	if len(hosts) == 0 {
		t.Fatal("expected at least apex host")
	}
	if hosts[0].Name != "example.com" || !hosts[0].IsApex() {
		t.Errorf("expected apex first, got %+v", hosts[0])
	}
}
