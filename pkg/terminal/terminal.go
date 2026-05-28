package terminal

import (
	"os"

	"golang.org/x/term"
)

// StdoutIsTerminal reports whether stdout is an interactive terminal.
func StdoutIsTerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// UseInteractiveUI is false when plain logs are forced or stdout is not a TTY.
func UseInteractiveUI(forcePlain bool) bool {
	if forcePlain {
		return false
	}
	return StdoutIsTerminal()
}
