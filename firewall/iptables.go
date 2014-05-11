package firewall

import (
	"net"
	"os/exec"
	"strings"
)

var (
	redirect        = "REDIRECT"
	accept          = "ACCEPT"
	outputChain     = "OUTPUT"
	preroutingChain = "PREROUTING"
	dockerIFName    = "docker0"
	rule            = map[string][]string{
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

// IPTablesRule represents iptables rule line
type IPTablesRule []string

func (r *IPTablesRule) exec() error {
	path, err := exec.LookPath("iptables")
	if err != nil {
		return err
	}
	_, err = exec.Command(path, *r...).CombinedOutput()
	return err
}

// Add adds iptables rule
func (r *IPTablesRule) Add() error {
	*r = append([]string{"-A"}, *r...)
	return r.exec()
}

// Del deletes iptables rule
func (r *IPTablesRule) Del() error {
	*r = append([]string{"-D"}, *r...)
	return r.exec()
}

// GetCommandStr returns commandline string
func (r *IPTablesRule) GetCommandStr() string {
	return "iptables " + strings.Join(*r, " ")
}

// GetRedirectRules returns iptables rules for redirect
func GetRedirectRules(excludes []string) []IPTablesRule {
	rules := []IPTablesRule{}
	for _, addr := range excludes {
		rules = append(rules, []string{outputChain, "-t", "nat", "-p", "tcp", "-j", accept, "-d", addr})
	}

	rules = append(rules, []string{outputChain, "-t", "nat", "-p", "tcp", "-j", redirect, "--dport", "80", "--to-ports", "10080"})
	rules = append(rules, []string{outputChain, "-t", "nat", "-p", "tcp", "-j", redirect, "--dport", "443", "--to-ports", "10080"})
	return rules
}

// GetRedirectDockerRules returns iptables rules for docker containers
func GetRedirectDockerRules(excludes []string) []IPTablesRule {
	rules := []IPTablesRule{}
	for _, addr := range excludes {
		rules = append(rules, []string{preroutingChain, "-t", "nat", "-p", "tcp", "-j", accept, "-d", addr, "-i", dockerIFName})
	}

	rules = append(rules, []string{preroutingChain, "-t", "nat", "-p", "tcp", "-j", redirect, "--dport", "80", "--to-ports", "10080", "-i", dockerIFName})
	rules = append(rules, []string{preroutingChain, "-t", "nat", "-p", "tcp", "-j", redirect, "--dport", "443", "--to-ports", "10080", "-i", dockerIFName})
	return rules
}

// LocalAddrs returns assigned local address
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

// GrepV4Addr returns only ip v4 address
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
