package firewall

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

var (
	pfctl = "pfctl"
)

func SetPFRule(excludeAddrs []string) error {
	path, err := exec.LookPath(pfctl)
	if err != nil {
		return fmt.Errorf("%s not found: %s", pfctl, err)
	}
	cmd := exec.Command(path, "-ef", "-")
	rules := []string{}
	rules = append(rules, "rdr pass inet proto tcp from any to any port = 80 -> 127.0.0.1 port 10080")
	rules = append(rules, "rdr pass inet proto tcp from any to any port = 443 -> 127.0.0.1 port 10080")
	for _, e := range excludeAddrs {
		rules = append(rules, fmt.Sprintf("pass out quick proto tcp from any to %s", e))
	}
	rules = append(rules, "pass out route-to lo0 inet proto tcp from any to any port 80 keep state")
	rules = append(rules, "pass out route-to lo0 inet proto tcp from any to any port 443 keep state")
	rulestr := strings.Join(rules, "\n") + "\n"
	log.Printf("set pf rules:\n%s", rulestr)
	cmd.Stdin = bytes.NewBuffer([]byte(rulestr))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute %s:\noutput=%s", pfctl, out)
	}
	return nil
}

func ResetPFRule() error {
	path, err := exec.LookPath(pfctl)
	if err != nil {
		return fmt.Errorf("%s not found: %s", pfctl, err)
	}
	cmd := exec.Command(path, "-df", "/etc/pf.conf")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute %s(reset): output=%s", pfctl, out)
	}
	return nil
}
