package configuration

import (
	"encoding/json"
	"os"

	"github.com/moorada/neferpitool/pkg/terminal"
)

type logPlainConfig struct {
	LOG_PLAIN bool `json:"LOG_PLAIN"`
}

// InitialPlainLogs decides console format before the full config loader runs (no logging yet).
func InitialPlainLogs() bool {
	if v := os.Getenv("LOG_PLAIN"); v != "" {
		return ParseBoolEnv(v)
	}
	// Pipes, redirects, tee, Docker, and K8s logs must not receive ANSI escape codes.
	if !terminal.StdoutIsTerminal() {
		return true
	}
	if plain, ok := plainLogsFromFile(); ok {
		return plain
	}
	return false
}

func plainLogsFromFile() (bool, bool) {
	f, err := os.Open(pathConfig)
	if err != nil {
		return false, false
	}
	defer f.Close()
	var c logPlainConfig
	if err := json.NewDecoder(f).Decode(&c); err != nil {
		return false, false
	}
	return c.LOG_PLAIN, true
}
