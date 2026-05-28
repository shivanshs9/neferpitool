package configuration

import (
	"testing"

	"github.com/moorada/neferpitool/pkg/terminal"
)

func TestInitialPlainLogs_nonTTYForcesPlain(t *testing.T) {
	if terminal.StdoutIsTerminal() {
		t.Skip("stdout is a TTY in this test runner")
	}
	if !InitialPlainLogs() {
		t.Error("expected plain logs when stdout is not a terminal")
	}
}
