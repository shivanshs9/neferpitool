package discovery

import (
	"os"
	"testing"

	"github.com/moorada/neferpitool/pkg/log"
)

func TestMain(m *testing.M) {
	_ = log.ActiveConsoleLog(true)
	os.Exit(m.Run())
}
