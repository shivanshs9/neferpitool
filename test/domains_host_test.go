package test

import (
	"testing"

	"github.com/moorada/neferpitool/pkg/constants"
	"github.com/moorada/neferpitool/pkg/domains"
)

func TestNewHost_and_IsApex(t *testing.T) {
	apex := domains.NewHost(googleDomain, googleDomain, "", constants.SourceApex)
	if !apex.IsApex() {
		t.Error("apex host should report IsApex() true")
	}
	if apex.Source != constants.SourceApex {
		t.Errorf("expected source %q, got %q", constants.SourceApex, apex.Source)
	}

	sub := domains.NewHost("www."+googleDomain, googleDomain, "", constants.SourceSubdomainWordlist)
	if sub.IsApex() {
		t.Error("subdomain should not be apex")
	}
	if sub.LegitDomain != googleDomain {
		t.Errorf("expected zone %q, got %q", googleDomain, sub.LegitDomain)
	}
}

func TestNewTypoDomain_setsTypoSource(t *testing.T) {
	td := domains.NewTypoDomain(swappingDomain, googleDomain, algorithm)
	if td.Source != constants.SourceTypo {
		t.Errorf("expected source %q, got %q", constants.SourceTypo, td.Source)
	}
}

func TestTypoList_FilterAssetsAndTypos(t *testing.T) {
	list := domains.TypoList{
		domains.NewHost(googleDomain, googleDomain, "", constants.SourceApex),
		domains.NewHost("www."+googleDomain, googleDomain, "", constants.SourceSubdomainWordlist),
		domains.NewTypoDomain(swappingDomain, googleDomain, algorithm),
	}
	if len(list.FilterAssets()) != 2 {
		t.Errorf("expected 2 assets, got %d", len(list.FilterAssets()))
	}
	if len(list.FilterTypos()) != 1 {
		t.Errorf("expected 1 typo, got %d", len(list.FilterTypos()))
	}
}

func TestTypoList_SourceSummary(t *testing.T) {
	list := domains.TypoList{
		domains.NewHost(googleDomain, googleDomain, "", constants.SourceApex),
		domains.NewHost("www."+googleDomain, googleDomain, "", constants.SourceSubdomainWordlist),
		domains.NewTypoDomain(swappingDomain, googleDomain, algorithm),
	}
	summary := list.SourceSummary()
	if summary == "" || summary == "none" {
		t.Errorf("expected non-empty summary, got %q", summary)
	}
}

func TestGetUnfilledCopy_preservesSource(t *testing.T) {
	td := domains.NewHost("api."+googleDomain, googleDomain, "", constants.SourceSubdomainCT)
	list := domains.TypoList{td}
	copy := list.GetUnfilledCopy()
	if len(copy) != 1 || copy[0].Source != constants.SourceSubdomainCT {
		t.Errorf("source not preserved: %+v", copy[0])
	}
}
