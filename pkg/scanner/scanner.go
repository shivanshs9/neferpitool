package scanner

import (
	"strings"
	"sync"
	"time"

	"github.com/moorada/neferpitool/pkg/configuration"
	"github.com/moorada/neferpitool/pkg/constants"
	"github.com/moorada/neferpitool/pkg/domains"
	"github.com/moorada/neferpitool/pkg/log"
)

type ScanOptions struct {
	Apex          string
	WhoisOnlyApex bool
	Concurrency   int
	Phase         string
}

func defaultScanOptions(apex string) ScanOptions {
	conf := configuration.GetConf()
	c := conf.SCAN_CONCURRENCY
	if c < 1 {
		c = 20
	}
	return ScanOptions{
		Apex:          apex,
		WhoisOnlyApex: true,
		Concurrency:   c,
	}
}

// UpdateTypoDomains scans hosts with WHOIS only on the apex when WhoisOnlyApex is set.
func UpdateTypoDomains(tds domains.TypoList, c chan int) map[string]error {
	apex := ""
	if len(tds) > 0 {
		apex = tds[0].LegitDomain
	}
	opts := defaultScanOptions(apex)
	return UpdateHosts(tds, opts, c)
}

func UpdateHosts(tds domains.TypoList, opts ScanOptions, c chan int) map[string]error {
	errs := make(map[string]error)
	var errsMu sync.Mutex

	if len(tds) == 0 {
		return errs
	}

	if opts.Apex == "" {
		opts.Apex = tds[0].LegitDomain
	}
	if opts.Concurrency < 1 {
		opts.Concurrency = 20
	}

	start := time.Now()
	phase := opts.Phase
	if phase == "" {
		phase = "scan"
	}
	summary := tds.SourceSummary()
	log.Info("Zone %s [%s]: scanning %d hosts (%s, WHOIS on apex only: %v)...",
		opts.Apex, phase, len(tds), summary, opts.WhoisOnlyApex)

	sem := make(chan struct{}, opts.Concurrency)
	var wg sync.WaitGroup
	mw := configuration.GetConf().TIMETOSLEEPWHOIS

	for i := range tds {
		wg.Add(1)
		sem <- struct{}{}
		go func(i int) {
			defer wg.Done()
			defer func() { <-sem }()

			name := tds[i].Name
			log.Debug("Checking DNS of %s", name)
			if err := tds[i].UpdateStatus(); err != nil {
				log.Debug("error updatestatus about %s: %s", name, err.Error())
				errsMu.Lock()
				errs[name] = err
				errsMu.Unlock()
			}

			needsWhois := shouldUpdateWhois(tds[i], opts)
			if needsWhois {
				time.Sleep(time.Duration(mw) * time.Millisecond)
				log.Debug("Checking WHOIS of %s", name)
				if err := tds[i].UpdateWhois(); err != nil {
					log.Debug("error updatewhois about %s: %s", name, err.Error())
					errsMu.Lock()
					errs[name] = err
					errsMu.Unlock()
				}
			}

			if c != nil {
				c <- i
			}
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)
	log.Info("Zone %s [%s]: scanned %d hosts in %s", opts.Apex, phase, len(tds), elapsed.String())
	return errs
}

func shouldUpdateWhois(td domains.TypoDomain, opts ScanOptions) bool {
	if td.Status != constants.INACTIVE && td.Status != constants.ACTIVE {
		return false
	}
	if !opts.WhoisOnlyApex {
		return true
	}
	return hostEqualsApex(td.Name, opts.Apex)
}

func hostEqualsApex(host, apex string) bool {
	return strings.EqualFold(strings.TrimSpace(host), strings.TrimSpace(apex))
}
