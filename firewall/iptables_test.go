package firewall

import (
	"testing"
)

func TestIPTablesRule(t *testing.T) {
	r := IPTablesRule{"OUTPUT"}
	if r.GetCommandStr() != "iptables OUTPUT" {
		t.Error("not match")
	}
	r = append(r, []string{"opt", "val"}...)
	if r.GetCommandStr() != "iptables OUTPUT opt val" {
		t.Error("not match")
	}
}

func TestGetRedirectRules(t *testing.T) {
	rules := GetRedirectIPTablesRules([]string{"127.0.0.1/8"})
	got := ""
	expected := "iptables OUTPUT -t nat -p tcp -j ACCEPT -d 127.0.0.1/8\n"
	expected += "iptables OUTPUT -t nat -p tcp -j REDIRECT --dport 80 --to-ports 10080\n"
	expected += "iptables OUTPUT -t nat -p tcp -j REDIRECT --dport 443 --to-ports 10080\n"
	for _, r := range rules {
		got += r.GetCommandStr() + "\n"
	}
	if got != expected {
		t.Error(got, expected)
	}
}

func TestGetRedirectNATRules(t *testing.T) {
	rules := GetRedirectIPTablesNATRules([]string{"127.0.0.1/8"})
	got := ""
	expected := "iptables PREROUTING -t nat -p tcp -j ACCEPT -d 127.0.0.1/8\n"
	expected += "iptables PREROUTING -t nat -p tcp -j REDIRECT --dport 80 --to-ports 10080\n"
	expected += "iptables PREROUTING -t nat -p tcp -j REDIRECT --dport 443 --to-ports 10080\n"
	for _, r := range rules {
		got += r.GetCommandStr() + "\n"
	}
	if got != expected {
		t.Error(got, expected)
	}
}
