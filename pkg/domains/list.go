package domains

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/moorada/neferpitool/pkg/constants"
	"github.com/moorada/neferpitool/pkg/format"
	"github.com/moorada/neferpitool/pkg/log"
)

type TypoList []TypoDomain

type LegitList []LegitDomain

func (a TypoList) Len() int {
	return len(a)
}
func (a TypoList) Less(i, j int) bool {

	date1, err := a[i].GetExpiryDate()
	if err != nil {
		log.Error("%s", err.Error())
	}
	date2, err := a[j].GetExpiryDate()
	if err != nil {
		log.Error("%s", err.Error())
	}
	return date1.Before(date2)
}
func (a TypoList) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (tdl TypoList) ToExpiryTable() ([]string, [][]string) {

	sort.Sort(tdl)

	var headers []string
	headers = append(headers, "Typo Domain")
	headers = append(headers, "Expiry Date")

	var data [][]string
	for _, d := range tdl {
		//w, _ := d.GetWhois()
		w := d.GetWhois()

		timeString := ""

		expiration, err := format.StringToTime(w.Parsed.Registrar.ExpirationDate)
		if err != nil {
			log.Error("Error to parse data of %s, Data: %s, Error: %s", d.Name, w.Parsed.Registrar.ExpirationDate, err.Error())
		}
		timeString = format.TimeToStringConsole(expiration)

		data = append(data, []string{d.Name, timeString})
	}
	return headers, data
}

func (tdl TypoList) GetUnfilledCopy() TypoList {
	var tdlnew TypoList
	for _, td := range tdl {
		tdlnew = append(tdlnew, NewHost(td.Name, td.LegitDomain, td.Algorithm, td.Source))
	}
	return tdlnew
}

func (tdl TypoList) FilterInExpiration(d int) TypoList {

	var tdtInExpiration []TypoDomain

	now := time.Now()

	for _, td := range tdl {

		//if td.Whois != "" {
		/*w, err := td.GetWhois()
		if err != nil {
			log.Error("%s", err.Error())
		}
		*/
		w := td.Whois
		if w.Parsed.Registrar.ExpirationDate != "" {
			expiryDate, err := format.StringToTime(w.Parsed.Registrar.ExpirationDate)
			if err != nil {
				log.Error("%s %s %s", td.Name, w.Parsed.Registrar.ExpirationDate, err)
			} else {
				diff := expiryDate.Sub(now)

				days := int(diff.Hours() / 24)
				if days < d && days > 0 {
					tdtInExpiration = append(tdtInExpiration, td)
				}
			}
		}

		//}

	}

	return tdtInExpiration
}

func (tds TypoList) ToMap() map[string]TypoDomain {
	tdsMap := make(map[string]TypoDomain)
	for _, d := range tds {
		tdsMap[d.Name] = d
	}
	return tdsMap
}

// Apex returns the parent zone when all hosts share the same LegitDomain.
// FilterAssets returns apex and discovered subdomain hosts (excludes typos).
func (tds TypoList) FilterAssets() TypoList {
	var out TypoList
	for _, td := range tds {
		if td.Source != constants.SourceTypo {
			out = append(out, td)
		}
	}
	return out
}

// FilterTypos returns only typo-squat generated hosts.
func (tds TypoList) FilterTypos() TypoList {
	var out TypoList
	for _, td := range tds {
		if td.Source == constants.SourceTypo {
			out = append(out, td)
		}
	}
	return out
}

func (tds TypoList) Apex() string {
	if len(tds) == 0 {
		return ""
	}
	apex := tds[0].LegitDomain
	for _, td := range tds[1:] {
		if td.LegitDomain != apex {
			return ""
		}
	}
	return apex
}

// SourceSummary returns a compact breakdown like "apex=1, subdomain_wordlist=11, typo=87".
func (tds TypoList) SourceSummary() string {
	if len(tds) == 0 {
		return "none"
	}
	counts := make(map[string]int)
	for _, td := range tds {
		src := td.Source
		if src == "" {
			src = "unknown"
		}
		counts[src]++
	}
	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%d", k, counts[k]))
	}
	return strings.Join(parts, ", ")
}

func (ds LegitList) ToMap() map[string]LegitDomain {
	dsMap := make(map[string]LegitDomain)
	for _, d := range ds {
		dsMap[d.Name] = d
	}
	return dsMap
}
