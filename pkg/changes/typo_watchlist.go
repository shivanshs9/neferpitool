package changes

import (
	"github.com/moorada/neferpitool/pkg/constants"
	"github.com/moorada/neferpitool/pkg/domains"
)

// IsResolvableStatus reports DNS-derived statuses that indicate the name is in use.
func IsResolvableStatus(status int) bool {
	return status == constants.INACTIVE || status == constants.ACTIVE || status == constants.ALIAS
}

// MakeTypoWatchlistChanges detects typo watchlist signals: registration (status) and optional full DNS diffs.
func MakeTypoWatchlistChanges(tdOld, tdNew domains.TypoDomain, fullDNS bool) ChangeList {
	var tdcs ChangeList

	status := tdOld.Status == tdNew.Status
	if !status {
		tdcs = append(tdcs, Change{
			tdOld.Name,
			STATUS,
			tdOld.StatusToString(),
			tdNew.StatusToString(),
		})
	}

	if fullDNS {
		tdcs = append(tdcs, makeDNSChanges(tdOld.Name, tdOld.Dns, tdNew.Dns)...)
	}

	return tdcs
}

// IsTypoActivationTransition is true when a typo becomes resolvable (registered / in use).
func IsTypoActivationTransition(oldStatus, newStatus int) bool {
	wasAvailable := oldStatus == constants.AVAILABLE || oldStatus == constants.UNKNOWN
	return wasAvailable && IsResolvableStatus(newStatus)
}
