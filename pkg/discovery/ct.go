package discovery

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type crtShEntry struct {
	NameValue string `json:"name_value"`
}

func discoverCT(apex string) ([]string, error) {
	q := url.QueryEscape("%." + apex)
	reqURL := fmt.Sprintf("https://crt.sh/?q=%s&output=json", q)
	return discoverCTFromURL(reqURL, apex)
}

func discoverCTFromURL(reqURL, apex string) ([]string, error) {
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(reqURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("crt.sh status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return parseCTNames(body, apex)
}

func parseCTNames(body []byte, apex string) ([]string, error) {
	var entries []crtShEntry
	if err := json.Unmarshal(body, &entries); err != nil {
		return nil, err
	}

	seen := map[string]bool{}
	var names []string
	for _, e := range entries {
		for _, part := range strings.Split(e.NameValue, "\n") {
			part = strings.TrimSpace(strings.ToLower(part))
			if part == "" || strings.HasPrefix(part, "*.") {
				continue
			}
			if part == apex || strings.HasSuffix(part, "."+apex) {
				if !seen[part] {
					seen[part] = true
					names = append(names, part)
				}
			}
		}
	}
	return names, nil
}
