package firewall

import (
	"log"
	"os/exec"
	"strings"
)

var (
	ADD  = "-A"
	DEL  = "-D"
	rule = map[string][]string{
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

func IptablesGetCommands(mode string) (ret [][]string) {
	prefix := []string{
		"iptables",
		"-t", "nat",
		mode,
		//"PREROUTING",
		"OUTPUT",
		//"-i", "docker0",
		"-j", "REDIRECT",
		"-p", "tcp",
	}

	for _, v := range rule {
		ret = append(ret, append(prefix, v...))
	}
	return ret
}

func IptablesAdd() {
	for _, v := range IptablesGetCommands(ADD) {
		DoCommand(v...)
	}
}
func IptablesDel() {
	for _, v := range IptablesGetCommands(DEL) {
		DoCommand(v...)
	}
}
