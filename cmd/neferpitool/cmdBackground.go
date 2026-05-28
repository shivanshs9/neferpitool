package cmd

import (
	"time"

	"github.com/moorada/neferpitool/pkg/changes"
	"github.com/moorada/neferpitool/pkg/configuration"
	"github.com/moorada/neferpitool/pkg/constants"
	"github.com/moorada/neferpitool/pkg/db"
	"github.com/moorada/neferpitool/pkg/log"
	"github.com/moorada/neferpitool/pkg/notification"
	"github.com/moorada/neferpitool/pkg/telemetry"
	"github.com/moorada/neferpitool/pkg/reliableChanges"
	"github.com/robfig/cron/v3"
)

func background() {
	timeToSleepBackground := time.Duration(configuration.GetConf().MINUTESLEEPBACKGROUNDMONITORING) * time.Minute
	for _, c := range configuration.GetConf().REPORTFREQUENCY {
		runCronJob(c)
	}
	mds := db.GetMainDomainListFromDB()
	if len(mds) > 0 {
		logZonesInDB(mds)
		for {
			backgroundWork()
			sleepBetweenCycles(timeToSleepBackground)
		}
	} else {
		log.Error("No domains in the Database")
	}

}

func backgroundWork() {
	start := time.Now()
	ctx, endCycle := telemetry.MonitorCycleSpan(telemetry.BackgroundContext())
	defer endCycle()
	_ = ctx

	backgroundCycleHeader(db.GetMainDomainListFromDB())

	conf := configuration.GetConf()
	if conf.DISCOVERY_REFRESH {
		mds := db.GetMainDomainListFromDB()
		for _, d := range mds {
			newHosts, errs, err := monitorService.RefreshSubdomains(d.Name)
			if err != nil {
				log.Error("Subdomain refresh for %s: %s", d.Name, err.Error())
			}
			if len(errs) > 0 {
				log.Debug("Subdomain refresh scan errors for %s: %v", d.Name, len(errs))
			}
			if len(newHosts) > 0 {
				log.Info("Discovered %d new hosts under %s", len(newHosts), d.Name)
			}
		}
	}

	checkChangesOfAll()

	if conf.TypoMode() == constants.TypoModeDeferred {
		log.Info("Processing deferred typo-domain generation for all registered zones...")
		if errs := monitorService.ProcessDeferredTypos(nil); len(errs) > 0 {
			log.Debug("Deferred typo scan finished with %v errors", len(errs))
		}
	}

	if len(changesToSend) > 0 {
		prepareAndSendEmail()
	}
	elapsed := time.Since(start)
	elapsedMin := int(elapsed / time.Minute)
	log.Debug("Time of full scansion: %v", elapsed)
	if elapsedMin != 0 {
		log.Debug("Typodomains scanned per minute: %v", totaltd/elapsedMin)
	}
}

func runCronJob(expression string) {
	c := cron.New()
	_, err := c.AddFunc(expression, func() {
		log.Info("Cron %s started", expression)
		prepareAndSendReportEmail(expression)
	})

	if err != nil {
		log.Error("Cron error with cron %s: %s", expression, err)
	} else {
		log.Debug("Cron Job started with cron %s", expression)
		c.Start()
	}
}

func prepareAndSendReportEmail(expression string) {
	err, reliableChangesReportToSend := db.GetRelaibleChangesFromDBWithExpression(expression)
	if err != nil {
		log.Error("%s", err.Error())
	}

	conf := configuration.GetConf()

	if conf.EMAIL != "" && conf.PASSWORD != "" && len(conf.EMAILTONOTIFY) != 0 {

		tdsInExpiration := getTypoDomainsInExpiration()
		headersStatus, datasStatus, headersWhois, datasWhois := reliableChangesReportToSend.ToTables()
		hExpiry, dExpiry := tdsInExpiration.ToExpiryTable()

		request := notification.Request{
			From:     conf.EMAIL,
			Password: conf.PASSWORD,
			To:       conf.EMAILTONOTIFY,
			Subject:  "domain monitoring - Report of the day - Expression: " + expression,
		}

		tpl := notification.TemplateData{
			H1:            "Domains Monitoring",
			TextStatus:    "There are status changes",
			TextWhois:     "There are whois changes",
			HeadersStatus: headersStatus,
			HeadersWhois:  headersWhois,
			DatasStatus:   datasStatus,
			DatasWhois:    datasWhois,
			TextExpiry:    "Typodomains in expiration",
			HeadersExpiry: hExpiry,
			DatasExpiry:   dExpiry,
		}

		err := notification.EmailChanges(tpl, request)
		if err != nil {
			log.Error("%s", err.Error())
		} else {
			log.Info("Report sent by email, Expression: %s", expression)
		}
		for _, c := range reliableChangesReportToSend {
			i := reliableChanges.Contains(c.Crons, expression)
			if i >= 0 {
				db.DeleteExprToChDB(c, *c.Crons[i])
			}
		}

	} else {
		log.Info("No email to send")
	}
}

func prepareAndSendEmail() {

	conf := configuration.GetConf()

	if conf.EMAIL != "" && conf.PASSWORD != "" && len(conf.EMAILTONOTIFY) != 0 {

		tdsInExpiration := getTypoDomainsInExpiration()
		headersStatus, datasStatus, headersWhois, datasWhois := changesToSend.ToTables()
		hExpiry, dExpiry := tdsInExpiration.ToExpiryTable()

		request := notification.Request{
			From:     conf.EMAIL,
			Password: conf.PASSWORD,
			To:       conf.EMAILTONOTIFY,
			Subject:  "domain monitoring - Updates",
		}

		tpl := notification.TemplateData{
			H1:            "Domains Monitoring",
			TextStatus:    "There are status changes",
			TextWhois:     "There are whois changes",
			HeadersStatus: headersStatus,
			HeadersWhois:  headersWhois,
			DatasStatus:   datasStatus,
			DatasWhois:    datasWhois,
			TextExpiry:    "Typodomains in expiration",
			HeadersExpiry: hExpiry,
			DatasExpiry:   dExpiry,
		}

		err := notification.EmailChanges(tpl, request)
		if err != nil {
			log.Error("%s", err.Error())
		} else {
			log.Info("Changes sent by email")
		}
	} else {
		log.Info("No email to send")
	}
	changesToSend = changes.ChangeList{}
}
