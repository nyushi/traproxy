package firewall

import (
	"net"
	"os/exec"
	"strings"
)

var (
	ADD              = "-A"
	DEL              = "-D"
	REDIRECT         = "REDIRECT"
	ACCEPT           = "ACCEPT"
	OUTPUT_CHAIN     = "OUTPUT"
	PREROUTING_CHAIN = "PREROUTING"
	DOCKER_IFNAME    = "docker0"
	rule             = map[string][]string{
		"http": []string{
			"--dport", "80",
			"--to-ports", "10080",
		},
		"https": []string{
			"--dport", "443",
			"--to-ports", "10080",
		},
	}
)

type IPTablesRule []string

func (r *IPTablesRule) exec() ([]byte, error) {
	path, err := exec.LookPath("iptables")
	if err != nil {
		return nil, err
	}
	output, err := exec.Command(path, *r...).CombinedOutput()
	return output, err
}
func (r *IPTablesRule) Add() {
	*r = append([]string{"-A"}, *r...)
	r.exec()
}
func (r *IPTablesRule) Del() {
	*r = append([]string{"-D"}, *r...)
	r.exec()
}
func (r *IPTablesRule) GetCommandStr() string {
	return "iptables " + strings.Join(*r, " ")
}

func GetRedirectRules(excludes []string) []IPTablesRule {
	rules := []IPTablesRule{}
	for _, addr := range excludes {
		rules = append(rules, []string{OUTPUT_CHAIN, "-t", "nat", "-p", "tcp", "-j", ACCEPT, "-d", addr})
	}

	rules = append(rules, []string{OUTPUT_CHAIN, "-t", "nat", "-p", "tcp", "-j", REDIRECT, "--dport", "80", "--to-ports", "10080"})
	rules = append(rules, []string{OUTPUT_CHAIN, "-t", "nat", "-p", "tcp", "-j", REDIRECT, "--dport", "443", "--to-ports", "10080"})
	return rules
}

func GetRedirectDockerRules(excludes []string) []IPTablesRule {
	rules := []IPTablesRule{}
	for _, addr := range excludes {
		rules = append(rules, []string{PREROUTING_CHAIN, "-t", "nat", "-p", "tcp", "-j", ACCEPT, "-d", addr, "-i", DOCKER_IFNAME})
	}

	rules = append(rules, []string{PREROUTING_CHAIN, "-t", "nat", "-p", "tcp", "-j", REDIRECT, "--dport", "80", "--to-ports", "10080", "-i", DOCKER_IFNAME})
	rules = append(rules, []string{PREROUTING_CHAIN, "-t", "nat", "-p", "tcp", "-j", REDIRECT, "--dport", "443", "--to-ports", "10080", "-i", DOCKER_IFNAME})
	return rules
}

func LocalAddrs() ([]string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return []string{}, err
	}
	addrstrs := []string{}
	for _, v := range addrs {
		addrstrs = append(addrstrs, v.String())
	}
	return addrstrs, nil
}

func GrepV4Addr(addrs []string) []string {
	v4addrs := []string{}
	for _, v := range addrs {
		ip, _, err := net.ParseCIDR(v)
		if err != nil {
			continue
		}
		if ip.To4() == nil {
			continue
		}
		v4addrs = append(v4addrs, v)
	}
	return v4addrs
}
