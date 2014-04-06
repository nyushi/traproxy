package firewall

import (
	"strings"
	"testing"
)

func TestIptablesAdd(t *testing.T) {
	var expected string
	var got string

	cmds := IptablesAdd([]string{})
	if len(cmds) != 2 {
		t.Error("command size if not 2")
	}
	expected = "iptables -t nat -A OUTPUT -j REDIRECT -p tcp --dport 80 --to-ports 10080"
	got = strings.Join(cmds[0], " ")
	if got != expected {
		t.Errorf("expected=%s, got=%s", expected, got)
	}
	expected = "iptables -t nat -A OUTPUT -j REDIRECT -p tcp --dport 443 --to-ports 10080"
	got = strings.Join(cmds[1], " ")
	if got != expected {
		t.Errorf("expected=%s, got=%s", expected, got)
	}
}

func TestIptablesDel(t *testing.T) {
	var expected string
	var got string

	cmds := IptablesDel([]string{})
	if len(cmds) != 2 {
		t.Error("command size if not 2")
	}
	expected = "iptables -t nat -D OUTPUT -j REDIRECT -p tcp --dport 80 --to-ports 10080"
	got = strings.Join(cmds[0], " ")
	if got != expected {
		t.Errorf("expected=%s, got=%s", expected, got)
	}
	expected = "iptables -t nat -D OUTPUT -j REDIRECT -p tcp --dport 443 --to-ports 10080"
	got = strings.Join(cmds[1], " ")
	if got != expected {
		t.Errorf("expected=%s, got=%s", expected, got)
	}
}

func TestIptablesDockerAdd(t *testing.T) {
	var expected string
	var got string

	cmds := IptablesDockerAdd([]string{})
	if len(cmds) != 2 {
		t.Error("command size if not 2")
	}
	expected = "iptables -t nat -A PREROUTING -j REDIRECT -p tcp -i docker0 --dport 80 --to-ports 10080"
	got = strings.Join(cmds[0], " ")
	if got != expected {
		t.Errorf("expected=%s, got=%s", expected, got)
	}
	expected = "iptables -t nat -A PREROUTING -j REDIRECT -p tcp -i docker0 --dport 443 --to-ports 10080"
	got = strings.Join(cmds[1], " ")
	if got != expected {
		t.Errorf("expected=%s, got=%s", expected, got)
	}
}

func TestIptablesDockerDel(t *testing.T) {
	var expected string
	var got string

	cmds := IptablesDockerDel([]string{})
	if len(cmds) != 2 {
		t.Error("command size if not 2")
	}
	expected = "iptables -t nat -D PREROUTING -j REDIRECT -p tcp -i docker0 --dport 80 --to-ports 10080"
	got = strings.Join(cmds[0], " ")
	if got != expected {
		t.Errorf("expected=%s, got=%s", expected, got)
	}
	expected = "iptables -t nat -D PREROUTING -j REDIRECT -p tcp -i docker0 --dport 443 --to-ports 10080"
	got = strings.Join(cmds[1], " ")
	if got != expected {
		t.Errorf("expected=%s, got=%s", expected, got)
	}
}

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
