package firewall

import (
	"errors"
	"fmt"
	"log"
	"net"
)

// FWType represents type of firewall
type FWType int

const (
	//FWIptables represents iptables firewall
	FWIPTables FWType = 1 << iota
	//FWPF represents pf firewall
	FWPF
)

// Firewall represents firewall operation
type Firewall interface {
	Setup() error
	Teardown() error
}

// Config represents configutaion of firewall
type Config struct {
	FWType          FWType
	ProxyAddr       *string
	WithNat         bool
	ExcludeReserved bool
	Excludes        []string
}

// ProxyHost return proxy host
func (c *Config) ProxyHost() (*string, error) {
	if c.ProxyAddr == nil {
		return nil, nil
	}
	host, _, err := net.SplitHostPort(*c.ProxyAddr)
	if err != nil {
		return nil, err
	}
	return &host, nil
}

// New creates firewall by config
func New(c *Config) Firewall {
	switch c.FWType {
	case FWIPTables:
		return &iptablesFirewall{c}
	default:
		return &nopFirewall{}
	}
}

type nopFirewall struct{}

func (n *nopFirewall) Setup() error {
	return nil
}

func (n *nopFirewall) Teardown() error {
	return nil
}

type iptablesFirewall struct {
	c *Config
}

func (i *iptablesFirewall) Setup() error {
	return i.do(true)
}

func (i *iptablesFirewall) Teardown() error {
	return i.do(false)
}

func (i *iptablesFirewall) do(add bool) error {
	rules, err := i.rules()
	if err != nil {
		return fmt.Errorf("failed to get rules: %s", err)
	}
	var failed bool
	for _, r := range rules {
		var err error
		if add {
			log.Printf("-A %s\n", r.GetCommandStr())
			err = r.Add()
		} else {
			log.Printf("-D %s\n", r.GetCommandStr())
			err = r.Del()
		}
		if err != nil {
			log.Printf("failed to execute iptables command: %s", err)
			failed = true
		}
	}
	if failed {
		return errors.New("failed to setup firewall")
	}
	return nil

}
func (i *iptablesFirewall) rules() ([]IPTablesRule, error) {
	e, err := i.excludes()
	if err != nil {
		return nil, fmt.Errorf("failed to get exclude addrs: %s", err)
	}
	rules := GetRedirectIPTablesRules(e)
	if i.c.WithNat {
		rules = append(rules, GetRedirectIPTablesNATRules(e)...)
	}
	return rules, nil
}

func (i *iptablesFirewall) excludes() ([]string, error) {
	// exclude user specified addrs
	e := make([]string, len(i.c.Excludes))
	copy(e, i.c.Excludes)

	// exclude proxy host addr
	host, err := i.c.ProxyHost()
	if err != nil {
		return nil, fmt.Errorf("failed to get proxy host: %s", err)
	}
	if host != nil {
		e = append(e, *host)
	}

	// exclude local addrs
	locals, err := LocalAddrs()
	if err != nil {
		return nil, fmt.Errorf("failed to getlocal address: %s", err)
	}
	e = append(e, GrepV4Addr(locals)...)

	// exclude reserved addrs
	e = append(e, ReservedV4Addrs()...)

	return e, nil
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
