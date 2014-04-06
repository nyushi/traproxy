package firewall

import (
	"os/exec"
)

var (
	ADD              = "-A"
	DEL              = "-D"
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

func iptablesGetCommands(mode, chain string, ifName *string) (ret [][]string) {
	ret = [][]string{}
	prefix := []string{
		"iptables",
		"-t", "nat",
		mode,
		chain,
		"-j", "REDIRECT",
		"-p", "tcp",
	}
	if ifName != nil {
		prefix = append(prefix, "-i", *ifName)
	}

	for _, v := range rule {
		p := make([]string, len(prefix))
		copy(p, prefix)
		ret = append(ret, append(p, v...))
	}
	return ret
}

func IptablesAdd() []IPTablesCommand {
	cmds := []IPTablesCommand{}
	for _, v := range iptablesGetCommands(ADD, OUTPUT_CHAIN, nil) {
		cmds = append(cmds, v)
	}
	return cmds
}
func IptablesDel() []IPTablesCommand {
	cmds := []IPTablesCommand{}
	for _, v := range iptablesGetCommands(DEL, OUTPUT_CHAIN, nil) {
		cmds = append(cmds, v)
	}
	return cmds

}

func IptablesDockerAdd() []IPTablesCommand {
	cmds := []IPTablesCommand{}
	for _, v := range iptablesGetCommands(ADD, PREROUTING_CHAIN, &DOCKER_IFNAME) {
		cmds = append(cmds, v)
	}
	return cmds

}
func IptablesDockerDel() []IPTablesCommand {
	cmds := []IPTablesCommand{}
	for _, v := range iptablesGetCommands(DEL, PREROUTING_CHAIN, &DOCKER_IFNAME) {
		cmds = append(cmds, v)
	}
	return cmds
}
