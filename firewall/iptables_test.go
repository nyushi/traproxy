package firewall

import (
	"strings"
	"testing"
)

func TestIptablesAdd(t *testing.T) {
	var expected string
	var got string

	cmds := IptablesAdd()
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

	cmds := IptablesDel()
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

	cmds := IptablesDockerAdd()
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

	cmds := IptablesDockerDel()
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

/*
func TestIptablesGetCommands(t *testing.T) {
	var cmds [][]string
	expected_cmds := [][]string{[]string{}, []string{}}
	cmds = IptablesGetCommands("-A", "OUTPUT", nil)
	if len(cmds) != 2 {
		t.Error("command len is not 2")
	}
	expected_cmds[0] = []string{"iptables", "-t", "nat", "-A", "OUTPUT", "-j", "REDIRECT", "-p", "tcp", "--dport", "80", "--to-ports", "10080"}
	if strings.Join(cmds[0], "") != strings.Join(expected_cmds[0], "") {
		t.Errorf("cmd is not match: expected=%s, got=%s", expected_cmds[0], cmds[0])
	}
	expected_cmds[1] = []string{"iptables", "-t", "nat", "-A", "OUTPUT", "-j", "REDIRECT", "-p", "tcp", "--dport", "443", "--to-ports", "10080"}
	if strings.Join(cmds[1], "") != strings.Join(expected_cmds[1], "") {
		t.Errorf("cmd is not match: expected=%s, got=%s", expected_cmds[1], cmds[1])
	}

	ifName := "docker0"
	cmds = IptablesGetCommands("-A", "DOCKER", &ifName)
	if len(cmds) != 2 {
		t.Error("command len is not 2")
	}
	expected_cmds[0] = []string{"iptables", "-t", "nat", "-A", "DOCKER", "-j", "REDIRECT", "-p", "tcp", "-i", "docker0", "--dport", "80", "--to-ports", "10080"}
	if strings.Join(cmds[0], "") != strings.Join(expected_cmds[0], "") {
		t.Errorf("cmd is not match: expected=%s, got=%s", expected_cmds[0], cmds[0])
	}
	expected_cmds[1] = []string{"iptables", "-t", "nat", "-A", "DOCKER", "-j", "REDIRECT", "-p", "tcp", "-i", "docker0", "--dport", "443", "--to-ports", "10080"}
	if strings.Join(cmds[1], "") != strings.Join(expected_cmds[1], "") {
		t.Errorf("cmd is not match: expected=%s, got=%s", expected_cmds[1], cmds[1])
	}

}
*/
