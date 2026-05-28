package changes

import (
	"testing"

	"github.com/moorada/neferpitool/pkg/constants"
	"github.com/moorada/neferpitool/pkg/dns"
	"github.com/moorada/neferpitool/pkg/domains"
)

func TestMakeChange_DNSRecordDiff(t *testing.T) {
	old := domains.NewHost("api.example.com", "example.com", "", constants.SourceSubdomainWordlist)
	new := old
	new.Dns = dns.Dns{A: "1.2.3.4"}

	tdcs := MakeChange(old, new)
	if len(tdcs) != 1 {
		t.Fatalf("expected 1 DNS change, got %d: %+v", len(tdcs), tdcs)
	}
	if tdcs[0].Field != DNS_A || tdcs[0].Before != "" || tdcs[0].After != "1.2.3.4" {
		t.Errorf("unexpected change: %+v", tdcs[0])
	}
}

func TestMakeChange_WhoisOnlyOnApex(t *testing.T) {
	apexOld := domains.NewHost("example.com", "example.com", "", constants.SourceApex)
	apexNew := apexOld
	apexNew.Whois.Parsed.Registrant.RegistrantName = "changed-registrant"

	tdcs := MakeChange(apexOld, apexNew)
	foundWhois := false
	for _, c := range tdcs {
		if c.Field == NAME_REGISTRANT {
			foundWhois = true
		}
	}
	if !foundWhois {
		t.Error("expected WHOIS registrant change on apex")
	}

	subOld := domains.NewHost("api.example.com", "example.com", "", constants.SourceSubdomainWordlist)
	subNew := subOld
	subNew.Whois.Parsed.Registrant.RegistrantName = "changed-registrant"

	subTdcs := MakeChange(subOld, subNew)
	for _, c := range subTdcs {
		if IsWhoisField(c.Field) {
			t.Errorf("subdomain should not emit WHOIS change, got field %q", c.Field)
		}
	}
}

func TestMakeChange_WHOISNameServers(t *testing.T) {
	apexOld := domains.NewHost("example.com", "example.com", "", constants.SourceApex)
	apexNew := apexOld
	apexOld.Whois.Parsed.Registrar.NameServers = "ns1.example.net, ns2.example.net"
	apexNew.Whois.Parsed.Registrar.NameServers = "ns2.example.net, ns3.example.net"

	tdcs := MakeChange(apexOld, apexNew)
	found := false
	for _, c := range tdcs {
		if c.Field == WHOIS_NAME_SERVERS {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected WHOIS nameserver change, got %+v", tdcs)
	}
}

func TestMakeChange_WHOISNameServers_orderIgnored(t *testing.T) {
	apexOld := domains.NewHost("example.com", "example.com", "", constants.SourceApex)
	apexNew := apexOld
	apexOld.Whois.Parsed.Registrar.NameServers = "ns1.example.net, ns2.example.net"
	apexNew.Whois.Parsed.Registrar.NameServers = "ns2.example.net, ns1.example.net"

	tdcs := MakeChange(apexOld, apexNew)
	for _, c := range tdcs {
		if c.Field == WHOIS_NAME_SERVERS {
			t.Fatalf("order-only NS change should not alert, got %+v", c)
		}
	}
}

func TestMakeChange_StatusAndDNS(t *testing.T) {
	old := domains.NewHost("example.com", "example.com", "", constants.SourceApex)
	old.Status = constants.AVAILABLE
	old.Dns = dns.Dns{}

	new := old
	new.Status = constants.ACTIVE
	new.Dns = dns.Dns{A: "93.184.216.34"}

	tdcs := MakeChange(old, new)
	hasStatus, hasDNS := false, false
	for _, c := range tdcs {
		if c.Field == STATUS {
			hasStatus = true
		}
		if c.Field == DNS_A {
			hasDNS = true
		}
	}
	if !hasStatus {
		t.Error("expected status change")
	}
	if !hasDNS {
		t.Error("expected DNS A change")
	}
}

func TestToTables_SortsDNSIntoStatus(t *testing.T) {
	chs := ChangeList{
		{TypoDomain: "api.example.com", Field: DNS_MX, Before: "old", After: "new"},
		{TypoDomain: "example.com", Field: NAME_REGISTRANT, Before: "a", After: "b"},
	}
	headersStatus, datasStatus, _, datasWhois := chs.ToTables()
	if len(headersStatus) == 0 {
		t.Fatal("expected status headers")
	}
	if len(datasStatus) != 1 {
		t.Fatalf("expected 1 status row (DNS), got %d", len(datasStatus))
	}
	if datasStatus[0][1] != DNS_MX {
		t.Errorf("expected DNS field in status table, got %v", datasStatus[0])
	}
	if len(datasWhois) != 1 {
		t.Fatalf("expected 1 whois row, got %d", len(datasWhois))
	}
}
