package cmd

import (
	"fmt"
	"strings"

	"github.com/moorada/neferpitool/pkg/db"
	"github.com/moorada/neferpitool/pkg/domains"
	"github.com/moorada/neferpitool/pkg/log"
)

func logZonesInDB(mds domains.LegitList) {
	names := make([]string, 0, len(mds))
	for _, d := range mds {
		names = append(names, d.Name)
	}
	log.Info("Background monitoring %d zone(s): %s", len(mds), strings.Join(names, ", "))
}

func backgroundCycleHeader(mds domains.LegitList) {
	names := make([]string, 0, len(mds))
	for _, d := range mds {
		hosts := db.GetTypoDomainListFromDB(d.Name)
		names = append(names, fmt.Sprintf("%s (%d hosts)", d.Name, len(hosts)))
	}
	log.Info("Monitoring cycle — zones: %s", strings.Join(names, "; "))
}
