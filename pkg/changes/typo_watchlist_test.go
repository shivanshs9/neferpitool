package changes

import (
	"testing"

	"github.com/moorada/neferpitool/pkg/constants"
	"github.com/moorada/neferpitool/pkg/domains"
)

func TestIsTypoActivationTransition(t *testing.T) {
	if !IsTypoActivationTransition(constants.AVAILABLE, constants.ACTIVE) {
		t.Error("Available -> Active should be activation")
	}
	if IsTypoActivationTransition(constants.ACTIVE, constants.ACTIVE) {
		t.Error("Active -> Active is not activation")
	}
}

func TestMakeTypoWatchlistChanges_statusOnly(t *testing.T) {
	old := domains.NewTypoDomain("evil.com", "example.com", "co")
	old.Status = constants.AVAILABLE
	new := old
	new.Status = constants.ACTIVE

	tdcs := MakeTypoWatchlistChanges(old, new, false)
	if len(tdcs) != 1 || tdcs[0].Field != STATUS {
		t.Fatalf("expected status change only, got %+v", tdcs)
	}
}
