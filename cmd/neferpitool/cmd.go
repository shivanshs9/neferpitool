package cmd

//cmdroot
import (
	"context"
	"flag"
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"

	"github.com/cheggaaa/pb/v3"
	"github.com/moorada/neferpitool/pkg/app"

	"github.com/moorada/neferpitool/pkg/configuration"
	"github.com/moorada/neferpitool/pkg/constants"

	"github.com/moorada/neferpitool/pkg/console"
	"github.com/moorada/neferpitool/pkg/db"
	"github.com/moorada/neferpitool/pkg/domains"
	"github.com/moorada/neferpitool/pkg/log"
	"github.com/moorada/neferpitool/pkg/telemetry"
	"github.com/moorada/neferpitool/pkg/terminal"
)

var otelShutdown func(context.Context) error = func(context.Context) error { return nil }

var totaltd int
var monitorService = app.NewService()

const PathConfigFolder = "./config"

func CmdRoot() {

	// init flags
	logs := flag.Bool("logs", false, "Avtive logs on file")
	singleTd := flag.String("td", "", "Manage one typodomain")
	bg := flag.Bool("bg", false, "Active monitoring in background")
	pd := flag.Bool("pd", false, "Check if domains are present")
	makeConfig := flag.Bool("mc", false, "Make config file")
	importTds := flag.String("it", "", "Import Typos from file - main domain")
	pathImportTds := flag.String("p", "", "Import Typos from file - path of the file")

	flag.Parse()

	// Log must be ready before configuration.GetConf() (initConf uses log.Debug).
	if err := log.ActiveConsoleLog(configuration.InitialPlainLogs()); err != nil {
		panic(err)
	}
	defer log.Close()

	ctx := context.Background()
	if shutdown, err := telemetry.Init(ctx); err != nil {
		log.Warning("OTEL init failed (traces disabled): %s", err.Error())
	} else {
		otelShutdown = shutdown
		if telemetry.Enabled() {
			log.Info("OTEL trace export enabled (service: check OTEL_SERVICE_NAME / OTEL_RESOURCE_ATTRIBUTES)")
		}
	}
	defer func() { _ = otelShutdown(context.Background()) }()

	err := os.MkdirAll(PathConfigFolder, os.ModePerm)
	if err != nil {
		log.Error("%s", err.Error())
	}

	db.InitDB("config/database")
	defer db.CloseDB()

	if *logs {
		if err := log.ActiveDebugLog(); err != nil {
			panic(err)
		}
	}

	if *makeConfig {
		configuration.MakeConfigFile()
		return
	}

	logDegubInfo()

	if *bg {
		log.Info("Starting background monitoring (args: %v)", flag.Args())
		ensureZonesRegistered(flag.Args())
		background()
		return
	}

	if *singleTd != "" {
		SingleTdMode(*singleTd)
		return
	}

	if *importTds != "" {
		if *pathImportTds != "" {
			importTypos(*importTds, *pathImportTds)
			return
		} else {
			log.Error("%s", "Specify the path of the import file: -p /path/example/ ")
		}

	}

	args := flag.Args()

	if len(args) < 1 {
		MonitorCmd()
		return
	}

	if *pd {
		presenceDomains(args)
		return
	}

	if len(args) == 1 {
		domain := args[0]
		SingleDomainMode(domain)
	} else {
		multipleDomainsMode(args)
	}

	os.Exit(0)
}

func presenceDomains(domains []string) {
	presence := monitorService.DomainPresence(domains)
	for _, s := range domains {
		if presence[s] {
			log.Info("%s present", s)
		} else {
			log.Info("%s NOT present", s)
		}
	}
}

func multipleDomainsMode(domains []string) {
	for _, d := range domains {
		tds, errs, err := addDomainAndHisTypos(d)
		if err != nil {
			log.Error("%s, error: %s", d, err.Error())
		}
		if len(errs) != 0 {
			console.PrintTableErrs(errs)
		}
		console.PrintTableTypoDomains(tds)
	}
}

func logDegubInfo() {
	// debug info...
	totaltd = 0
	totald := db.GetMainDomainListFromDB()

	domains := ""

	for _, d := range totald {
		td1d := db.GetTypoDomainListFromDB(d.Name)
		domains += d.Name + ", "
		totaltd = totaltd + len(td1d)

	}
	log.Debug("%v domains in db: %s", len(totald), domains)
	log.Debug("Number of typodomains in db: %v", totaltd)
	//
}

func UpdateTypoDomainsWithProgressBar(tds domains.TypoList) map[string]error {

	bar := pb.Full.Start(len(tds))
	errs := monitorService.ScanTypoDomains(tds, func(done, total int) {
		bar.SetCurrent(int64(done))
	})
	bar.Finish()
	if len(tds) > 0 {
		log.Debug("Stats Scansion, Typodomains: %v, Errors: %v, Percentage of errors: %v", len(tds), len(errs), len(errs)*100/len(tds))
	} else {
		log.Debug("no typodomains")
	}
	return errs
}

func addDomainAndHisTypos(domain string) (domains.TypoList, map[string]error, error) {

	var tds domains.TypoList
	var errs map[string]error
	var err error
	if terminal.UseInteractiveUI(configuration.GetConf().LOG_PLAIN) {
		s := spinner.New(spinner.CharSets[26], 200*time.Millisecond)
		s.Prefix = "Discovering hosts and scanning "
		s.Start()
		tds, errs, err = monitorService.AddDomainAndTypos(domain, nil)
		s.Stop()
	} else {
		log.Info("Discovering hosts and scanning zone %s...", domain)
		tds, errs, err = monitorService.AddDomainAndTypos(domain, nil)
	}

	if err != nil {
		return tds, errs, err
	}

	log.Info("Zone %s added (%d monitored hosts)", domain, len(tds))
	if configuration.GetConf().TypoMode() == constants.TypoModeDeferred {
		log.Info("Typo-domain generation queued for background (TYPO_MODE=deferred)")
	}
	return tds, errs, nil

}

// ensureZonesRegistered adds apex zones from CLI args when they are not already in the DB.
func ensureZonesRegistered(args []string) {
	for _, raw := range args {
		domain := strings.TrimSpace(strings.ToLower(raw))
		if domain == "" {
			continue
		}
		if monitorService.DomainPresence([]string{domain})[domain] {
			log.Info("Zone %s already registered", domain)
			continue
		}
		log.Info("Registering zone %s...", domain)
		_, errs, err := monitorService.AddDomainAndTypos(domain, nil)
		if err != nil {
			log.Error("Zone %s: %s", domain, err.Error())
		}
		if len(errs) > 0 {
			console.PrintTableErrs(errs)
		}
	}
}

func importTypos(domain string, path string) {

	s := spinner.New(spinner.CharSets[26], 200*time.Millisecond)
	s.Prefix = "Importing typodomains "
	s.Start()

	_, errs, err := monitorService.ImportTypos(domain, path, nil)
	if err != nil {
		log.Error("Importing typos from file for %s, error: %s", domain, err.Error())
		s.Stop()
		return
	}
	s.Stop()

	if len(errs) != 0 {
		console.PrintTableErrs(errs)
	}

	log.Info("Typodomains imported")

}
