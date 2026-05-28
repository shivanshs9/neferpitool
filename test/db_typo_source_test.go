package test

import (
	"os"
	"testing"

	"github.com/moorada/neferpitool/pkg/constants"
	"github.com/moorada/neferpitool/pkg/db"
	"github.com/moorada/neferpitool/pkg/domains"
)

func TestHasTypoHostsForZone(t *testing.T) {
	_ = os.Remove(dbFile)
	db.InitDB(dbName)
	defer os.Remove(dbFile)

	if db.HasTypoHostsForZone(googleDomain) {
		t.Error("expected no typo hosts before insert")
	}

	td := domains.NewTypoDomain(swappingDomain, googleDomain, algorithm)
	db.AddTypoDomainToDB(td)

	if !db.HasTypoHostsForZone(googleDomain) {
		t.Error("expected typo hosts after insert")
	}

	sub := domains.NewHost("www."+googleDomain, googleDomain, "", constants.SourceSubdomainWordlist)
	db.AddTypoDomainToDB(sub)
	if !db.HasTypoHostsForZone(googleDomain) {
		t.Error("typo count should still be > 0")
	}
}

func TestTypoDomainSource_persisted(t *testing.T) {
	_ = os.Remove(dbFile)
	db.InitDB(dbName)
	defer os.Remove(dbFile)

	host := domains.NewHost("www."+googleDomain, googleDomain, "", constants.SourceSubdomainWordlist)
	db.AddTypoDomainToDB(host)

	got := db.GetTypoDomainFromDB("www." + googleDomain)
	if got.Source != constants.SourceSubdomainWordlist {
		t.Errorf("expected source persisted, got %q", got.Source)
	}
}
