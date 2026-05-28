package cmd

import (
	"github.com/cheggaaa/pb/v3"
	"github.com/manifoldco/promptui"
	"github.com/moorada/neferpitool/pkg/changes"
	"github.com/moorada/neferpitool/pkg/configuration"
	"github.com/moorada/neferpitool/pkg/console"
	"github.com/moorada/neferpitool/pkg/db"
	"github.com/moorada/neferpitool/pkg/domains"
	"github.com/moorada/neferpitool/pkg/log"
	"github.com/moorada/neferpitool/pkg/terminal"
)

func checkForChanges() {

	mds := db.GetMainDomainListFromDB()
	var mdsnames []string

	mdsnames = append(mdsnames, "All")

	for _, d := range mds {
		mdsnames = append(mdsnames, d.Name)
	}

	prompt := promptui.Select{
		Label: "Check for changes",
		Items: mdsnames,
	}

	_, result, err := prompt.Run()

	if err != nil {
		log.Error("Prompt failed %v\n", err)
		return
	}
	if result == "All" {
		checkChangesOfAll()
	} else {
		log.Info("Checking changes about %s", result)
		checkZoneChanges(db.GetTypoDomainListFromDB(result))
	}
}

func showTypoDomainsInExpiration() {
	tdsEx := getTypoDomainsInExpiration()
	if len(tdsEx) != 0 {
		console.PrintTableTypoDomains(tdsEx)
	} else {
		log.Error("No expiry typodomains in the Database")
	}

}

func checkChangesOfAll() {
	conf := configuration.GetConf()
	mds := db.GetMainDomainListFromDB()
	for i, d := range mds {
		all := db.GetTypoDomainListFromDB(d.Name)
		assets := all.FilterAssets()
		typos := all.FilterTypos()

		log.Info("Zone %s: monitoring pass (%d assets, %d typos), zone %d of %d",
			d.Name, len(assets), len(typos), i+1, len(mds))

		if conf.MONITOR_ASSET_DNS_CHANGES {
			checkAssetChanges(assets)
		}
		if conf.MONITOR_TYPO_WATCHLIST && conf.TypoEnabled() {
			checkTypoWatchlistChanges(typos)
		}
	}
}

func checkZoneChanges(all domains.TypoList) bool {
	conf := configuration.GetConf()
	changed := false
	if conf.MONITOR_ASSET_DNS_CHANGES {
		changed = checkAssetChanges(all.FilterAssets()) || changed
	}
	if conf.MONITOR_TYPO_WATCHLIST && conf.TypoEnabled() {
		changed = checkTypoWatchlistChanges(all.FilterTypos()) || changed
	}
	return changed
}

func checkAssetChanges(assets domains.TypoList) bool {
	if len(assets) == 0 {
		return false
	}
	tdsChanged, chs, errs := iterateCheckAssetChanges(assets)
	if len(errs) > 0 {
		console.PrintTableErrs(errs)
	}
	return applyChanges(assets.Apex(), tdsChanged, chs, "asset")
}

func checkTypoWatchlistChanges(typos domains.TypoList) bool {
	if len(typos) == 0 {
		return false
	}
	tdsChanged, chs, errs := monitorService.CheckTypoWatchlist(typos, scanProgressFn())
	if len(errs) > 0 {
		console.PrintTableErrs(errs)
	}
	return applyChanges(typos.Apex(), tdsChanged, chs, "typo watchlist")
}

func applyChanges(apex string, tdsChanged domains.TypoList, chs changes.ChangeList, kind string) bool {
	if len(chs) == 0 {
		log.Info("Zone %s: no %s changes", apex, kind)
		return false
	}
	monitorService.SaveReliableChanges(chs)
	db.AddTypoListToDB(tdsChanged)
	console.PrintChanges(chs)
	changesToSend = append(changesToSend, chs...)
	log.Info("Zone %s: %d %s change(s) detected", apex, len(chs), kind)
	return true
}

func iterateCheckAssetChanges(tds domains.TypoList) (domains.TypoList, changes.ChangeList, map[string]error) {
	tdsReliable, changesReliable, errs := monitorService.IterateCheckAssetChanges(tds, scanProgressFn())
	return tdsReliable, changesReliable, errs
}

func scanProgressFn() func(done, total int) {
	if !terminal.UseInteractiveUI(configuration.GetConf().LOG_PLAIN) {
		return nil
	}
	var bar *pb.ProgressBar
	return func(done, total int) {
		if total == 0 {
			return
		}
		if done == 1 {
			bar = pb.Full.Start(total)
		}
		if bar != nil {
			bar.SetCurrent(int64(done))
			if done == total {
				bar.Finish()
				bar = nil
			}
		}
	}
}

func getTypoDomainsInExpiration() domains.TypoList {
	return monitorService.GetTypoDomainsInExpiration()
}
