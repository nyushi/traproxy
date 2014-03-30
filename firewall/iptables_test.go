package firewall

import (
	"strings"
	"testing"
)

func TestIptablesGetCommands(t *testing.T) {
	cmds := IptablesGetCommands("-A")
	if len(cmds) != 2 {
		t.Error("command len is not 2")
	}

	expected_cmd0 := []string{"iptables", "-t", "nat", "-A", "OUTPUT", "-j", "REDIRECT", "-p", "tcp", "--dport", "80", "--to-ports", "10080"}
	if strings.Join(cmds[0], "") != strings.Join(expected_cmd0, "") {
		t.Errorf("cmd is not match: expected=%s, got=%s", expected_cmd0, cmds[0])
	}

	expected_cmd1 := []string{"iptables", "-t", "nat", "-A", "OUTPUT", "-j", "REDIRECT", "-p", "tcp", "--dport", "443", "--to-ports", "10080"}
	if strings.Join(cmds[1], "") != strings.Join(expected_cmd1, "") {
		t.Errorf("cmd is not match: expected=%s, got=%s", expected_cmd1, cmds[1])
	}
}
