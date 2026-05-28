package configuration

import (
	"os"
	"strconv"
	"strings"
)

// applyEnvOverrides applies container-friendly settings from environment variables.
// Env vars take precedence over config.json when set.
func (c *configuration) applyEnvOverrides() {
	if v := os.Getenv("LOG_PLAIN"); v != "" {
		c.LOG_PLAIN = parseBoolEnv(v)
	}
	if v := os.Getenv("EVENTS_ENABLED"); v != "" {
		c.EVENTS_ENABLED = parseBoolEnv(v)
	}
	if v := os.Getenv("TYPO_MODE"); v != "" {
		c.TYPO_MODE = strings.ToLower(strings.TrimSpace(v))
	}
	if v := os.Getenv("DISCOVERY_CT"); v != "" {
		c.DISCOVERY_CT = parseBoolEnv(v)
	}
	if v := os.Getenv("DISCOVERY_WORDLIST"); v != "" {
		c.DISCOVERY_WORDLIST = parseBoolEnv(v)
	}
	if v := os.Getenv("PATHRESOLVER"); v != "" {
		c.PATHRESOLVER = v
	}
	if v := os.Getenv("MINUTESLEEPBACKGROUNDMONITORING"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			c.MINUTESLEEPBACKGROUNDMONITORING = n
		}
	}
}

func parseBoolEnv(v string) bool {
	v = strings.ToLower(strings.TrimSpace(v))
	return v == "1" || v == "true" || v == "yes" || v == "on"
}

// ParseBoolEnv parses common truthy strings from environment variables.
func ParseBoolEnv(v string) bool {
	return parseBoolEnv(v)
}
