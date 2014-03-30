package firewall

import (
	"log"
	"os/exec"
	"strings"
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

func DoCommand(args ...string) {
	path, err := exec.LookPath(args[0])
	if err != nil {
		log.Println(err)
		return
	}

	output, err := exec.Command(path, args[1:]...).CombinedOutput()
	if err != nil {
		log.Println(strings.Join(args, " "))
		log.Print(string(output))
	}
}

func IptablesGetCommands(mode, chain string, ifName *string) (ret [][]string) {
	ret = [][]string{}
	prefix := []string{
		"iptables",
		"-t", "nat",
		mode,
		chain,
		//"-i", "docker0",
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

func IptablesAdd() {
	for _, v := range IptablesGetCommands(ADD, OUTPUT_CHAIN, nil) {
		DoCommand(v...)
	}
}
func IptablesDel() {
	for _, v := range IptablesGetCommands(DEL, OUTPUT_CHAIN, nil) {
		DoCommand(v...)
	}
}

func IptablesDockerAdd() {
	for _, v := range IptablesGetCommands(ADD, PREROUTING_CHAIN, &DOCKER_IFNAME) {
		DoCommand(v...)
	}
}
func IptablesDockerDel() {
	for _, v := range IptablesGetCommands(DEL, PREROUTING_CHAIN, &DOCKER_IFNAME) {
		DoCommand(v...)
	}
}
