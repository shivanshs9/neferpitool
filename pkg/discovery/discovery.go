package discovery

import (
	"strings"

	"github.com/moorada/neferpitool/pkg/configuration"
	"github.com/moorada/neferpitool/pkg/constants"
	"github.com/moorada/neferpitool/pkg/domains"
	"github.com/moorada/neferpitool/pkg/log"
)

// DiscoverSubdomains returns hosts to monitor under apex (apex itself + discovered subdomains).
func DiscoverSubdomains(apex string) domains.TypoList {
	apex = strings.TrimSpace(strings.ToLower(apex))
	conf := configuration.GetConf()

	seen := map[string]bool{}
	var hosts domains.TypoList

	add := func(name, source string) {
		name = strings.TrimSpace(strings.ToLower(name))
		if name == "" || seen[name] {
			return
		}
		if !strings.HasSuffix(name, apex) && name != apex {
			return
		}
		seen[name] = true
		hosts = append(hosts, domains.NewHost(name, apex, "", source))
	}

	add(apex, constants.SourceApex)

	if conf.DISCOVERY_CT {
		names, err := discoverCT(apex)
		if err != nil {
			log.Warning("CT discovery for %s: %s", apex, err.Error())
		} else {
			for _, n := range names {
				add(n, constants.SourceSubdomainCT)
			}
			log.Info("CT discovery for %s: %d names", apex, len(names))
		}
	}

	if conf.DISCOVERY_WORDLIST {
		names, err := discoverWordlist(apex, conf.SUBDOMAIN_WORDLIST_PATH)
		if err != nil {
			log.Warning("Wordlist discovery for %s: %s", apex, err.Error())
		} else {
			for _, n := range names {
				add(n, constants.SourceSubdomainWordlist)
			}
			log.Info("Wordlist discovery for %s: %d resolvable names", apex, len(names))
		}
	}

	return hosts
}
