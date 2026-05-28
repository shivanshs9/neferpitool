package configuration

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/tkanos/gonfig"

	"github.com/moorada/neferpitool/pkg/constants"
	"github.com/moorada/neferpitool/pkg/log"
)

type configuration struct {
	TYPOSALGHORITM                  []string
	EXPIRATIONTIME                  int
	MAXATTEMPTSWHOIS                int
	MAXATTEMPTSSOA                  int
	TIMETOSLEEPWHOIS                int
	TIMETOSLEEPSOA                  int
	EMAIL                           string
	PASSWORD                        string
	EMAILTONOTIFY                   []string
	SHOWSTATUS                      []string
	PATHRESOLVER                    string
	MINUTESLEEPBACKGROUNDMONITORING int
	CHECKRELIABILITYTIME            int
	REPORTFREQUENCY                 []string

	DISCOVERY_CT              bool
	DISCOVERY_WORDLIST        bool
	SUBDOMAIN_WORDLIST_PATH   string
	TYPO_MODE                 string
	SCAN_CONCURRENCY          int
	DISCOVERY_REFRESH         bool

	MONITOR_ASSET_DNS_CHANGES bool
	MONITOR_TYPO_WATCHLIST    bool
	TYPO_FULL_DNS_CHANGE_CHECK bool
	EVENTS_ENABLED            bool
	LOG_PLAIN                 bool
}

const (
	pathConfig = "./config/config.json"
)

var (
	initVar bool

	standardConf = configuration{
		TYPOSALGHORITM:                  []string{"co", "cs", "hg"},
		EXPIRATIONTIME:                  7,
		MAXATTEMPTSWHOIS:                2,
		MAXATTEMPTSSOA:                  2,
		TIMETOSLEEPWHOIS:                2000,
		TIMETOSLEEPSOA:                  100,
		EMAIL:                           "",
		PASSWORD:                        "",
		EMAILTONOTIFY:                   []string{},
		SHOWSTATUS:                      []string{"Inactive", "Active", "Available", "Unknown", "Alias"},
		PATHRESOLVER:                    "./config/resolv.conf",
		MINUTESLEEPBACKGROUNDMONITORING: 1,
		CHECKRELIABILITYTIME:            2000,
		REPORTFREQUENCY:                 []string{"0 0/4 * * * "},

		DISCOVERY_CT:            true,
		DISCOVERY_WORDLIST:      true,
		SUBDOMAIN_WORDLIST_PATH: "./config/subdomains.txt",
		TYPO_MODE:               constants.TypoModeDeferred,
		SCAN_CONCURRENCY:        20,
		DISCOVERY_REFRESH:       true,

		MONITOR_ASSET_DNS_CHANGES:  true,
		MONITOR_TYPO_WATCHLIST:     true,
		TYPO_FULL_DNS_CHANGE_CHECK: false,
		EVENTS_ENABLED:             true,
		LOG_PLAIN:                  false,
	}
)

func initConf() {

	configuration := configuration{}
	err := gonfig.GetConf(pathConfig, &configuration)
	if err != nil {
		log.Debug("Basic configuration, err: %s", err)
	} else {
		log.Debug("configuration by file")
		standardConf = configuration
	}
	standardConf.applyEnvOverrides()
	standardConf.normalize()
}

func (c *configuration) normalize() {
	c.TYPO_MODE = strings.ToLower(strings.TrimSpace(c.TYPO_MODE))
	switch c.TYPO_MODE {
	case constants.TypoModeOff, constants.TypoModeImmediate, constants.TypoModeDeferred:
	default:
		c.TYPO_MODE = constants.TypoModeDeferred
	}
	if c.SCAN_CONCURRENCY < 1 {
		c.SCAN_CONCURRENCY = 20
	}
}

func init() {
	standardConf.applyEnvOverrides()
	standardConf.normalize()
}

func GetConf() configuration {
	if !initVar {
		initConf()
		initVar = true
	}
	return standardConf
}

func (c configuration) TypoMode() string {
	return c.TYPO_MODE
}

func (c configuration) TypoEnabled() bool {
	return c.TYPO_MODE != constants.TypoModeOff
}

func MakeConfigFile() error {

	config := GetConf()

	jsonFile, err := os.Create(pathConfig)
	if err != nil {
		panic(err)
	} else {
		jsonData, err := json.MarshalIndent(config, "", "  ")

		_, err = jsonFile.WriteAt(jsonData, 0)
		if err != nil {
			panic(err)
		}

	}

	if err != nil {
		log.Error("%s", err.Error())
	} else {
		log.Info("File \"config.json\" created")
	}

	jsonFile.Close()
	return nil
}
