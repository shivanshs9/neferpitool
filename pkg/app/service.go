package app

import (
	"bufio"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/moorada/neferpitool/pkg/log"

	"github.com/moorada/neferpitool/pkg/changes"
	"github.com/moorada/neferpitool/pkg/configuration"
	"github.com/moorada/neferpitool/pkg/constants"
	"github.com/moorada/neferpitool/pkg/db"
	"github.com/moorada/neferpitool/pkg/discovery"
	"github.com/moorada/neferpitool/pkg/domains"
	"github.com/moorada/neferpitool/pkg/events"
	"github.com/moorada/neferpitool/pkg/generator"
	"github.com/moorada/neferpitool/pkg/reliableChanges"
	"github.com/moorada/neferpitool/pkg/scanner"
)

type ProgressFn func(done, total int)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) DomainPresence(domainNames []string) map[string]bool {
	domainSet := map[string]bool{}
	for _, d := range db.GetMainDomainListFromDB() {
		domainSet[d.Name] = true
	}

	presence := map[string]bool{}
	for _, name := range domainNames {
		presence[name] = domainSet[name]
	}
	return presence
}

func (s *Service) scanHosts(tds domains.TypoList, apex, phase string, progress ProgressFn) map[string]error {
	c := make(chan int, len(tds))
	errsCh := make(chan map[string]error, 1)

	if apex == "" {
		apex = tds.Apex()
	}

	opts := scanner.ScanOptions{
		Apex:          apex,
		WhoisOnlyApex: true,
		Concurrency:   configuration.GetConf().SCAN_CONCURRENCY,
		Phase:         phase,
	}

	go func() {
		errsCh <- scanner.UpdateHosts(tds, opts, c)
	}()

	done := 0
	total := len(tds)
	for done < total {
		<-c
		done++
		if progress != nil {
			progress(done, total)
		}
	}

	return <-errsCh
}

// ScanTypoDomains scans hosts; WHOIS runs only on the apex domain.
func (s *Service) ScanTypoDomains(tds domains.TypoList, progress ProgressFn) map[string]error {
	apex := ""
	if len(tds) > 0 {
		apex = tds[0].LegitDomain
	}
	return s.scanHosts(tds, apex, "scan", progress)
}

// AddDomainAndTypos registers an apex zone, discovers subdomains, scans hosts, and optionally generates typos.
func (s *Service) AddDomainAndTypos(domain string, progress ProgressFn) (domains.TypoList, map[string]error, error) {
	return s.AddZone(domain, progress)
}

func (s *Service) AddZone(domain string, progress ProgressFn) (domains.TypoList, map[string]error, error) {
	domain = strings.TrimSpace(strings.ToLower(domain))
	var allErrs map[string]error

	md := domains.NewLegitDomain(domain)
	db.AddLegitDomainToDB(md)

	hosts := discovery.DiscoverSubdomains(domain)
	errs := s.scanHosts(hosts, domain, "initial-discovery", progress)
	allErrs = mergeErrs(allErrs, errs)
	db.AddTypoListToDB(hosts)

	if err := md.Update(); err != nil {
		return hosts, allErrs, err
	}
	db.AddLegitDomainToDB(md)

	var typos domains.TypoList
	switch configuration.GetConf().TypoMode() {
	case constants.TypoModeImmediate:
		typos = generator.GetUnfilledTypoDomains(domain)
		errs = s.scanHosts(typos, domain, "typo-immediate", progress)
		allErrs = mergeErrs(allErrs, errs)
		db.AddTypoListToDB(typos)
		hosts = append(hosts, typos...)
	case constants.TypoModeDeferred, constants.TypoModeOff:
	}

	return hosts, allErrs, nil
}

func (s *Service) ImportTypos(domain, path string, progress ProgressFn) (domains.TypoList, map[string]error, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	sc := bufio.NewScanner(file)
	var tds domains.TypoList
	for sc.Scan() {
		name := strings.TrimSpace(sc.Text())
		if name == "" {
			continue
		}
		tds = append(tds, domains.NewHost(name, domain, "imported", constants.SourceImported))
	}

	if err := sc.Err(); err != nil {
		return nil, nil, err
	}

	if len(tds) == 0 {
		return nil, nil, errors.New("empty file")
	}

	errs := s.scanHosts(tds, domain, "import", progress)
	db.AddTypoListToDB(tds)
	return tds, errs, nil
}

// RefreshSubdomains discovers new CT/wordlist hosts and returns only names not yet in the DB.
func (s *Service) RefreshSubdomains(apex string) (domains.TypoList, map[string]error, error) {
	existing := db.GetTypoDomainListFromDB(apex).ToMap()
	discovered := discovery.DiscoverSubdomains(apex)

	var newHosts domains.TypoList
	for _, h := range discovered {
		if h.Source == constants.SourceTypo {
			continue
		}
		if _, ok := existing[h.Name]; !ok {
			newHosts = append(newHosts, h)
		}
	}
	if len(newHosts) == 0 {
		return nil, nil, nil
	}

	errs := s.scanHosts(newHosts, apex, "subdomain-refresh", nil)
	db.AddTypoListToDB(newHosts)
	for _, h := range newHosts {
		events.EmitHostDiscovered(apex, h.Name, h.Source, h.StatusToString())
	}
	return newHosts, errs, nil
}

// ProcessDeferredTypos generates and scans typo-squat domains for zones that do not have them yet.
func (s *Service) ProcessDeferredTypos(progress ProgressFn) map[string]error {
	if configuration.GetConf().TypoMode() != constants.TypoModeDeferred {
		return nil
	}

	var allErrs map[string]error
	for _, d := range db.GetMainDomainListFromDB() {
		if db.HasTypoHostsForZone(d.Name) {
			continue
		}
		typos := generator.GetUnfilledTypoDomains(d.Name)
		if len(typos) == 0 {
			continue
		}
		log.Info("Zone %s: generating %d deferred typo hosts...", d.Name, len(typos))
		errs := s.scanHosts(typos, d.Name, "deferred-typos", progress)
		db.AddTypoListToDB(typos)
		EmitTypoDiscoveredEvents(d.Name, typos)
		allErrs = mergeErrs(allErrs, errs)
	}
	return allErrs
}

func (s *Service) GetTypoDomainsInExpiration() domains.TypoList {
	var total domains.TypoList
	expirationDays := configuration.GetConf().EXPIRATIONTIME

	for _, d := range db.GetMainDomainListFromDB() {
		tds := db.GetTypoDomainListWithStatusFromDB(d.Name, []int{constants.INACTIVE, constants.ACTIVE, constants.ALIAS})
		total = append(total, tds.FilterInExpiration(expirationDays)...)
	}

	return total
}

// IterateCheckAssetChanges runs full DNS/WHOIS (apex) change detection for zone assets (not typos).
func (s *Service) IterateCheckAssetChanges(assets domains.TypoList, progress ProgressFn) (tdsReliable []domains.TypoDomain, changesReliable []changes.Change, scanErrs map[string]error) {
	if len(assets) == 0 {
		return nil, nil, nil
	}
	apex := assets.Apex()
	log.Info("Zone %s: asset DNS change check (%d hosts)...", apex, len(assets))
	return s.iterateCheckGetChangesInternal(assets, apex, progress)
}

func (s *Service) IterateCheckGetChanges(tds domains.TypoList, progress ProgressFn) (tdsReliable []domains.TypoDomain, changesReliable []changes.Change, scanErrs map[string]error) {
	apex := tds.Apex()
	return s.iterateCheckGetChangesInternal(tds, apex, progress)
}

func (s *Service) iterateCheckGetChangesInternal(tds domains.TypoList, apex string, progress ProgressFn) (tdsReliable []domains.TypoDomain, changesReliable []changes.Change, scanErrs map[string]error) {
	tdsNew := tds.GetUnfilledCopy()
	scanErrs = s.scanHosts(tdsNew, apex, "change-check", progress)

	tdsOldCh, tdsNewCh, chs := changes.MakeChangeList(tds, tdsNew)
	if len(tdsOldCh) > 0 {
		time.Sleep(time.Duration(configuration.GetConf().CHECKRELIABILITYTIME) * time.Millisecond)
		scanErrs = mergeErrs(scanErrs, s.scanHosts(tdsNewCh, apex, "change-verify", progress))

		tdsOldChNext, tdsNewChNext, chsNext := changes.MakeChangeList(tdsOldCh, tdsNewCh)
		tdsReliable, changesReliable = chsNext.FilterReliableWithPrev(chs, tdsNewCh, tdsNewChNext)
		_ = tdsOldChNext
	} else {
		tdsReliable = nil
		changesReliable = nil
	}

	for _, c := range changesReliable {
		switch {
		case changes.IsWhoisField(c.Field):
			events.EmitWHOISChanged(apex, c.TypoDomain, c.Field, c.Before, c.After)
		case len(c.Field) >= 4 && c.Field[:4] == "DNS ":
			events.Emit(events.DNSChanged, map[string]string{
				"zone":                     apex,
				events.AttrMonitoredDomain: c.TypoDomain,
				"field":                    c.Field,
				"before":                   c.Before,
				"after":                    c.After,
				"change.type":              "dns",
			})
		}
	}

	return tdsReliable, changesReliable, scanErrs
}

func (s *Service) SaveReliableChanges(changesList changes.ChangeList) {
	var crons []*reliableChanges.CronExpression
	for _, expression := range configuration.GetConf().REPORTFREQUENCY {
		cronExpr := db.GetCronExpressionFromDB(expression)
		if cronExpr.ID != 0 {
			crons = append(crons, &cronExpr)
		} else {
			crons = append(crons, &reliableChanges.CronExpression{Exrpression: expression})
		}
	}

	var rChanges []reliableChanges.ReliableChange
	for _, c := range changesList {
		rChanges = append(rChanges, reliableChanges.ReliableChange{
			TypoDomain: c.TypoDomain,
			Field:      c.Field,
			Before:     c.Before,
			After:      c.After,
			Crons:      crons,
		})
	}
	db.AddReliableChangeListToDB(rChanges)
}

func mergeErrs(first, second map[string]error) map[string]error {
	if first == nil && second == nil {
		return map[string]error{}
	}
	if first == nil {
		first = map[string]error{}
	}
	for k, v := range second {
		first[k] = v
	}
	return first
}
