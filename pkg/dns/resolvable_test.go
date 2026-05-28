package dns

import (
	"testing"

	"github.com/moorada/neferpitool/pkg/constants"
)

func TestIsResolvable(t *testing.T) {
	if !IsResolvable(constants.ACTIVE) {
		t.Error("ACTIVE should be resolvable")
	}
	if !IsResolvable(constants.INACTIVE) {
		t.Error("INACTIVE should be resolvable")
	}
	if !IsResolvable(constants.ALIAS) {
		t.Error("ALIAS should be resolvable")
	}
	if IsResolvable(constants.AVAILABLE) {
		t.Error("AVAILABLE should not be resolvable")
	}
	if IsResolvable(constants.UNKNOWN) {
		t.Error("UNKNOWN should not be resolvable")
	}
}
