package discovery

import (
	"bufio"
	"os"
	"strings"
	"sync"

	"github.com/moorada/neferpitool/pkg/configuration"
	"github.com/moorada/neferpitool/pkg/dns"
)

func discoverWordlist(apex, path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var prefixes []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		p := strings.TrimSpace(sc.Text())
		if p == "" || strings.HasPrefix(p, "#") {
			continue
		}
		prefixes = append(prefixes, p)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}

	concurrency := configuration.GetConf().SCAN_CONCURRENCY
	if concurrency < 1 {
		concurrency = 20
	}

	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex
	seen := map[string]bool{}
	var found []string

	for _, prefix := range prefixes {
		wg.Add(1)
		sem <- struct{}{}
		go func(prefix string) {
			defer wg.Done()
			defer func() { <-sem }()

			host := strings.ToLower(prefix) + "." + apex
			status, _, _, _ := dns.CheckDNS(host)
			if !dns.IsResolvable(status) {
				return
			}
			mu.Lock()
			if !seen[host] {
				seen[host] = true
				found = append(found, host)
			}
			mu.Unlock()
		}(prefix)
	}
	wg.Wait()
	return found, nil
}
