package firewall

import (
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
func GetRedirectIPTablesRules(excludes []string) []IPTablesRule {
	rules := []IPTablesRule{}
	for _, addr := range excludes {
		rules = append(rules, []string{outputChain, "-t", "nat", "-p", "tcp", "-j", accept, "-d", addr})
	}

	rules = append(rules, []string{outputChain, "-t", "nat", "-p", "tcp", "-j", redirect, "--dport", "80", "--to-ports", "10080"})
	rules = append(rules, []string{outputChain, "-t", "nat", "-p", "tcp", "-j", redirect, "--dport", "443", "--to-ports", "10080"})
	return rules
}

// GetRedirectNATRules returns iptables rules for nat
func GetRedirectIPTablesNATRules(excludes []string) []IPTablesRule {
	rules := []IPTablesRule{}
	for _, addr := range excludes {
		rules = append(rules, []string{preroutingChain, "-t", "nat", "-p", "tcp", "-j", accept, "-d", addr})
	}

	rules = append(rules, []string{preroutingChain, "-t", "nat", "-p", "tcp", "-j", redirect, "--dport", "80", "--to-ports", "10080"})
	rules = append(rules, []string{preroutingChain, "-t", "nat", "-p", "tcp", "-j", redirect, "--dport", "443", "--to-ports", "10080"})
	return rules
}
