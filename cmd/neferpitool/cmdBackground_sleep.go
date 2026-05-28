package cmd

import (
	"time"

	"github.com/briandowns/spinner"
	"github.com/moorada/neferpitool/pkg/configuration"
	"github.com/moorada/neferpitool/pkg/log"
	"github.com/moorada/neferpitool/pkg/terminal"
)

func sleepBetweenCycles(d time.Duration) {
	if terminal.UseInteractiveUI(configuration.GetConf().LOG_PLAIN) {
		s := spinner.New(spinner.CharSets[26], 200*time.Millisecond)
		s.Prefix = "Sleeping "
		s.Start()
		time.Sleep(d)
		s.Stop()
		return
	}
	log.Info("Next monitoring cycle in %s", d.String())
	time.Sleep(d)
}
