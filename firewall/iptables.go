package firewall

import (
	"log"
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

type IPTablesCommand []string

func (i *IPTablesCommand) Exec() ([]byte, error) {
	log.Println(strings.Join(*i, " "))
	path, err := exec.LookPath((*i)[0])
	if err != nil {
		return []byte{}, err
	}

	output, err := exec.Command(path, (*i)[1:]...).CombinedOutput()
	if err != nil {
		return output, err
	}
	return output, nil
}

func iptablesGetCommands(mode, chain, jump string, withRule bool, ifName, dst *string) (ret [][]string) {
	ret = [][]string{}
	prefix := []string{
		"iptables",
		"-t", "nat",
		mode,
		chain,
		"-j", jump,
		"-p", "tcp",
	}
	if ifName != nil {
		prefix = append(prefix, "-i", *ifName)
	}
	if dst != nil {
		prefix = append(prefix, "-d", *dst)
	}

	if !withRule {
		p := make([]string, len(prefix))
		copy(p, prefix)
		ret = append(ret, p)
		return ret
	}
	for _, v := range rule {
		p := make([]string, len(prefix))
		copy(p, prefix)
		ret = append(ret, append(p, v...))
	}
	return ret
}

func IptablesAdd(excludes []string) []IPTablesCommand {
	cmds := []IPTablesCommand{}
	for _, exc := range excludes {
		for _, v := range iptablesGetCommands(ADD, OUTPUT_CHAIN, ACCEPT, false, nil, &exc) {
			cmds = append(cmds, v)
		}
	}
	for _, v := range iptablesGetCommands(ADD, OUTPUT_CHAIN, REDIRECT, true, nil, nil) {
		cmds = append(cmds, v)
	}
	return cmds
}

func IptablesDel(excludes []string) []IPTablesCommand {
	cmds := []IPTablesCommand{}
	for _, exc := range excludes {
		for _, v := range iptablesGetCommands(DEL, OUTPUT_CHAIN, ACCEPT, false, nil, &exc) {
			cmds = append(cmds, v)
		}
	}

	for _, v := range iptablesGetCommands(DEL, OUTPUT_CHAIN, REDIRECT, true, nil, nil) {
		cmds = append(cmds, v)
	}
	return cmds

}

func IptablesDockerAdd(excludes []string) []IPTablesCommand {
	cmds := []IPTablesCommand{}
	for _, exc := range excludes {
		for _, v := range iptablesGetCommands(ADD, PREROUTING_CHAIN, ACCEPT, false, &DOCKER_IFNAME, &exc) {
			cmds = append(cmds, v)
		}
	}

	for _, v := range iptablesGetCommands(ADD, PREROUTING_CHAIN, REDIRECT, true, &DOCKER_IFNAME, nil) {
		cmds = append(cmds, v)
	}
	return cmds

}
func IptablesDockerDel(excludes []string) []IPTablesCommand {
	cmds := []IPTablesCommand{}
	for _, exc := range excludes {
		for _, v := range iptablesGetCommands(DEL, PREROUTING_CHAIN, ACCEPT, false, &DOCKER_IFNAME, &exc) {
			cmds = append(cmds, v)
		}
	}

	for _, v := range iptablesGetCommands(DEL, PREROUTING_CHAIN, REDIRECT, true, &DOCKER_IFNAME, nil) {
		cmds = append(cmds, v)
	}
	return cmds
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
