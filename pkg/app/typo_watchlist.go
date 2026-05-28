package app

import (
	"github.com/moorada/neferpitool/pkg/changes"
	"github.com/moorada/neferpitool/pkg/configuration"
	"github.com/moorada/neferpitool/pkg/domains"
	"github.com/moorada/neferpitool/pkg/events"
	"github.com/moorada/neferpitool/pkg/log"
)

// CheckTypoWatchlist scans typo hosts for registration (status) changes; optional full DNS diff.
func (s *Service) CheckTypoWatchlist(typos domains.TypoList, progress ProgressFn) (domains.TypoList, changes.ChangeList, map[string]error) {
	if len(typos) == 0 {
		return nil, nil, nil
	}

	apex := typos.Apex()
	conf := configuration.GetConf()
	fullDNS := conf.TYPO_FULL_DNS_CHANGE_CHECK

	log.Info("Zone %s: typo watchlist check (%d hosts, full_dns=%v)...", apex, len(typos), fullDNS)

	tdsNew := typos.GetUnfilledCopy()
	scanErrs := s.scanHosts(tdsNew, apex, "typo-watchlist", progress)

	var changed domains.TypoList
	var chs changes.ChangeList

	oldMap := typos.ToMap()
	for _, tdNew := range tdsNew {
		tdOld, ok := oldMap[tdNew.Name]
		if !ok {
			continue
		}
		if !tdNew.IsReliableAboutPrev(tdOld) {
			continue
		}

		tdcs := changes.MakeTypoWatchlistChanges(tdOld, tdNew, fullDNS)
		if len(tdcs) == 0 {
			continue
		}

		for _, c := range tdcs {
			if c.Field == changes.STATUS && changes.IsTypoActivationTransition(tdOld.Status, tdNew.Status) {
				events.EmitTypoActivated(apex, tdNew.Name, tdOld.StatusToString(), tdNew.StatusToString())
			}
		}

		changed = append(changed, tdNew)
		chs = append(chs, tdcs...)
	}

	return changed, chs, scanErrs
}

// EmitTypoDiscoveredEvents emits one structured event per newly generated typo host.
func EmitTypoDiscoveredEvents(zone string, typos domains.TypoList) {
	for _, td := range typos {
		events.EmitTypoDiscovered(zone, td.Name, td.Algorithm, td.StatusToString())
	}
}
