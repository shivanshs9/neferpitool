package discovery

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDiscoverCT_parsesNames(t *testing.T) {
	entries := []crtShEntry{
		{NameValue: "api.example.com"},
		{NameValue: "www.example.com\n*.example.com"},
		{NameValue: "other.org"},
	}
	body, _ := json.Marshal(entries)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(body)
	}))
	defer srv.Close()

	names, err := discoverCTFromURL(srv.URL, "example.com")
	if err != nil {
		t.Fatal(err)
	}

	if len(names) != 2 {
		t.Fatalf("expected 2 names, got %v", names)
	}
	seen := map[string]bool{}
	for _, n := range names {
		seen[n] = true
	}
	if !seen["api.example.com"] || !seen["www.example.com"] {
		t.Errorf("unexpected names: %v", names)
	}
}
