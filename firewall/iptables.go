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

// GetRedirectNATRules returns iptables rules for nat
func GetRedirectNATRules(excludes []string) []IPTablesRule {
	rules := []IPTablesRule{}
	for _, addr := range excludes {
		rules = append(rules, []string{preroutingChain, "-t", "nat", "-p", "tcp", "-j", accept, "-d", addr})
	}

	rules = append(rules, []string{preroutingChain, "-t", "nat", "-p", "tcp", "-j", redirect, "--dport", "80", "--to-ports", "10080"})
	rules = append(rules, []string{preroutingChain, "-t", "nat", "-p", "tcp", "-j", redirect, "--dport", "443", "--to-ports", "10080"})
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

// ReservedV4Addrs returns reserved ipv4 addresses
func ReservedV4Addrs() (addrs []string) {
	return []string{
		"0.0.0.0/8",
		"10.0.0.0/8",
		"100.64.0.0/10",
		"127.0.0.0/8",
		"169.254.0.0/16",
		"172.16.0.0/12",
		"192.0.0.0/24",
		"192.0.2.0/24",
		"192.88.99.0/24",
		"192.168.0.0/16",
		"198.18.0.0/15",
		"198.51.100.0/24",
		"203.0.113.0/24",
		"224.0.0.0/4",
		"240.0.0.0/4",
		"255.255.255.255",
	}
}
