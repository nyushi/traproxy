package firewall

import (
	"testing"
)

func TestLocalNetworks(t *testing.T) {
	_, err := LocalAddrs()
	if err != nil {
		t.Error(err)
	}
}

func TestGrepV4Addr(t *testing.T) {
	addrs := []string{"127.0.0.1/16", "", "fe80::1/64", "192.168.0.1/24"}
	v4addrs := GrepV4Addr(addrs)
	if len(v4addrs) != 2 {
		t.Error("invalid number of v4addrs")
	}
	if v4addrs[0] != "127.0.0.1/16" {
		t.Error("first element is not 127.0.0.1")
	}
	if v4addrs[1] != "192.168.0.1/24" {
		t.Error("second element is not 192.168.0.1")
	}

}

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
	rules := GetRedirectRules([]string{"127.0.0.1/8"})
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

func TestGetRedirectDockerRules(t *testing.T) {
	rules := GetRedirectDockerRules([]string{"127.0.0.1/8"})
	got := ""
	expected := "iptables PREROUTING -t nat -p tcp -j ACCEPT -d 127.0.0.1/8 -i docker0\n"
	expected += "iptables PREROUTING -t nat -p tcp -j REDIRECT --dport 80 --to-ports 10080 -i docker0\n"
	expected += "iptables PREROUTING -t nat -p tcp -j REDIRECT --dport 443 --to-ports 10080 -i docker0\n"
	for _, r := range rules {
		got += r.GetCommandStr() + "\n"
	}
	if got != expected {
		t.Error(got, expected)
	}
}
