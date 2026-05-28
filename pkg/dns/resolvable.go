package dns

import "github.com/moorada/neferpitool/pkg/constants"

// IsResolvable reports whether DNS status indicates the name has records worth monitoring.
func IsResolvable(status int) bool {
	return status == constants.INACTIVE || status == constants.ACTIVE || status == constants.ALIAS
}
